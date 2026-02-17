package models

import "time"

type DisbandVote struct {
	ID        string    `json:"id" gorm:"primaryKey"`
	TeamID    string    `json:"team_id" gorm:"not null;uniqueIndex:idx_disband_team_user"`
	UserID    string    `json:"user_id" gorm:"not null;uniqueIndex:idx_disband_team_user"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`

	User User `json:"user,omitempty" gorm:"foreignKey:UserID"`
	Team Team `json:"team,omitempty" gorm:"foreignKey:TeamID"`
}
