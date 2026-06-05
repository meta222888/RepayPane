package cloudsync

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/relaypane/relaypane/internal/config"
)

const apiBase = "https://api.pc530.com/v1/storage"

type CloudStatus struct {
	Exists    bool
	UpdatedAt string
}

type apiResponse struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    *apiStorageData `json:"data"`
}

type apiStorageData struct {
	Key       string `json:"key"`
	Value     string `json:"value"`
	ExpiredAt string `json:"expired_at"`
	UpdatedAt string `json:"updated_at"`
}

type Client struct {
	secret string
	http   *http.Client
}

func NewClient(apiSecret string) *Client {
	return &Client{
		secret: strings.TrimSpace(apiSecret),
		http:   &http.Client{Timeout: 30 * time.Second},
	}
}

func (c *Client) auth(req *http.Request) {
	req.Header.Set("Authorization", "Bearer "+c.secret)
	req.Header.Set("Content-Type", "application/json")
}

func (c *Client) QueryStatus() (CloudStatus, error) {
	if c.secret == "" {
		return CloudStatus{}, fmt.Errorf("api secret required")
	}
	u := apiBase + "/get?key=" + url.QueryEscape(StorageKey)
	req, err := http.NewRequest(http.MethodGet, u, nil)
	if err != nil {
		return CloudStatus{}, err
	}
	c.auth(req)
	resp, err := c.http.Do(req)
	if err != nil {
		return CloudStatus{}, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return CloudStatus{}, err
	}
	if resp.StatusCode == http.StatusNotFound {
		return CloudStatus{Exists: false}, nil
	}
	var out apiResponse
	if err := json.Unmarshal(body, &out); err != nil {
		return CloudStatus{}, fmt.Errorf("invalid response: %w", err)
	}
	if resp.StatusCode == http.StatusUnauthorized || out.Code == 401 {
		return CloudStatus{}, fmt.Errorf("unauthorized: check API secret")
	}
	if out.Code != 200 || out.Data == nil {
		return CloudStatus{}, fmt.Errorf("%s", out.Message)
	}
	return CloudStatus{Exists: true, UpdatedAt: out.Data.UpdatedAt}, nil
}

func (c *Client) GetEncryptedValue() (string, CloudStatus, error) {
	if c.secret == "" {
		return "", CloudStatus{}, fmt.Errorf("api secret required")
	}
	u := apiBase + "/get?key=" + url.QueryEscape(StorageKey)
	req, err := http.NewRequest(http.MethodGet, u, nil)
	if err != nil {
		return "", CloudStatus{}, err
	}
	c.auth(req)
	resp, err := c.http.Do(req)
	if err != nil {
		return "", CloudStatus{}, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", CloudStatus{}, err
	}
	if resp.StatusCode == http.StatusNotFound {
		return "", CloudStatus{Exists: false}, nil
	}
	var out apiResponse
	if err := json.Unmarshal(body, &out); err != nil {
		return "", CloudStatus{}, fmt.Errorf("invalid response: %w", err)
	}
	if resp.StatusCode == http.StatusUnauthorized || out.Code == 401 {
		return "", CloudStatus{}, fmt.Errorf("unauthorized: check API secret")
	}
	if out.Code != 200 || out.Data == nil {
		return "", CloudStatus{}, fmt.Errorf("%s", out.Message)
	}
	return out.Data.Value, CloudStatus{Exists: true, UpdatedAt: out.Data.UpdatedAt}, nil
}

func (c *Client) SetEncryptedValue(value string) (updatedAt string, err error) {
	if c.secret == "" {
		return "", fmt.Errorf("api secret required")
	}
	payload := map[string]any{
		"key":       StorageKey,
		"value":     value,
		"expire_in": 0,
	}
	data, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	req, err := http.NewRequest(http.MethodPost, apiBase+"/set", bytes.NewReader(data))
	if err != nil {
		return "", err
	}
	c.auth(req)
	resp, err := c.http.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	var out apiResponse
	if err := json.Unmarshal(body, &out); err != nil {
		return "", fmt.Errorf("invalid response: %w", err)
	}
	if resp.StatusCode == http.StatusUnauthorized || out.Code == 401 {
		return "", fmt.Errorf("unauthorized: check API secret")
	}
	if resp.StatusCode == http.StatusTooManyRequests || out.Code == 429 {
		return "", fmt.Errorf("rate limited, try again later")
	}
	if out.Code != 200 {
		if out.Message != "" {
			return "", fmt.Errorf("%s", out.Message)
		}
		return "", fmt.Errorf("upload failed (HTTP %d)", resp.StatusCode)
	}
	if out.Data != nil {
		return out.Data.UpdatedAt, nil
	}
	return time.Now().Format("2006-01-02 15:04:05"), nil
}

func (c *Client) Delete() error {
	if c.secret == "" {
		return fmt.Errorf("api secret required")
	}
	payload := map[string]string{"key": StorageKey}
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	req, err := http.NewRequest(http.MethodDelete, apiBase+"/delete", bytes.NewReader(data))
	if err != nil {
		return err
	}
	c.auth(req)
	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if resp.StatusCode == http.StatusNotFound {
		return nil
	}
	var out apiResponse
	if err := json.Unmarshal(body, &out); err != nil {
		return fmt.Errorf("invalid response: %w", err)
	}
	if resp.StatusCode == http.StatusUnauthorized || out.Code == 401 {
		return fmt.Errorf("unauthorized: check API secret")
	}
	if out.Code != 200 && out.Code != 404 {
		if out.Message != "" {
			return fmt.Errorf("%s", out.Message)
		}
		return fmt.Errorf("delete failed (HTTP %d)", resp.StatusCode)
	}
	return nil
}

func Upload(store *config.Store, apiSecret, password string) (cloudUpdatedAt string, err error) {
	plain, err := PackStore(store)
	if err != nil {
		return "", err
	}
	enc, err := Encrypt(password, plain)
	if err != nil {
		return "", err
	}
	return NewClient(apiSecret).SetEncryptedValue(enc)
}

func Download(apiSecret, password string) ([]byte, CloudStatus, error) {
	client := NewClient(apiSecret)
	enc, status, err := client.GetEncryptedValue()
	if err != nil {
		return nil, CloudStatus{}, err
	}
	if !status.Exists || enc == "" {
		return nil, status, nil
	}
	plain, err := Decrypt(password, enc)
	if err != nil {
		return nil, status, fmt.Errorf("decrypt failed: wrong password or corrupted data")
	}
	return plain, status, nil
}
