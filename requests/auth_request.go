package requests

// CreateUserRequest ユーザー作成リクエスト（multipart/form-data）
type CreateUserRequest struct {
	Name       string `form:"name"`
	Age        int    `form:"age"`
	Gender     string `form:"gender"`
	Weight     int    `form:"weight"`
	Chronotype string `form:"chronotype"`
	// avatar は c.FormFile で取得
}

// UpdateUserRequest ユーザー更新リクエスト（multipart/form-data、部分更新）
type UpdateUserRequest struct {
	Name       string `form:"name"`
	Age        int    `form:"age"`
	Gender     string `form:"gender"`
	Weight     int    `form:"weight"`
	Chronotype string `form:"chronotype"`
	// avatar は c.FormFile で取得
}
