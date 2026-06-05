package remote

import (
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

type FileInfo struct {
	Name    string
	Path    string
	Size    int64
	IsDir   bool
	ModTime time.Time
	Mode    os.FileMode
}

type Client struct {
	ssh  *ssh.Client
	sftp *sftp.Client

	heartbeatMu   sync.Mutex
	heartbeatStop chan struct{}
}

type ConnectOptions struct {
	Host          string
	Port          int
	Username      string
	Password      string
	AutoSSHKey    bool
	PrivateKey    string
	KeyPassphrase []byte
}

func Connect(opts ConnectOptions) (*Client, error) {
	if opts.Port == 0 {
		opts.Port = 22
	}
	authMethods := []ssh.AuthMethod{}
	if opts.Password != "" {
		authMethods = append(authMethods, ssh.Password(opts.Password))
	}
	if opts.AutoSSHKey {
		signers, err := loadSSHDirKeys(opts.KeyPassphrase)
		if err != nil {
			return nil, err
		}
		authMethods = append(authMethods, ssh.PublicKeys(signers...))
	} else if opts.PrivateKey != "" {
		signer, err := loadPrivateKeyFile(opts.PrivateKey, opts.KeyPassphrase)
		if err != nil {
			return nil, err
		}
		authMethods = append(authMethods, ssh.PublicKeys(signer))
	}
	if len(authMethods) == 0 {
		return nil, fmt.Errorf("password or private key required")
	}

	addr := fmt.Sprintf("%s:%d", opts.Host, opts.Port)
	sshClient, err := ssh.Dial("tcp", addr, &ssh.ClientConfig{
		User:            opts.Username,
		Auth:            authMethods,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), //nolint:gosec // MVP; pin keys in production
		Timeout:         15 * time.Second,
	})
	if err != nil {
		return nil, fmt.Errorf("ssh dial: %w", err)
	}

	sftpClient, err := sftp.NewClient(sshClient)
	if err != nil {
		_ = sshClient.Close()
		return nil, fmt.Errorf("sftp: %w", err)
	}

	return &Client{ssh: sshClient, sftp: sftpClient}, nil
}

func (c *Client) Close() error {
	c.StopHeartbeat()
	if c.sftp != nil {
		_ = c.sftp.Close()
	}
	if c.ssh != nil {
		return c.ssh.Close()
	}
	return nil
}

func (c *Client) ListDir(dir string) ([]FileInfo, error) {
	dir = normalizeRemote(dir)
	entries, err := c.sftp.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	out := make([]FileInfo, 0, len(entries))
	for _, e := range entries {
		name := e.Name()
		if name == "." {
			continue
		}
		full := path.Join(dir, name)
		out = append(out, FileInfo{
			Name:    name,
			Path:    full,
			Size:    e.Size(),
			IsDir:   e.IsDir(),
			ModTime: e.ModTime(),
			Mode:    e.Mode(),
		})
	}
	return out, nil
}

func (c *Client) Stat(p string) (FileInfo, error) {
	p = normalizeRemote(p)
	st, err := c.sftp.Stat(p)
	if err != nil {
		return FileInfo{}, err
	}
	return FileInfo{
		Name:    path.Base(p),
		Path:    p,
		Size:    st.Size(),
		IsDir:   st.IsDir(),
		ModTime: st.ModTime(),
		Mode:    st.Mode(),
	}, nil
}

func (c *Client) ReadFile(p string) ([]byte, error) {
	p = normalizeRemote(p)
	f, err := c.sftp.Open(p)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return io.ReadAll(f)
}

func (c *Client) WriteFile(p string, data []byte) error {
	p = normalizeRemote(p)
	dir := path.Dir(p)
	if dir != "" && dir != "/" {
		_ = c.sftp.MkdirAll(dir)
	}
	f, err := c.sftp.Create(p)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.Write(data)
	return err
}

type ProgressFunc func(transferred int64)

type progressReader struct {
	r    io.Reader
	done int64
	fn   ProgressFunc
}

func (p *progressReader) Read(b []byte) (int, error) {
	n, err := p.r.Read(b)
	if n > 0 {
		p.done += int64(n)
		if p.fn != nil {
			p.fn(p.done)
		}
	}
	return n, err
}

