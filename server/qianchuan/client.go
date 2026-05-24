package qianchuan

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/example/qianchuan-saas/db"
	"github.com/example/qianchuan-saas/models"
)

type Client struct {
	HTTP       *http.Client
	OAuth      *OAuthClient
	MaxRetries int
}

func NewClient(appID, secret string) *Client {
	return &Client{
		HTTP:       &http.Client{Timeout: 30 * time.Second},
		OAuth:      NewOAuthClient(appID, secret),
		MaxRetries: 3,
	}
}

type APIResponse struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
}

func (c *Client) Do(method, urlStr string, body interface{}, accountID uint) (*APIResponse, error) {
	token, err := c.getValidToken(accountID)
	if err != nil {
		return nil, fmt.Errorf("get token: %w", err)
	}
	return c.doWithRetry(method, urlStr, body, token, accountID, c.MaxRetries)
}

func (c *Client) doWithRetry(method, urlStr string, body interface{}, token string, accountID uint, retries int) (*APIResponse, error) {
	var bodyReader io.Reader
	if body != nil {
		b, _ := json.Marshal(body)
		bodyReader = bytes.NewReader(b)
	}

	req, _ := http.NewRequest(method, urlStr, bodyReader)
	req.Header.Set("Access-Token", token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTP.Do(req)
	if err != nil {
		if retries > 0 {
			time.Sleep(time.Duration(c.MaxRetries-retries+1) * time.Second)
			return c.doWithRetry(method, urlStr, body, token, accountID, retries-1)
		}
		return nil, fmt.Errorf("request failed after retries: %w", err)
	}
	defer resp.Body.Close()

	var apiResp APIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("decode response failed: %w", err)
	}

	if apiResp.Code == 401 || apiResp.Code == 40102 {
		return c.handleTokenExpiry(method, urlStr, body, accountID)
	}
	if apiResp.Code == 429 && retries > 0 {
		time.Sleep(time.Duration(c.MaxRetries-retries+1) * 2 * time.Second)
		return c.doWithRetry(method, urlStr, body, token, accountID, retries-1)
	}

	return &apiResp, nil
}

func (c *Client) handleTokenExpiry(method, urlStr string, body interface{}, accountID uint) (*APIResponse, error) {
	var account models.QianchuanAccount
	if err := db.DB.First(&account, accountID).Error; err != nil {
		return nil, fmt.Errorf("account not found: %w", err)
	}
	tr, err := c.OAuth.RefreshToken(account.RefreshToken)
	if err != nil {
		return nil, fmt.Errorf("refresh failed: %w", err)
	}
	db.DB.Model(&account).Updates(map[string]interface{}{
		"access_token":  tr.AccessToken,
		"refresh_token": tr.RefreshToken,
	})
	return c.doWithRetry(method, urlStr, body, tr.AccessToken, accountID, 1)
}

func (c *Client) getValidToken(accountID uint) (string, error) {
	var account models.QianchuanAccount
	if err := db.DB.First(&account, accountID).Error; err != nil {
		return "", fmt.Errorf("account not found: %w", err)
	}
	log.Printf("using token for account %d (advertiser_id=%d)", accountID, account.AdvertiserID)
	return account.AccessToken, nil
}
