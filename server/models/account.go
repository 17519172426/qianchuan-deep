package models

import (
	"database/sql"
	"time"
)

type QianchuanAccount struct {
	ID           uint         `gorm:"primaryKey" json:"id"`
	AccountName  string       `gorm:"size:255;not null" json:"account_name"`
	AdvertiserID int64        `gorm:"uniqueIndex;not null" json:"advertiser_id"`
	AccessToken  string       `gorm:"type:text;not null" json:"-"`
	RefreshToken string       `gorm:"type:text;not null" json:"-"`
	Status       string       `gorm:"size:50;default:active" json:"status"`
	Balance      float64      `gorm:"type:decimal(15,2);default:0" json:"balance"`
	LastSyncAt   sql.NullTime `json:"last_sync_at"`
	CreatedAt    time.Time    `json:"created_at"`
}