func (c *Client) Upload(localPath, remotePath string) error {
	return c.UploadWithProgress(localPath, remotePath, nil)
}

func (c *Client) UploadWithProgress(localPath, remotePath string, fn ProgressFunc) error {
	localFile, err := os.Open(localPath)
	if err != nil {
		return err
	}
	defer localFile.Close()

	remotePath = normalizeRemote(remotePath)
	dir := path.Dir(remotePath)
	if dir != "" && dir != "/" {
		_ = c.sftp.MkdirAll(dir)
	}

	remoteFile, err := c.sftp.Create(remotePath)
	if err != nil {
		return err
	}
	defer remoteFile.Close()
	src := io.Reader(localFile)
	if fn != nil {
		src = &progressReader{r: localFile, fn: fn}
	}
	_, err = io.Copy(remoteFile, src)
	return err
}

func (c *Client) Download(remotePath, localPath string) error {
	return c.DownloadWithProgress(remotePath, localPath, nil)
}

func (c *Client) DownloadWithProgress(remotePath, localPath string, fn ProgressFunc) error {
	remotePath = normalizeRemote(remotePath)
	remoteFile, err := c.sftp.Open(remotePath)
	if err != nil {
		return err
	}
	defer remoteFile.Close()

	if err := os.MkdirAll(filepath.Dir(localPath), 0o755); err != nil {
		return err
	}
	localFile, err := os.Create(localPath)
	if err != nil {
		return err
	}
	defer localFile.Close()
	src := io.Reader(remoteFile)
	if fn != nil {
		src = &progressReader{r: remoteFile, fn: fn}
	}
	_, err = io.Copy(localFile, src)
	return err
}

func (c *Client) Mkdir(p string) error {
	return c.sftp.Mkdir(normalizeRemote(p))
}

func (c *Client) Rename(oldPath, newPath string) error {
	oldPath = normalizeRemote(oldPath)
	newPath = normalizeRemote(newPath)
	if err := c.sftp.Rename(oldPath, newPath); err != nil {
		return fmt.Errorf("rename: %w", err)
	}
	return nil
}

func (c *Client) Remove(p string) error {
	p = normalizeRemote(p)
	st, err := c.sftp.Stat(p)
	if err != nil {
		return err
	}
	if st.IsDir() {
		return c.sftp.RemoveDirectory(p)
	}
	return c.sftp.Remove(p)
}

func (c *Client) RemoveAll(p string) error {
	p = normalizeRemote(p)
	st, err := c.sftp.Stat(p)
	if err != nil {
		return err
	}
	if !st.IsDir() {
		return c.sftp.Remove(p)
	}
	entries, err := c.ListDir(p)
	if err != nil {
		return err
	}
	for _, e := range entries {
		child := e.Path
		if e.IsDir {
			if err := c.RemoveAll(child); err != nil {
				return err
			}
			continue
		}
		if err := c.sftp.Remove(child); err != nil {
			return err
		}
	}
	return c.sftp.RemoveDirectory(p)
}

func (c *Client) CopyPath(src, dst string) error {
	src = normalizeRemote(src)
	dst = normalizeRemote(dst)
	st, err := c.Stat(src)
	if err != nil {
		return err
	}
	if st.IsDir {
		return c.copyDirRemote(src, dst)
	}
	data, err := c.ReadFile(src)
	if err != nil {
		return err
	}
	return c.WriteFile(dst, data)
}

func (c *Client) copyDirRemote(src, dst string) error {
	_ = c.Mkdir(dst)
	entries, err := c.ListDir(src)
	if err != nil {
		return err
	}
	for _, e := range entries {
		dstChild := path.Join(dst, e.Name)
		if e.IsDir {
			if err := c.copyDirRemote(e.Path, dstChild); err != nil {
				return err
			}
			continue
		}
		if err := c.CopyPath(e.Path, dstChild); err != nil {
			return err
		}
	}
	return nil
}

func normalizeRemote(p string) string {
	p = strings.ReplaceAll(p, "\\", "/")
	if p == "" {
		return "/"
	}
	if !strings.HasPrefix(p, "/") {
		p = "/" + p
	}
	return path.Clean(p)
}
