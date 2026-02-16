package response

import (
	"time"

	"github.com/trihackathon/api/models"
)

// UserResponse ユーザー情報レスポンス
type UserResponse struct {
	ID         string `json:"id" example:"firebaseUID123"`
	Name       string `json:"name" example:"山田太郎"`
	Age        int    `json:"age" example:"25"`
	Gender     string `json:"gender" example:"male"`
	Weight     int    `json:"weight" example:"70"`
	Chronotype string `json:"chronotype" example:"morning"`
	AvatarURL  string `json:"avatar_url" example:"https://r2.example.com/avatars/uid/image.jpg"`
	CreatedAt  string `json:"created_at" example:"2024-01-01T00:00:00Z"`
	UpdatedAt  string `json:"updated_at" example:"2024-01-01T00:00:00Z"`
}

func NewUserResponse(user models.User) UserResponse {
	return UserResponse{
		ID:         user.ID,
		Name:       user.Name,
		Age:        user.Age,
		Gender:     user.Gender,
		Weight:     user.Weight,
		Chronotype: user.Chronotype,
		AvatarURL:  user.AvatarURL,
		CreatedAt:  user.CreatedAt.Format(time.RFC3339),
		UpdatedAt:  user.UpdatedAt.Format(time.RFC3339),
	}
}

// ErrorResponse エラーレスポンス
type ErrorResponse struct {
	Error   string `json:"error" example:"invalid_token"`
	Message string `json:"message" example:"エラーが発生しました"`
}
