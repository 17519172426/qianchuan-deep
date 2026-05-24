package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"
)

type JSONMap map[string]interface{}

func (j JSONMap) Value() (driver.Value, error) {
	return json.Marshal(j)
}

func (j *JSONMap) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(bytes, j)
}

type UniAd struct {
	ID              uint      `gorm:"primaryKey" json:"id"`
	AccountID       uint      `gorm:"index;not null" json:"account_id"`
	QianchuanAdID   *int64    `json:"qianchuan_ad_id"`
	Name            string    `gorm:"size:100;not null" json:"name"`
	MarketingGoal   string    `gorm:"size:32;not null" json:"marketing_goal"`
	AwemeID         *int64    `json:"aweme_id"`
	ProductIDs      JSONMap   `gorm:"type:jsonb;default:'[]'" json:"product_ids"`
	DeliverySetting JSONMap   `gorm:"type:jsonb;not null;default:'{}'" json:"delivery_setting"`
	CreativeSetting JSONMap   `gorm:"type:jsonb;not null;default:'{}'" json:"creative_setting"`
	Status          string    `gorm:"size:32;default:create" json:"status"`
	MetricsJSON     JSONMap   `gorm:"type:jsonb;default:'{}'" json:"metrics_json"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}
