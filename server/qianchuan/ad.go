package qianchuan

import (
	"encoding/json"
	"fmt"
)

const (
	UniAdCreateURL = APIBaseURL + "/v1.0/qianchuan/uni_aweme/ad/create/"
	UniAdUpdateURL = APIBaseURL + "/v1.0/qianchuan/uni_aweme/ad/update/"
	UniAdStatusURL = APIBaseURL + "/v1.0/qianchuan/uni_aweme/ad/status/update/"
	UniAdListURL   = APIBaseURL + "/v1.0/qianchuan/uni_aweme/ad/list/get/"
	UniAdDetailURL = APIBaseURL + "/v1.0/qianchuan/uni_aweme/ad/detail/get/"
)

type CreateAdRequest struct {
	AdvertiserID    int64                  `json:"advertiser_id"`
	Name            string                 `json:"name,omitempty"`
	AwemeID         int64                  `json:"aweme_id,omitempty"`
	MarketingGoal   string                 `json:"marketing_goal"`
	ProductIDs      []int64                `json:"product_ids,omitempty"`
	DeliverySetting map[string]interface{} `json:"delivery_setting"`
	CreativeSetting map[string]interface{} `json:"creative_setting,omitempty"`
}

type UpdateAdRequest struct {
	AdvertiserID    int64                  `json:"advertiser_id"`
	AdID            int64                  `json:"ad_id"`
	Name            string                 `json:"name,omitempty"`
	DeliverySetting map[string]interface{} `json:"delivery_setting,omitempty"`
	CreativeSetting map[string]interface{} `json:"creative_setting,omitempty"`
}

type StatusRequest struct {
	AdvertiserID int64   `json:"advertiser_id"`
	AdIDs        []int64 `json:"ad_ids"`
	OptStatus    string  `json:"opt_status"`
}

type ListRequest struct {
	AdvertiserID int64                  `json:"advertiser_id"`
	Page         int                    `json:"page,omitempty"`
	PageSize     int                    `json:"page_size,omitempty"`
	Filtering    map[string]interface{} `json:"filtering,omitempty"`
}

type AdAccount struct {
	ID           uint
	AdvertiserID int64
}

func (c *Client) CreateUniAd(account *AdAccount, req *CreateAdRequest) (int64, error) {
	req.AdvertiserID = account.AdvertiserID
	resp, err := c.Do("POST", UniAdCreateURL, req, account.ID)
	if err != nil {
		return 0, err
	}
	if resp.Code != 0 {
		return 0, fmt.Errorf("qianchuan create ad failed: code=%d msg=%s", resp.Code, resp.Message)
	}
	var result struct{ AdID int64 `json:"ad_id"` }
	if err := json.Unmarshal(resp.Data, &result); err != nil {
		return 0, fmt.Errorf("parse ad_id failed: %w", err)
	}
	return result.AdID, nil
}

func (c *Client) UpdateUniAd(account *AdAccount, req *UpdateAdRequest) error {
	req.AdvertiserID = account.AdvertiserID
	resp, err := c.Do("POST", UniAdUpdateURL, req, account.ID)
	if err != nil {
		return err
	}
	if resp.Code != 0 {
		return fmt.Errorf("qianchuan update ad failed: code=%d msg=%s", resp.Code, resp.Message)
	}
	return nil
}

func (c *Client) UpdateUniAdStatus(account *AdAccount, adIDs []int64, status string) error {
	resp, err := c.Do("POST", UniAdStatusURL, &StatusRequest{
		AdvertiserID: account.AdvertiserID,
		AdIDs:        adIDs,
		OptStatus:    status,
	}, account.ID)
	if err != nil {
		return err
	}
	if resp.Code != 0 {
		return fmt.Errorf("qianchuan update status failed: code=%d msg=%s", resp.Code, resp.Message)
	}
	return nil
}

func (c *Client) ListUniAds(account *AdAccount, page, pageSize int) (*APIResponse, error) {
	return c.Do("GET", fmt.Sprintf("%s?advertiser_id=%d&page=%d&page_size=%d",
		UniAdListURL, account.AdvertiserID, page, pageSize), nil, account.ID)
}

func (c *Client) GetUniAdDetail(account *AdAccount, adID int64) (*APIResponse, error) {
	url := fmt.Sprintf("%s?advertiser_id=%d&ad_id=%d", UniAdDetailURL, account.AdvertiserID, adID)
	return c.Do("GET", url, nil, account.ID)
}
