package qianchuan

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

const (
	BaseURL    = "https://ad.oceanengine.com/open_api"
	APIBaseURL = "https://api.oceanengine.com/open_api"
	TokenURL   = "/oauth2/access_token/"
	RefreshURL = "/oauth2/refresh_token/"
)

type TokenResponse struct {
	Code         int    `json:"code"`
	Message      string `json:"message"`
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
}

type OAuthClient struct {
	AppID  string
	Secret string
	HTTP   *http.Client
}

func NewOAuthClient(appID, secret string) *OAuthClient {
	return &OAuthClient{
		AppID:  appID,
		Secret: secret,
		HTTP:   &http.Client{},
	}
}

func (c *OAuthClient) GetToken(authCode string) (*TokenResponse, error) {
	data := url.Values{}
	data.Set("app_id", c.AppID)
	data.Set("secret", c.Secret)
	data.Set("grant_type", "authorization_code")
	data.Set("auth_code", authCode)

	req, _ := http.NewRequest("POST", BaseURL+TokenURL, strings.NewReader(data.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, fmt.Errorf("get token request failed: %w", err)
	}
	defer resp.Body.Close()

	var tr TokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tr); err != nil {
		return nil, fmt.Errorf("decode token response failed: %w", err)
	}
	if tr.Code != 0 {
		return nil, fmt.Errorf("qianchuan token error: code=%d msg=%s", tr.Code, tr.Message)
	}
	return &tr, nil
}

func (c *OAuthClient) RefreshToken(refreshToken string) (*TokenResponse, error) {
	data := url.Values{}
	data.Set("app_id", c.AppID)
	data.Set("secret", c.Secret)
	data.Set("grant_type", "refresh_token")
	data.Set("refresh_token", refreshToken)

	req, _ := http.NewRequest("POST", BaseURL+RefreshURL, strings.NewReader(data.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, fmt.Errorf("refresh token request failed: %w", err)
	}
	defer resp.Body.Close()

	var tr TokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tr); err != nil {
		return nil, fmt.Errorf("decode refresh response failed: %w", err)
	}
	if tr.Code != 0 {
		return nil, fmt.Errorf("qianchuan refresh error: code=%d msg=%s", tr.Code, tr.Message)
	}
	return &tr, nil
}
