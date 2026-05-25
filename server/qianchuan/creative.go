package qianchuan

import (
	"encoding/json"
	"fmt"
)

const (
	CreativeListURL   = APIBaseURL + "/v1.0/qianchuan/creative/get/"
	CreativeCreateURL = APIBaseURL + "/v1.0/qianchuan/creative/create/"
)

type CreativeListRequest struct {
	AdvertiserID int64 `json:"advertiser_id"`
	Page         int   `json:"page"`
	PageSize     int   `json:"page_size"`
}

type CreativeListResponse struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
}

type CreativeCreateRequest struct {
	AdvertiserID     int64                  `json:"advertiser_id"`
	AdID             int64                  `json:"ad_id"`
	CreativeMaterial map[string]interface{} `json:"creative_material"`
	TitleMaterial    map[string]interface{} `json:"title_material,omitempty"`
	ImageMode        string                 `json:"image_mode"`
	ImageInfo        map[string]interface{} `json:"image_info,omitempty"`
}

func (c *Client) ListCreatives(account *AdAccount, page, pageSize int, adID int64) (*APIResponse, error) {
	url := fmt.Sprintf("%s?advertiser_id=%d&page=%d&page_size=%d", CreativeListURL, account.AdvertiserID, page, pageSize)
	if adID > 0 {
		url += fmt.Sprintf("&ad_id=%d", adID)
	}
	return c.Do("GET", url, nil, account.ID)
}

func (c *Client) CreateCreative(account *AdAccount, req *CreativeCreateRequest) (int64, error) {
	req.AdvertiserID = account.AdvertiserID
	resp, err := c.Do("POST", CreativeCreateURL, req, account.ID)
	if err != nil {
		return 0, err
	}
	if resp.Code != 0 {
		return 0, fmt.Errorf("create creative failed: code=%d msg=%s", resp.Code, resp.Message)
	}
	var result struct {
		CreativeID int64 `json:"creative_id"`
	}
	if err := json.Unmarshal(resp.Data, &result); err != nil {
		return 0, err
	}
	return result.CreativeID, nil
}
