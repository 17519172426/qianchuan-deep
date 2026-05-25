package worker

import (
	"encoding/json"
	"log"
	"time"

	"github.com/example/qianchuan-saas/db"
	"github.com/example/qianchuan-saas/models"
	"github.com/example/qianchuan-saas/qianchuan"
	"gorm.io/gorm/clause"
)

type SyncWorker struct {
	QC       *qianchuan.Client
	Interval time.Duration
}

func NewSyncWorker(qc *qianchuan.Client) *SyncWorker {
	return &SyncWorker{QC: qc, Interval: 5 * time.Minute}
}

func (w *SyncWorker) Start() {
	log.Printf("sync worker started, interval=%s", w.Interval)
	ticker := time.NewTicker(w.Interval)
	go func() {
		for range ticker.C {
			w.syncAds()
			w.syncReports()
			w.syncCreatives()
			w.syncAccountInfo()
		}
	}()
}

func (w *SyncWorker) syncAds() {
	var accounts []models.QianchuanAccount
	db.DB.Find(&accounts)

	for _, acc := range accounts {
		accRef := qianchuan.AdAccount{ID: acc.ID, AdvertiserID: acc.AdvertiserID}
		resp, err := w.QC.ListUniAds(&accRef, 1, 100)
		if err != nil {
			log.Printf("sync ads failed for account %d: %v", acc.ID, err)
			continue
		}
		if resp.Code != 0 {
			log.Printf("sync ads error for account %d: code=%d msg=%s", acc.ID, resp.Code, resp.Message)
			continue
		}
		var result struct {
			List []struct {
				AdID    int64                  `json:"ad_id"`
				Name    string                 `json:"name"`
				Status  string                 `json:"status"`
				Metrics map[string]interface{} `json:"metrics"`
			} `json:"list"`
		}
		if err := json.Unmarshal(resp.Data, &result); err != nil {
			log.Printf("parse ads response failed: %v", err)
			continue
		}
		for _, item := range result.List {
			var ad models.UniAd
			if err := db.DB.Where("qianchuan_ad_id = ? AND account_id = ?", item.AdID, acc.ID).First(&ad).Error; err != nil {
				continue
			}
			updates := map[string]interface{}{"status": item.Status}
			if item.Metrics != nil {
				m := models.JSONMap(item.Metrics)
				updates["metrics_json"] = m
			}
			db.DB.Model(&ad).Updates(updates)
		}
		db.DB.Model(&acc).Update("last_sync_at", time.Now())
	}
}

func (w *SyncWorker) syncReports() {
	var accounts []models.QianchuanAccount
	db.DB.Where("status = ?", "active").Find(&accounts)

	yesterday := time.Now().AddDate(0, 0, -1).Format("2006-01-02")
	today := time.Now().Format("2006-01-02")

	for _, acc := range accounts {
		accRef := qianchuan.AdAccount{ID: acc.ID, AdvertiserID: acc.AdvertiserID}

		page := 1
		for {
			resp, err := w.QC.GetReports(&accRef, yesterday, today, nil)
			if err != nil || resp.Code != 0 {
				log.Printf("sync reports failed for account %d: %v", acc.ID, err)
				break
			}

			var result struct {
				List []struct {
					AdID         int64   `json:"ad_id"`
					StatDatetime string  `json:"stat_datetime"`
					Impressions  int64   `json:"impressions"`
					Clicks       int64   `json:"clicks"`
					Cost         float64 `json:"cost"`
					Conversions  int     `json:"conversions"`
					ROI          float64 `json:"roi"`
					CTR          float64 `json:"ctr"`
					ECPM         float64 `json:"ecpm"`
					PayOrderCnt  int     `json:"pay_order_cnt"`
					PayOrderAmt  float64 `json:"pay_order_amt"`
				} `json:"list"`
				PageInfo struct {
					Page        int `json:"page"`
					PageSize    int `json:"page_size"`
					TotalNumber int `json:"total_number"`
					TotalPage   int `json:"total_page"`
				} `json:"page_info"`
			}
			if err := json.Unmarshal(resp.Data, &result); err != nil {
				break
			}

			for _, item := range result.List {
				var ad models.UniAd
				if err := db.DB.Where("qianchuan_ad_id = ? AND account_id = ?", item.AdID, acc.ID).First(&ad).Error; err != nil {
					continue
				}

				parsedTime, err := time.Parse("2006-01-02 15:04", item.StatDatetime)
				if err != nil {
					parsedTime, err = time.Parse("2006-01-02 15:04:05", item.StatDatetime)
					if err != nil {
						continue
					}
				}

				reportDate := time.Date(parsedTime.Year(), parsedTime.Month(), parsedTime.Day(), 0, 0, 0, 0, parsedTime.Location())
				reportHour := parsedTime.Hour()

				report := models.UniAdReport{
					UniAdID:      ad.ID,
					ReportDate:   reportDate,
					ReportHour:   reportHour,
					Impressions:  item.Impressions,
					Clicks:       item.Clicks,
					Cost:         item.Cost,
					Conversions:  item.Conversions,
					ROI:          item.ROI,
					CTR:          item.CTR,
					ECPM:         item.ECPM,
					PayOrderCnt:  item.PayOrderCnt,
					PayOrderAmt:  item.PayOrderAmt,
				}

				db.DB.Clauses(clause.OnConflict{
					Columns:   []clause.Column{{Name: "uni_ad_id"}, {Name: "report_date"}, {Name: "report_hour"}},
					DoUpdates: clause.AssignmentColumns([]string{
						"impressions", "clicks", "cost", "conversions",
						"roi", "ctr", "ecpm", "pay_order_cnt", "pay_order_amt",
					}),
				}).Create(&report)
			}

			if page >= result.PageInfo.TotalPage {
				break
			}
			page++
		}
	}
}

