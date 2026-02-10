package models

import "time"

type User struct {
	ID        string    `json:"id" gorm:"primaryKey"` // FirebaseUID を主キーにする
	Name      string    `json:"name" gorm:"not null"`
	Age       int       `json:"age" gorm:"not null"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}
