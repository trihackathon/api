package models

import "time"

type Team struct {
	ID           string     `json:"id" gorm:"primaryKey"`
	Name         string     `json:"name" gorm:"not null"`
	ExerciseType string     `json:"exercise_type" gorm:"not null"`          // running / gym
	Strictness   string     `json:"strictness" gorm:"default:'normal'"`     // normal / strict / relaxed
	Status       string     `json:"status" gorm:"default:'forming'"`        // forming / active / completed / disbanded
	MaxHP        int        `json:"max_hp" gorm:"default:100"`
	CurrentHP    int        `json:"current_hp" gorm:"default:100"`
	CurrentWeek  int        `json:"current_week" gorm:"default:0"`
	StartedAt    *time.Time `json:"started_at"`
	CreatedAt    time.Time  `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt    time.Time  `gorm:"autoUpdateTime" json:"updated_at"`

	Members []TeamMember `json:"members,omitempty" gorm:"foreignKey:TeamID"`
}
