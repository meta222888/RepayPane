package walkui

import (
	"io"
	"os"
	"path/filepath"
)

func copyPathLocal(src, dst string) error {
	info, err := os.Stat(src)
	if err != nil {
		return err
	}
	if info.IsDir() {
		return copyDirLocal(src, dst, info.Mode())
	}
	return copyFileLocal(src, dst)
}

func copyFileLocal(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	return err
}

func copyDirLocal(src, dst string, mode os.FileMode) error {
	if err := os.MkdirAll(dst, mode); err != nil {
		return err
	}
	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}
	for _, e := range entries {
		srcPath := filepath.Join(src, e.Name())
		dstPath := filepath.Join(dst, e.Name())
		info, err := e.Info()
		if err != nil {
			continue
		}
		if info.IsDir() {
			if err := copyDirLocal(srcPath, dstPath, info.Mode()); err != nil {
				return err
			}
			continue
		}
		if err := copyFileLocal(srcPath, dstPath); err != nil {
			return err
		}
	}
	return nil
}

func removePathLocal(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		return err
	}
	if info.IsDir() {
		return os.RemoveAll(path)
	}
	return os.Remove(path)
}
