package controller

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/trihackathon/api/models"
	"github.com/trihackathon/api/requests"
	"github.com/trihackathon/api/response"
	"gorm.io/gorm"
)

type UserController struct {
	db *gorm.DB
}

func NewUserController(db *gorm.DB) *UserController {
	return &UserController{db: db}
}

// GetMe 自分のユーザー情報を取得
// @Summary      自分のユーザー情報を取得
// @Tags         users
// @Produce      json
// @Success      200  {object}  response.UserResponse
// @Failure      404  {object}  response.ErrorResponse
// @Router       /api/users/me [get]
// @Security     BearerAuth
func (ctrl *UserController) GetMe(c echo.Context) error {
	uid := c.Get("uid").(string)

	var user models.User
	if err := ctrl.db.First(&user, "id = ?", uid).Error; err != nil {
		return c.JSON(http.StatusNotFound, response.ErrorResponse{
			Error:   "user_not_found",
			Message: "ユーザーが見つかりません",
		})
	}

	return c.JSON(http.StatusOK, response.NewUserResponse(user))
}

// CreateMe ユーザーを作成（初回登録）
// @Summary      ユーザー作成
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        body  body      requests.CreateUserRequest  true  "ユーザー情報"
// @Success      201   {object}  response.UserResponse
// @Failure      400   {object}  response.ErrorResponse
// @Failure      409   {object}  response.ErrorResponse
// @Router       /api/users/me [post]
// @Security     BearerAuth
func (ctrl *UserController) CreateMe(c echo.Context) error {
	uid := c.Get("uid").(string)

	// 既に存在するかチェック
	var existing models.User
	if err := ctrl.db.First(&existing, "id = ?", uid).Error; err == nil {
		return c.JSON(http.StatusConflict, response.ErrorResponse{
			Error:   "user_already_exists",
			Message: "ユーザーは既に登録されています",
		})
	}

	req := new(requests.CreateUserRequest)
	if err := c.Bind(req); err != nil {
		return c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error:   "invalid_request",
			Message: "リクエストの形式が不正です",
		})
	}

	user := models.User{
		ID:   uid,
		Name: req.Name,
		Age:  req.Age,
	}

	if err := ctrl.db.Create(&user).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Error:   "create_failed",
			Message: "ユーザーの作成に失敗しました",
		})
	}

	return c.JSON(http.StatusCreated, response.NewUserResponse(user))
}

// UpdateMe 自分のユーザー情報を更新
// @Summary      自分のユーザー情報を更新
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        body  body      requests.UpdateUserRequest  true  "更新するユーザー情報"
// @Success      200   {object}  response.UserResponse
// @Failure      400   {object}  response.ErrorResponse
// @Failure      404   {object}  response.ErrorResponse
// @Router       /api/users/me [put]
// @Security     BearerAuth
func (ctrl *UserController) UpdateMe(c echo.Context) error {
	uid := c.Get("uid").(string)

	var user models.User
	if err := ctrl.db.First(&user, "id = ?", uid).Error; err != nil {
		return c.JSON(http.StatusNotFound, response.ErrorResponse{
			Error:   "user_not_found",
			Message: "ユーザーが見つかりません",
		})
	}

	req := new(requests.UpdateUserRequest)
	if err := c.Bind(req); err != nil {
		return c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error:   "invalid_request",
			Message: "リクエストの形式が不正です",
		})
	}

	user.Name = req.Name
	user.Age = req.Age

	if err := ctrl.db.Save(&user).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Error:   "update_failed",
			Message: "ユーザー情報の更新に失敗しました",
		})
	}

	return c.JSON(http.StatusOK, response.NewUserResponse(user))
}
