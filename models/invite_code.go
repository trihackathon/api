package models

import "time"

type InviteCode struct {
	Code      string     `json:"code" gorm:"primaryKey;size:6"`
	TeamID    string     `json:"team_id" gorm:"not null"`
	CreatedBy string     `json:"created_by" gorm:"not null"`
	UsedBy    *string    `json:"used_by"`
	UsedAt    *time.Time `json:"used_at"`
	ExpiresAt time.Time  `json:"expires_at" gorm:"not null"`
	CreatedAt time.Time  `gorm:"autoCreateTime" json:"created_at"`

	Team Team `json:"team,omitempty" gorm:"foreignKey:TeamID"`
}
