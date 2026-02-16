package models

import "time"

type WeeklyEvaluation struct {
	ID               string    `json:"id" gorm:"primaryKey"`
	TeamID           string    `json:"team_id" gorm:"not null;index:idx_eval_team_week"`
	UserID           string    `json:"user_id" gorm:"not null;index:idx_eval_team_week"`
	WeekNumber       int       `json:"week_number" gorm:"not null;index:idx_eval_team_week"`
	TargetMet        bool      `json:"target_met" gorm:"default:false"`
	TotalDistanceKM  float64   `json:"total_distance_km" gorm:"default:0"`
	TotalVisits      int       `json:"total_visits" gorm:"default:0"`
	TotalDurationMin int       `json:"total_duration_min" gorm:"default:0"`
	HPChange         int       `json:"hp_change" gorm:"default:0"`
	EvaluatedAt      time.Time `json:"evaluated_at"`
	CreatedAt        time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt        time.Time `gorm:"autoUpdateTime" json:"updated_at"`

	User User `json:"user,omitempty" gorm:"foreignKey:UserID"`
	Team Team `json:"team,omitempty" gorm:"foreignKey:TeamID"`
}
