package models

import "time"

type UniAdReport struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	UniAdID     uint      `gorm:"index;not null" json:"uni_ad_id"`
	ReportDate  time.Time `gorm:"index;not null" json:"report_date"`
	ReportHour  int       `gorm:"default:0" json:"report_hour"`
	Impressions int64     `gorm:"default:0" json:"impressions"`
	Clicks      int64     `gorm:"default:0" json:"clicks"`
	Cost        float64   `gorm:"type:decimal(15,2);default:0" json:"cost"`
	Conversions int       `gorm:"default:0" json:"conversions"`
	ROI         float64   `gorm:"type:decimal(10,4);default:0" json:"roi"`
	CTR         float64   `gorm:"type:decimal(10,4);default:0" json:"ctr"`
	ECPM        float64   `gorm:"type:decimal(15,4);default:0" json:"ecpm"`
	PayOrderCnt int       `gorm:"default:0" json:"pay_order_cnt"`
	PayOrderAmt float64   `gorm:"type:decimal(15,2);default:0" json:"pay_order_amt"`
	CreatedAt   time.Time `json:"created_at"`
}
