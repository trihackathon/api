package controller

import (
	"fmt"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/oklog/ulid/v2"
	"github.com/trihackathon/api/adapter"
	"github.com/trihackathon/api/models"
	"github.com/trihackathon/api/response"
	"gorm.io/gorm"
)

type UserController struct {
	db *gorm.DB
	r2 *adapter.R2Adapter
}

func NewUserController(db *gorm.DB, r2 *adapter.R2Adapter) *UserController {
	return &UserController{db: db, r2: r2}
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
// @Accept       multipart/form-data
// @Produce      json
// @Param        name       formData  string  true   "名前"
// @Param        age        formData  int     true   "年齢"
// @Param        gender     formData  string  true   "性別 (male/female/other)"
// @Param        weight     formData  int     true   "体重 (kg)"
// @Param        chronotype formData  string  true   "朝型夜型 (morning/night/both)"
// @Param        avatar     formData  file    false  "プロフィール写真"
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

	name := c.FormValue("name")
	ageStr := c.FormValue("age")
	gender := c.FormValue("gender")
	weightStr := c.FormValue("weight")
	chronotype := c.FormValue("chronotype")

	if name == "" || ageStr == "" || gender == "" || weightStr == "" || chronotype == "" {
		return c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error:   "invalid_request",
			Message: "name, age, gender, weight, chronotype は必須です",
		})
	}

	age, err := strconv.Atoi(ageStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error:   "invalid_request",
			Message: "age は整数で指定してください",
		})
	}

	weight, err := strconv.Atoi(weightStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error:   "invalid_request",
			Message: "weight は整数で指定してください",
		})
	}

	var avatarURL string
	file, err := c.FormFile("avatar")
	if err == nil && file != nil {
		src, err := file.Open()
		if err != nil {
			return c.JSON(http.StatusBadRequest, response.ErrorResponse{
				Error:   "invalid_file",
				Message: "ファイルの読み込みに失敗しました",
			})
		}
		defer src.Close()

		ext := strings.ToLower(filepath.Ext(file.Filename))
		key := fmt.Sprintf("avatars/%s/%s%s", uid, ulid.Make().String(), ext)
		contentType := file.Header.Get("Content-Type")

		url, err := ctrl.r2.Upload(c.Request().Context(), key, src, contentType)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, response.ErrorResponse{
				Error:   "upload_failed",
				Message: "プロフィール写真のアップロードに失敗しました",
			})
		}
		avatarURL = url
	}

	user := models.User{
		ID:         uid,
		Name:       name,
		Age:        age,
		Gender:     gender,
		Weight:     weight,
		Chronotype: chronotype,
		AvatarURL:  avatarURL,
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
// @Accept       multipart/form-data
// @Produce      json
// @Param        name       formData  string  false  "名前"
// @Param        age        formData  int     false  "年齢"
// @Param        gender     formData  string  false  "性別 (male/female/other)"
// @Param        weight     formData  int     false  "体重 (kg)"
// @Param        chronotype formData  string  false  "朝型夜型 (morning/night/both)"
// @Param        avatar     formData  file    false  "プロフィール写真"
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

	// 部分更新: 値がある場合のみ更新
	if name := c.FormValue("name"); name != "" {
		user.Name = name
	}
	if ageStr := c.FormValue("age"); ageStr != "" {
		age, err := strconv.Atoi(ageStr)
		if err != nil {
			return c.JSON(http.StatusBadRequest, response.ErrorResponse{
				Error:   "invalid_request",
				Message: "age は整数で指定してください",
			})
		}
		user.Age = age
	}
	if gender := c.FormValue("gender"); gender != "" {
		user.Gender = gender
	}
	if weightStr := c.FormValue("weight"); weightStr != "" {
		weight, err := strconv.Atoi(weightStr)
		if err != nil {
			return c.JSON(http.StatusBadRequest, response.ErrorResponse{
				Error:   "invalid_request",
				Message: "weight は整数で指定してください",
			})
		}
		user.Weight = weight
	}
	if chronotype := c.FormValue("chronotype"); chronotype != "" {
		user.Chronotype = chronotype
	}

	// アバター更新
	file, err := c.FormFile("avatar")
	if err == nil && file != nil {
		src, err := file.Open()
		if err != nil {
			return c.JSON(http.StatusBadRequest, response.ErrorResponse{
				Error:   "invalid_file",
				Message: "ファイルの読み込みに失敗しました",
			})
		}
		defer src.Close()

		// 古いアバターを削除
		if user.AvatarURL != "" {
			oldKey := extractR2Key(user.AvatarURL)
			if oldKey != "" {
				_ = ctrl.r2.Delete(c.Request().Context(), oldKey)
			}
		}

		ext := strings.ToLower(filepath.Ext(file.Filename))
		key := fmt.Sprintf("avatars/%s/%s%s", uid, ulid.Make().String(), ext)
		contentType := file.Header.Get("Content-Type")

		url, err := ctrl.r2.Upload(c.Request().Context(), key, src, contentType)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, response.ErrorResponse{
				Error:   "upload_failed",
				Message: "プロフィール写真のアップロードに失敗しました",
			})
		}
		user.AvatarURL = url
	}

	if err := ctrl.db.Save(&user).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Error:   "update_failed",
			Message: "ユーザー情報の更新に失敗しました",
		})
	}

	return c.JSON(http.StatusOK, response.NewUserResponse(user))
}

// extractR2Key はR2の公開URLからキーを抽出する
func extractR2Key(url string) string {
	idx := strings.Index(url, "avatars/")
	if idx == -1 {
		return ""
	}
	return url[idx:]
}
