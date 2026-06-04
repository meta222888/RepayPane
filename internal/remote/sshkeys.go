package remote

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/crypto/ssh"
)

var ErrPassphraseRequired = errors.New("ssh key passphrase required")

var preferredKeyNames = []string{"id_ed25519", "id_rsa", "id_ecdsa", "id_dsa"}

func sshDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".ssh"), nil
}

func parsePrivateKey(data, passphrase []byte) (ssh.Signer, error) {
	if len(passphrase) > 0 {
		return ssh.ParsePrivateKeyWithPassphrase(data, passphrase)
	}
	signer, err := ssh.ParsePrivateKey(data)
	if err == nil {
		return signer, nil
	}
	if strings.Contains(err.Error(), "cannot decode encrypted") {
		return nil, ErrPassphraseRequired
	}
	return nil, err
}

func loadPrivateKeyFile(path string, passphrase []byte) (ssh.Signer, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read private key: %w", err)
	}
	signer, err := parsePrivateKey(data, passphrase)
	if err != nil {
		return nil, err
	}
	return signer, nil
}

func loadSSHDirKeys(passphrase []byte) ([]ssh.Signer, error) {
	dir, err := sshDir()
	if err != nil {
		return nil, err
	}
	var signers []ssh.Signer
	needPass := false
	seen := map[string]bool{}

	for _, name := range preferredKeyNames {
		path := filepath.Join(dir, name)
		signer, err := loadPrivateKeyFile(path, passphrase)
		if errors.Is(err, ErrPassphraseRequired) {
			needPass = true
			continue
		}
		if err != nil {
			continue
		}
		signers = append(signers, signer)
		seen[name] = true
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		if len(signers) == 0 && needPass {
			return nil, ErrPassphraseRequired
		}
		return signers, nil
	}
	for _, e := range entries {
		if e.IsDir() || strings.HasSuffix(e.Name(), ".pub") || seen[e.Name()] {
			continue
		}
		path := filepath.Join(dir, e.Name())
		signer, err := loadPrivateKeyFile(path, passphrase)
		if errors.Is(err, ErrPassphraseRequired) {
			needPass = true
			continue
		}
		if err != nil {
			continue
		}
		signers = append(signers, signer)
	}
	if len(signers) == 0 && needPass {
		return nil, ErrPassphraseRequired
	}
	if len(signers) == 0 {
		return nil, fmt.Errorf("no usable private keys in %s", dir)
	}
	return signers, nil
}

func ListSSHKeyFiles() ([]string, error) {
	dir, err := sshDir()
	if err != nil {
		return nil, err
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	var paths []string
	add := func(name string) {
		path := filepath.Join(dir, name)
		if st, err := os.Stat(path); err == nil && !st.IsDir() && !strings.HasSuffix(name, ".pub") {
			paths = append(paths, path)
		}
	}
	for _, name := range preferredKeyNames {
		add(name)
	}
	for _, e := range entries {
		if e.IsDir() || strings.HasSuffix(e.Name(), ".pub") {
			continue
		}
		path := filepath.Join(dir, e.Name())
		found := false
		for _, p := range paths {
			if p == path {
				found = true
				break
			}
		}
		if !found {
			add(e.Name())
		}
	}
	return paths, nil
}

func NeedsPassphrase(autoSSHKey bool, privateKeyPath string) bool {
	var paths []string
	if privateKeyPath != "" {
		paths = []string{privateKeyPath}
	} else if autoSSHKey {
		dir, err := sshDir()
		if err != nil {
			return false
		}
		for _, name := range preferredKeyNames {
			paths = append(paths, filepath.Join(dir, name))
		}
	}
	for _, path := range paths {
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		if _, err := parsePrivateKey(data, nil); errors.Is(err, ErrPassphraseRequired) {
			return true
		}
	}
	return false
}
