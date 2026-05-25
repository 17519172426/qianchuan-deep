package models

import "time"

type Rule struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	Name          string    `gorm:"size:100;not null" json:"name"`
	Description   string    `gorm:"type:text" json:"description"`
	AccountID     uint      `gorm:"index;not null" json:"account_id"`
	ScopeJSON     JSONMap   `gorm:"type:jsonb;default:'{}'" json:"scope_json"`
	ConditionJSON JSONMap   `gorm:"type:jsonb;not null" json:"condition_json"`
	ActionJSON    JSONMap   `gorm:"type:jsonb;not null" json:"action_json"`
	Schedule      string    `gorm:"size:64;default:'*/5 * * * *'" json:"schedule"`
	Cooldown      string    `gorm:"size:32;default:'1h'" json:"cooldown"`
	Enabled       bool      `gorm:"default:false" json:"enabled"`
	CreatedAt     time.Time `json:"created_at"`
}

type RuleExecution struct {
	ID            uint       `gorm:"primaryKey" json:"id"`
	RuleID        uint       `gorm:"index;not null" json:"rule_id"`
	UniAdID       uint       `gorm:"index;not null" json:"uni_ad_id"`
	TriggeredAt   *time.Time `json:"triggered_at"`
	ConditionJSON JSONMap    `gorm:"type:jsonb" json:"condition_json"`
	ActionJSON    JSONMap    `gorm:"type:jsonb" json:"action_json"`
	Status        string     `gorm:"size:32;default:'success'" json:"status"`
	ResultJSON    JSONMap    `gorm:"type:jsonb" json:"result_json"`
	ExecutedAt    *time.Time `json:"executed_at"`
}
