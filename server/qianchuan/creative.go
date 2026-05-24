package qianchuan

import "encoding/json"

const (
	CreativeListURL   = "https://ad.oceanengine.com/open_api/v1.0/qianchuan/creative/get/"
	CreativeCreateURL = "https://ad.oceanengine.com/open_api/v1.0/qianchuan/creative/create/"
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
