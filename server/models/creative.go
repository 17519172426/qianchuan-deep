package models

import "time"

type Creative struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	AccountID   uint      `gorm:"index;not null" json:"account_id"`
	Name        string    `gorm:"size:255;not null" json:"name"`
	Type        string    `gorm:"size:20;not null" json:"type"`
	URL         string    `gorm:"type:text;not null" json:"url"`
	FileSize    int64     `gorm:"default:0" json:"file_size"`
	Duration    float64   `gorm:"default:0" json:"duration"`
	Tags        JSONMap   `gorm:"type:jsonb;default:'[]'" json:"tags"`
	MetricsJSON JSONMap   `gorm:"type:jsonb;default:'{}'" json:"metrics_json"`
	CreatedAt   time.Time `json:"created_at"`
}

type UniAdCreative struct {
	UniAdID    uint `gorm:"primaryKey" json:"uni_ad_id"`
	CreativeID uint `gorm:"primaryKey" json:"creative_id"`
	IsBlocked  bool `gorm:"default:false" json:"is_blocked"`
}

type Rule struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	Name          string    `gorm:"size:100;not null" json:"name"`
	Description   string    `gorm:"type:text;default:''" json:"description"`
	AccountID     uint      `gorm:"index" json:"account_id"`
	ScopeJSON     JSONMap   `gorm:"type:jsonb;default:'{}'" json:"scope_json"`
	ConditionJSON JSONMap   `gorm:"type:jsonb;not null;default:'{}'" json:"condition_json"`
	ActionJSON    JSONMap   `gorm:"type:jsonb;not null;default:'{}'" json:"action_json"`
	Schedule      string    `gorm:"size:50;default:*/5 * * * *" json:"schedule"`
	Cooldown      string    `gorm:"size:20;default:1h" json:"cooldown"`
	Enabled       bool      `gorm:"default:false" json:"enabled"`
	CreatedAt     time.Time `json:"created_at"`
}

type RuleExecution struct {
	ID            uint       `gorm:"primaryKey" json:"id"`
	RuleID        uint       `json:"rule_id"`
	UniAdID       uint       `json:"uni_ad_id"`
	TriggeredAt   *time.Time `json:"triggered_at"`
	ConditionJSON JSONMap    `gorm:"type:jsonb;default:'{}'" json:"condition_json"`
	ActionJSON    JSONMap    `gorm:"type:jsonb;default:'{}'" json:"action_json"`
	Status        string     `gorm:"size:20;default:pending" json:"status"`
	ResultJSON    JSONMap    `gorm:"type:jsonb;default:'{}'" json:"result_json"`
	ExecutedAt    *time.Time `json:"executed_at"`
}

type AIRecommendation struct {
	ID              uint       `gorm:"primaryKey" json:"id"`
	UniAdID         *uint      `json:"uni_ad_id"`
	Type            string     `gorm:"size:30;not null" json:"type"`
	Title           string     `gorm:"size:255;not null" json:"title"`
	Description     string     `gorm:"type:text;default:''" json:"description"`
	MetricsJSON     JSONMap    `gorm:"type:jsonb;default:'{}'" json:"metrics_json"`
	Confidence      float64    `gorm:"default:0" json:"confidence"`
	SuggestedAction JSONMap    `gorm:"type:jsonb;default:'{}'" json:"suggested_action"`
	Status          string     `gorm:"size:20;default:pending" json:"status"`
	ReviewedBy      *uint      `json:"reviewed_by"`
	ReviewedAt      *time.Time `json:"reviewed_at"`
	CreatedAt       time.Time  `json:"created_at"`
}
