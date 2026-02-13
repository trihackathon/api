package controller

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/trihackathon/api/requests"
	"github.com/trihackathon/api/response"
	"gorm.io/gorm"
)

type ActivityController struct {
	db *gorm.DB
}

func NewActivityController(db *gorm.DB) *ActivityController {
	return &ActivityController{db: db}
}

// StartRunning ランニング開始
// @Summary      ランニング開始
// @Description  ランニングアクティビティを開始する。同時に進行中にできるアクティビティは1つのみ。チームがactive状態かつexercise_typeがrunningの場合のみ。
// @Tags         activities-running
// @Accept       json
// @Produce      json
// @Param        body  body      requests.StartRunningRequest  true  "開始地点情報"
// @Success      201   {object}  response.ActivityResponse
// @Failure      404   {object}  response.ErrorResponse
// @Failure      409   {object}  response.ErrorResponse
// @Failure      422   {object}  response.ErrorResponse
// @Router       /api/activities/running/start [post]
// @Security     BearerAuth
func (ctrl *ActivityController) StartRunning(c echo.Context) error {
	uid := c.Get("uid").(string)

	req := new(requests.StartRunningRequest)
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

// FinishRunning ランニング完了
// @Summary      ランニング完了
// @Description  ランニングアクティビティを完了する。GPSポイントから総移動距離を再計算しdistance_kmを確定。
// @Tags         activities-running
// @Accept       json
// @Produce      json
// @Param        activityId  path      string                       true  "アクティビティID"
// @Param        body        body      requests.FinishRunningRequest  true  "終了地点情報"
// @Success      200         {object}  response.ActivityResponse
// @Failure      403         {object}  response.ErrorResponse
// @Failure      404         {object}  response.ErrorResponse
// @Failure      422         {object}  response.ErrorResponse
// @Router       /api/activities/running/{activityId}/finish [post]
// @Security     BearerAuth
func (ctrl *ActivityController) FinishRunning(c echo.Context) error {
	uid := c.Get("uid").(string)
	activityId := c.Param("activityId")

	req := new(requests.FinishRunningRequest)
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

// SendGPSPoints GPSポイント送信（バッチ）
// @Summary      GPSポイント送信（バッチ）
// @Description  バックグラウンドで蓄積したGPSデータをバッチ送信する。精度が50mを超えるポイントは距離計算から除外（保存はする）。
// @Tags         activities-running
// @Accept       json
// @Produce      json
// @Param        activityId  path      string                         true  "アクティビティID"
// @Param        body        body      requests.SendGPSPointsRequest  true  "GPSポイントデータ"
// @Success      200         {object}  response.SendGPSPointsResponse
// @Failure      400         {object}  response.ErrorResponse
// @Failure      403         {object}  response.ErrorResponse
// @Failure      404         {object}  response.ErrorResponse
// @Failure      422         {object}  response.ErrorResponse
// @Router       /api/activities/running/{activityId}/gps [post]
// @Security     BearerAuth
func (ctrl *ActivityController) SendGPSPoints(c echo.Context) error {
	uid := c.Get("uid").(string)
	activityId := c.Param("activityId")

	req := new(requests.SendGPSPointsRequest)
	if err := c.Bind(req); err != nil {
		return c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error:   "invalid_request",
			Message: "リクエストの形式が不正です",
		})
	}

	// TODO: ビジネスロジック実装
	_ = uid
	_ = activityId

	return c.JSON(http.StatusOK, response.SendGPSPointsResponse{})
}

// GetRunningActivity ランニング記録詳細
// @Summary      ランニング記録詳細
// @Description  指定したランニングアクティビティの詳細情報（GPSポイント含む）を取得する
// @Tags         activities-running
// @Produce      json
// @Param        activityId  path      string  true  "アクティビティID"
// @Success      200         {object}  response.ActivityResponse
// @Failure      403         {object}  response.ErrorResponse
// @Failure      404         {object}  response.ErrorResponse
// @Router       /api/activities/running/{activityId} [get]
// @Security     BearerAuth
func (ctrl *ActivityController) GetRunningActivity(c echo.Context) error {
	uid := c.Get("uid").(string)
	activityId := c.Param("activityId")

	// TODO: ビジネスロジック実装
	_ = uid
	_ = activityId

	return c.JSON(http.StatusOK, response.ActivityResponse{})
}

// GetMyActivities 自分のアクティビティ一覧
// @Summary      自分のアクティビティ一覧
// @Description  自分のアクティビティ一覧を取得する
// @Tags         activities
// @Produce      json
// @Success      200  {array}   response.ActivityResponse
// @Failure      404  {object}  response.ErrorResponse
// @Router       /api/activities [get]
// @Security     BearerAuth
func (ctrl *ActivityController) GetMyActivities(c echo.Context) error {
	uid := c.Get("uid").(string)

	// TODO: ビジネスロジック実装
	_ = uid

	return c.JSON(http.StatusOK, []response.ActivityResponse{})
}

// GetTeamActivities チーム全体のアクティビティ一覧
// @Summary      チーム全体のアクティビティ一覧
// @Description  チーム全体のアクティビティ一覧を取得する
// @Tags         activities
// @Produce      json
// @Param        teamId  path      string  true  "チームID"
// @Success      200     {array}   response.ActivityResponse
// @Failure      403     {object}  response.ErrorResponse
// @Failure      404     {object}  response.ErrorResponse
// @Router       /api/teams/{teamId}/activities [get]
// @Security     BearerAuth
func (ctrl *ActivityController) GetTeamActivities(c echo.Context) error {
	uid := c.Get("uid").(string)
	teamId := c.Param("teamId")

	// TODO: ビジネスロジック実装
	_ = uid
	_ = teamId

	return c.JSON(http.StatusOK, []response.ActivityResponse{})
}
