package models

import "time"

type TeamMember struct {
	ID               string    `json:"id" gorm:"primaryKey"`
	TeamID           string    `json:"team_id" gorm:"not null;uniqueIndex:idx_team_user"`
	UserID           string    `json:"user_id" gorm:"not null;uniqueIndex:idx_team_user"`
	Role             string    `json:"role" gorm:"default:'member'"` // leader / member
	JoinedAt         time.Time `json:"joined_at" gorm:"autoCreateTime"`
	TargetMultiplier float64   `json:"target_multiplier" gorm:"default:1"` // 翌週目標倍率（1.0=通常, 1.5=チーム未達成時に達成者へ課す）

	User User `json:"user,omitempty" gorm:"foreignKey:UserID"`
	Team Team `json:"team,omitempty" gorm:"foreignKey:TeamID"`
}
