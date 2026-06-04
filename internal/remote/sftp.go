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
	Host       string
	Port       int
	Username   string
	Password   string
	AutoSSHKey bool
	PrivateKey string
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
		signers, err := loadSSHDirKeys()
		if err != nil {
			return nil, err
		}
		if len(signers) > 0 {
			authMethods = append(authMethods, ssh.PublicKeys(signers...))
		}
	} else if opts.PrivateKey != "" {
		keyData, err := os.ReadFile(opts.PrivateKey)
		if err != nil {
			return nil, fmt.Errorf("read private key: %w", err)
		}
		signer, err := ssh.ParsePrivateKey(keyData)
		if err != nil {
			return nil, fmt.Errorf("parse private key: %w", err)
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

func (c *Client) Upload(localPath, remotePath string) error {
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
	_, err = io.Copy(remoteFile, localFile)
	return err
}

func (c *Client) Download(remotePath, localPath string) error {
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
	_, err = io.Copy(localFile, remoteFile)
	return err
}

func (c *Client) Mkdir(p string) error {
	return c.sftp.Mkdir(normalizeRemote(p))
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

func loadSSHDirKeys() ([]ssh.Signer, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("home dir: %w", err)
	}
	sshDir := filepath.Join(home, ".ssh")
	entries, err := os.ReadDir(sshDir)
	if err != nil {
		return nil, fmt.Errorf("read ~/.ssh: %w", err)
	}
	var signers []ssh.Signer
	for _, e := range entries {
		if e.IsDir() || strings.HasSuffix(e.Name(), ".pub") {
			continue
		}
		data, err := os.ReadFile(filepath.Join(sshDir, e.Name()))
		if err != nil {
			continue
		}
		signer, err := ssh.ParsePrivateKey(data)
		if err != nil {
			continue
		}
		signers = append(signers, signer)
	}
	return signers, nil
}
