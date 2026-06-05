package cloudsync

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"

	"golang.org/x/crypto/scrypt"
)

const (
	scryptN      = 32768
	scryptR      = 8
	scryptP      = 1
	scryptKeyLen = 32
	saltLen      = 16
)

func Encrypt(password string, plain []byte) (string, error) {
	if password == "" {
		return "", fmt.Errorf("encryption password required")
	}
	salt := make([]byte, saltLen)
	if _, err := io.ReadFull(rand.Reader, salt); err != nil {
		return "", err
	}
	key, err := scrypt.Key([]byte(password), salt, scryptN, scryptR, scryptP, scryptKeyLen)
	if err != nil {
		return "", err
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}
	out := append(append(salt, nonce...), gcm.Seal(nil, nonce, plain, nil)...)
	return base64.StdEncoding.EncodeToString(out), nil
}

func Decrypt(password, encoded string) ([]byte, error) {
	if password == "" {
		return nil, fmt.Errorf("encryption password required")
	}
	raw, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return nil, err
	}
	if len(raw) < saltLen+1 {
		return nil, fmt.Errorf("invalid encrypted payload")
	}
	salt := raw[:saltLen]
	key, err := scrypt.Key([]byte(password), salt, scryptN, scryptR, scryptP, scryptKeyLen)
	if err != nil {
		return nil, err
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	nonceSize := gcm.NonceSize()
	if len(raw) < saltLen+nonceSize+1 {
		return nil, fmt.Errorf("invalid encrypted payload")
	}
	nonce := raw[saltLen : saltLen+nonceSize]
	ciphertext := raw[saltLen+nonceSize:]
	return gcm.Open(nil, nonce, ciphertext, nil)
}
