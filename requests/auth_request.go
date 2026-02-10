package requests

// CreateUserRequest ユーザー作成リクエスト
type CreateUserRequest struct {
	Name string `json:"name" example:"山田太郎"`
	Age  int    `json:"age" example:"25"`
}

// UpdateUserRequest ユーザー更新リクエスト
type UpdateUserRequest struct {
	Name string `json:"name" example:"山田太郎"`
	Age  int    `json:"age" example:"25"`
}
