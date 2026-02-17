package models

import "time"

type ActivityReview struct {
	ID         string    `json:"id" gorm:"primaryKey"`
	ActivityID string    `json:"activity_id" gorm:"not null;uniqueIndex:idx_activity_reviewer"`
	ReviewerID string    `json:"reviewer_id" gorm:"not null;uniqueIndex:idx_activity_reviewer"`
	Status     string    `json:"status" gorm:"not null"` // "approved" | "rejected"
	Comment    string    `json:"comment"`
	CreatedAt  time.Time `json:"created_at" gorm:"autoCreateTime"`

	Reviewer User `json:"reviewer,omitempty" gorm:"foreignKey:ReviewerID"`
}
