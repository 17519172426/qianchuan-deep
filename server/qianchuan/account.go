package qianchuan

import "fmt"

const AccountInfoURL = APIBaseURL + "/v1.0/qianchuan/advertiser/info/"

// GetAccountInfo fetches advertiser account info (balance, status, etc.)
func (c *Client) GetAccountInfo(account *AdAccount) (*APIResponse, error) {
	url := fmt.Sprintf("%s?advertiser_id=%d", AccountInfoURL, account.AdvertiserID)
	return c.Do("GET", url, nil, account.ID)
}
