package controller

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/trihackathon/api/requests"
	"github.com/trihackathon/api/response"
	"gorm.io/gorm"
)

type GymController struct {
	db *gorm.DB
}

func NewGymController(db *gorm.DB) *GymController {
	return &GymController{db: db}
}

// CreateGymLocation ジム位置登録
// @Summary      ジム位置登録
// @Description  ジムの位置情報を登録する。ジオフェンス半径のデフォルトは100m。
// @Tags         gym
// @Accept       json
// @Produce      json
// @Param        body  body      requests.CreateGymLocationRequest  true  "ジム位置情報"
// @Success      201   {object}  response.GymLocationResponse
// @Failure      400   {object}  response.ErrorResponse
// @Router       /api/gym-locations [post]
// @Security     BearerAuth
func (ctrl *GymController) CreateGymLocation(c echo.Context) error {
	uid := c.Get("uid").(string)

	req := new(requests.CreateGymLocationRequest)
	if err := c.Bind(req); err != nil {
		return c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error:   "invalid_request",
			Message: "リクエストの形式が不正です",
		})
	}

	// TODO: ビジネスロジック実装
	_ = uid

	return c.JSON(http.StatusCreated, response.GymLocationResponse{})
}

// GetGymLocations 登録ジム一覧
// @Summary      登録ジム一覧
// @Description  ユーザーが登録したジムの一覧を取得する
// @Tags         gym
// @Produce      json
// @Success      200  {array}   response.GymLocationResponse
// @Router       /api/gym-locations [get]
// @Security     BearerAuth
func (ctrl *GymController) GetGymLocations(c echo.Context) error {
	uid := c.Get("uid").(string)

	// TODO: ビジネスロジック実装
	_ = uid

	return c.JSON(http.StatusOK, []response.GymLocationResponse{})
}

// DeleteGymLocation ジム位置削除
// @Summary      ジム位置削除
// @Description  登録したジムの位置情報を削除する。所有者のみ削除可能。
// @Tags         gym
// @Param        locationId  path  string  true  "ジム位置ID"
// @Success      204  "No Content"
// @Failure      403  {object}  response.ErrorResponse
// @Failure      404  {object}  response.ErrorResponse
// @Router       /api/gym-locations/{locationId} [delete]
// @Security     BearerAuth
func (ctrl *GymController) DeleteGymLocation(c echo.Context) error {
	uid := c.Get("uid").(string)
	locationId := c.Param("locationId")

	// TODO: ビジネスロジック実装
	_ = uid
	_ = locationId

	return c.NoContent(http.StatusNoContent)
}

// GymCheckin ジムチェックイン
// @Summary      ジムチェックイン
// @Description  ジムにチェックインする。現在位置とジムの距離を検証し、radius_m以内であることを確認。
// @Tags         gym
// @Accept       json
// @Produce      json
// @Param        body  body      requests.GymCheckinRequest  true  "チェックイン情報"
// @Success      201   {object}  response.ActivityResponse
// @Failure      404   {object}  response.ErrorResponse
// @Failure      409   {object}  response.ErrorResponse
// @Failure      422   {object}  response.ErrorResponse
// @Router       /api/activities/gym/checkin [post]
// @Security     BearerAuth
func (ctrl *GymController) GymCheckin(c echo.Context) error {
	uid := c.Get("uid").(string)

	req := new(requests.GymCheckinRequest)
	if err := c.Bind(req); err != nil {
		return c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error:   "invalid_request",
			Message: "リクエストの形式が不正です",
		})
	}

	// TODO: ビジネスロジック実装
	_ = uid

	return c.JSON(http.StatusCreated, response.ActivityResponse{})
}

// GymCheckout ジムチェックアウト
// @Summary      ジムチェックアウト
// @Description  ジムからチェックアウトする。duration_minをended_at - started_atから算出。
// @Tags         gym
// @Accept       json
// @Produce      json
// @Param        activityId  path      string                       true  "アクティビティID"
// @Param        body        body      requests.GymCheckoutRequest  true  "チェックアウト情報"
// @Success      200         {object}  response.ActivityResponse
// @Failure      403         {object}  response.ErrorResponse
// @Failure      404         {object}  response.ErrorResponse
// @Failure      422         {object}  response.ErrorResponse
// @Router       /api/activities/gym/{activityId}/checkout [post]
// @Security     BearerAuth
func (ctrl *GymController) GymCheckout(c echo.Context) error {
	uid := c.Get("uid").(string)
	activityId := c.Param("activityId")

	req := new(requests.GymCheckoutRequest)
	if err := c.Bind(req); err != nil {
		return c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error:   "invalid_request",
			Message: "リクエストの形式が不正です",
		})
	}

	// TODO: ビジネスロジック実装
	_ = uid
	_ = activityId

	return c.JSON(http.StatusOK, response.ActivityResponse{})
}

// GetGymActivity ジム記録詳細
// @Summary      ジム記録詳細
// @Description  指定したジムアクティビティの詳細情報を取得する
// @Tags         gym
// @Produce      json
// @Param        activityId  path      string  true  "アクティビティID"
// @Success      200         {object}  response.ActivityResponse
// @Failure      403         {object}  response.ErrorResponse
// @Failure      404         {object}  response.ErrorResponse
// @Router       /api/activities/gym/{activityId} [get]
// @Security     BearerAuth
func (ctrl *GymController) GetGymActivity(c echo.Context) error {
	uid := c.Get("uid").(string)
	activityId := c.Param("activityId")

	// TODO: ビジネスロジック実装
	_ = uid
	_ = activityId

	return c.JSON(http.StatusOK, response.ActivityResponse{})
}
