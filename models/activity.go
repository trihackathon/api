package models

import "time"

type Activity struct {
	ID            string     `json:"id" gorm:"primaryKey"`
	UserID        string     `json:"user_id" gorm:"not null;index"`
	TeamID        *string    `json:"team_id" gorm:"index"`          // nullable
	ExerciseType  string     `json:"exercise_type" gorm:"not null"` // running / gym
	Status        string     `json:"status" gorm:"default:'in_progress'"`
	StartedAt     time.Time  `json:"started_at" gorm:"not null"`
	EndedAt       *time.Time `json:"ended_at"`
	DistanceKM    float64    `json:"distance_km" gorm:"default:0"`
	GymLocationID *string    `json:"gym_location_id"`
	AutoDetected  bool       `json:"auto_detected" gorm:"default:false"`
	DurationMin   int        `json:"duration_min" gorm:"default:0"`
	CreatedAt     time.Time  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt     time.Time  `json:"updated_at" gorm:"autoUpdateTime"`

	User      User       `json:"user,omitempty" gorm:"foreignKey:UserID"`
	GPSPoints []GPSPoint `json:"gps_points,omitempty" gorm:"foreignKey:ActivityID"`
}

type GPSPoint struct {
	ID         string    `json:"id" gorm:"primaryKey"`
	ActivityID string    `json:"activity_id" gorm:"not null;index:idx_activity_timestamp"`
	ClientID   *string   `json:"client_id" gorm:"uniqueIndex:idx_client_id"` // PWA重複防止用
	Latitude   float64   `json:"latitude" gorm:"not null"`
	Longitude  float64   `json:"longitude" gorm:"not null"`
	Accuracy   float64   `json:"accuracy"`
	Timestamp  time.Time `json:"timestamp" gorm:"not null;index:idx_activity_timestamp"`
}
