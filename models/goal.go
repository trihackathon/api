package models

import "time"

type Goal struct {
	ID                   string    `json:"id" gorm:"primaryKey"`
	TeamID               string    `json:"team_id" gorm:"not null;uniqueIndex"`
	ExerciseType         string    `json:"exercise_type" gorm:"not null"` // running / gym
	TargetDistanceKM     *float64  `json:"target_distance_km"`           // running用
	TargetVisitsPerWeek  *int      `json:"target_visits_per_week"`       // gym用
	TargetMinDurationMin *int      `json:"target_min_duration_min"`      // gym用
	CreatedAt            time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt            time.Time `gorm:"autoUpdateTime" json:"updated_at"`

	Team Team `json:"team,omitempty" gorm:"foreignKey:TeamID"`
}
