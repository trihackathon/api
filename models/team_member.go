package models

import "time"

type TeamMember struct {
	ID       string    `json:"id" gorm:"primaryKey"`
	TeamID   string    `json:"team_id" gorm:"not null;uniqueIndex:idx_team_user"`
	UserID   string    `json:"user_id" gorm:"not null;uniqueIndex:idx_team_user"`
	Role     string    `json:"role" gorm:"default:'member'"` // leader / member
	JoinedAt time.Time `json:"joined_at" gorm:"autoCreateTime"`

	User User `json:"user,omitempty" gorm:"foreignKey:UserID"`
	Team Team `json:"team,omitempty" gorm:"foreignKey:TeamID"`
}