func (w *SyncWorker) syncCreatives() {
	var accounts []models.QianchuanAccount
	db.DB.Where("status = ?", "active").Find(&accounts)

	for _, acc := range accounts {
		accRef := qianchuan.AdAccount{ID: acc.ID, AdvertiserID: acc.AdvertiserID}

		var ads []models.UniAd
		if err := db.DB.Where("account_id = ? AND qianchuan_ad_id IS NOT NULL", acc.ID).Find(&ads).Error; err != nil {
			continue
		}

		for _, ad := range ads {
			resp, err := w.QC.ListCreatives(&accRef, 1, 100, *ad.QianchuanAdID)
			if err != nil || resp.Code != 0 {
				log.Printf("sync creatives failed for ad %d: %v", ad.ID, err)
				continue
			}

			var result struct {
				List []struct {
					CreativeID       int64                  `json:"creative_id"`
					Title            string                 `json:"title"`
					ImageMode        string                 `json:"image_mode"`
					CreativeMaterial map[string]interface{} `json:"creative_material"`
					Metrics          map[string]interface{} `json:"metrics"`
				} `json:"list"`
			}
			if err := json.Unmarshal(resp.Data, &result); err != nil {
				continue
			}

			for _, item := range result.List {
				var creative models.Creative
				err := db.DB.Where("qianchuan_creative_id = ? AND account_id = ?", item.CreativeID, acc.ID).First(&creative).Error
				if err != nil {
					creative = models.Creative{
						QianchuanCreativeID: &item.CreativeID,
						AccountID:   acc.ID,
						Name:        item.Title,
						Type:        item.ImageMode,
						URL:         "",
						Tags:        models.JSONMap{},
						MetricsJSON: models.JSONMap{},
					}
					if item.CreativeMaterial != nil {
						if urlStr, ok := item.CreativeMaterial["url"].(string); ok {
							creative.URL = urlStr
						}
					}
					if item.Metrics != nil {
						creative.MetricsJSON = models.JSONMap(item.Metrics)
					}
					db.DB.Create(&creative)
				} else {
					updates := map[string]interface{}{}
					if item.Metrics != nil {
						updates["metrics_json"] = models.JSONMap(item.Metrics)
					}
					if len(updates) > 0 {
						db.DB.Model(&creative).Updates(updates)
					}
				}

				var link models.UniAdCreative
				if err := db.DB.Where("uni_ad_id = ? AND creative_id = ?", ad.ID, creative.ID).First(&link).Error; err != nil {
					db.DB.Create(&models.UniAdCreative{
						UniAdID:    ad.ID,
						CreativeID: creative.ID,
					})
				}
			}
		}
	}
}

func (w *SyncWorker) syncAccountInfo() {
	var accounts []models.QianchuanAccount
	db.DB.Find(&accounts)

	for _, acc := range accounts {
		accRef := qianchuan.AdAccount{ID: acc.ID, AdvertiserID: acc.AdvertiserID}
		resp, err := w.QC.GetAccountInfo(&accRef)
		if err != nil || resp.Code != 0 {
			log.Printf("sync account info failed for account %d: %v", acc.ID, err)
			continue
		}

		var info struct {
			AdvertiserID   int64   `json:"advertiser_id"`
			AdvertiserName string  `json:"advertiser_name"`
			Status         string  `json:"status"`
			Balance        float64 `json:"balance"`
			ValidBalance   float64 `json:"valid_balance"`
		}
		if err := json.Unmarshal(resp.Data, &info); err != nil {
			continue
		}

		status := "active"
		if info.Status != "STATUS_ENABLE" && info.Status != "" {
			status = "inactive"
		}
		updates := map[string]interface{}{
			"balance":      info.ValidBalance,
			"status":       status,
			"last_sync_at": time.Now(),
		}
		if info.ValidBalance == 0 && info.Balance > 0 {
			updates["balance"] = info.Balance
		}
		db.DB.Model(&acc).Updates(updates)
	}
}
