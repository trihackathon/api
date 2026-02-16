package models

import "time"

type GymLocation struct {
	ID        string    `json:"id" gorm:"primaryKey"`
	UserID    string    `json:"user_id" gorm:"not null;index"`
	Name      string    `json:"name" gorm:"not null"`
	Latitude  float64   `json:"latitude" gorm:"not null"`
	Longitude float64   `json:"longitude" gorm:"not null"`
	RadiusM   int       `json:"radius_m" gorm:"default:100"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`

	User User `json:"user,omitempty" gorm:"foreignKey:UserID"`
}
