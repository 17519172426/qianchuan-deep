package worker

import (
	"encoding/json"
	"log"
	"time"

	"github.com/example/qianchuan-saas/db"
	"github.com/example/qianchuan-saas/models"
	"github.com/example/qianchuan-saas/qianchuan"
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
	var ads []models.UniAd
	if err := db.DB.Where("qianchuan_ad_id IS NOT NULL").Find(&ads).Error; err != nil {
		log.Printf("sync reports: fetch ads failed: %v", err)
		return
	}

	for _, ad := range ads {
		var acc models.QianchuanAccount
		if err := db.DB.First(&acc, ad.AccountID).Error; err != nil {
			continue
		}

		accRef := qianchuan.AdAccount{ID: acc.ID, AdvertiserID: acc.AdvertiserID}
		resp, err := w.QC.GetUniAdDetail(&accRef, *ad.QianchuanAdID)
		if err != nil || resp.Code != 0 {
			continue
		}

		var detail struct {
			Metrics struct {
				Impressions  int64   `json:"impressions"`
				Clicks       int64   `json:"clicks"`
				Cost         float64 `json:"cost"`
				Conversions  int     `json:"conversions"`
				ROI          float64 `json:"roi"`
				CTR          float64 `json:"ctr"`
				ECPM         float64 `json:"ecpm"`
				PayOrderCnt  int     `json:"pay_order_cnt"`
				PayOrderAmt  float64 `json:"pay_order_amt"`
			} `json:"metrics"`
		}
		if err := json.Unmarshal(resp.Data, &detail); err != nil {
			continue
		}

		now := time.Now()
		report := models.UniAdReport{
			UniAdID:     ad.ID,
			ReportDate:  time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()),
			ReportHour:  now.Hour(),
			Impressions: detail.Metrics.Impressions,
			Clicks:      detail.Metrics.Clicks,
			Cost:        detail.Metrics.Cost,
			Conversions: detail.Metrics.Conversions,
			ROI:         detail.Metrics.ROI,
			CTR:         detail.Metrics.CTR,
			ECPM:        detail.Metrics.ECPM,
			PayOrderCnt: detail.Metrics.PayOrderCnt,
			PayOrderAmt: detail.Metrics.PayOrderAmt,
		}
		db.DB.Create(&report)
	}
}
