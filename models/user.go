package models

import "time"

type User struct {
	ID         string    `json:"id" gorm:"primaryKey"` // FirebaseUID を主キーにする
	Name       string    `json:"name" gorm:"not null"`
	Age        int       `json:"age" gorm:"not null"`
	Gender     string    `json:"gender" gorm:"default:'other'"`    // male / female / other
	Weight     int       `json:"weight" gorm:"default:60"`         // kg
	Chronotype string    `json:"chronotype" gorm:"default:'both'"` // morning / night / both
	AvatarURL  string    `json:"avatar_url" gorm:"default:''"`
	CreatedAt  time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt  time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}
