package qianchuan

import (
	"fmt"
	"net/url"
)

const ReportURL = APIBaseURL + "/v1.0/qianchuan/report/ad/get/"

type ReportRequest struct {
	AdvertiserID int64   `json:"advertiser_id"`
	StartDate    string  `json:"start_date"`
	EndDate      string  `json:"end_date"`
	Granularity  string  `json:"granularity"`
	AdIDs        []int64 `json:"ad_ids,omitempty"`
	Page         int     `json:"page,omitempty"`
	PageSize     int     `json:"page_size,omitempty"`
}

// GetReports fetches hourly/daily ad reports
func (c *Client) GetReports(account *AdAccount, startDate, endDate string, adIDs []int64) (*APIResponse, error) {
	params := url.Values{}
	params.Set("advertiser_id", fmt.Sprintf("%d", account.AdvertiserID))
	params.Set("start_date", startDate)
	params.Set("end_date", endDate)
	params.Set("granularity", "HOURLY")
	params.Set("page_size", "100")

	for _, id := range adIDs {
		params.Add("ad_ids", fmt.Sprintf("%d", id))
	}

	u := ReportURL + "?" + params.Encode()
	return c.Do("GET", u, nil, account.ID)
}
