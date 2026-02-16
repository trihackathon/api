package controller

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/trihackathon/api/requests"
	"github.com/trihackathon/api/response"
	"gorm.io/gorm"
)

type GoalController struct {
	db *gorm.DB
}

func NewGoalController(db *gorm.DB) *GoalController {
	return &GoalController{db: db}
}

// CreateGoal 目標設定
// @Summary      目標設定
// @Description  チームの目標を設定する。リーダーのみ設定可能。チームメンバーが3人揃った状態でのみ設定可能。目標設定完了でチームstatusをactiveに変更。
// @Tags         goals
// @Accept       json
// @Produce      json
// @Param        teamId  path      string                    true  "チームID"
// @Param        body    body      requests.CreateGoalRequest  true  "目標情報"
// @Success      201     {object}  response.GoalResponse
// @Failure      400     {object}  response.ErrorResponse
// @Failure      403     {object}  response.ErrorResponse
// @Failure      409     {object}  response.ErrorResponse
// @Failure      422     {object}  response.ErrorResponse
// @Router       /api/teams/{teamId}/goal [post]
// @Security     BearerAuth
func (ctrl *GoalController) CreateGoal(c echo.Context) error {
	uid := c.Get("uid").(string)
	teamId := c.Param("teamId")

	req := new(requests.CreateGoalRequest)
	if err := c.Bind(req); err != nil {
		return c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error:   "invalid_request",
			Message: "リクエストの形式が不正です",
		})
	}

	// TODO: ビジネスロジック実装
	_ = uid
	_ = teamId

	return c.JSON(http.StatusCreated, response.GoalResponse{})
}

// GetGoal 目標取得
// @Summary      目標取得
// @Description  チームの目標を取得する
// @Tags         goals
// @Produce      json
// @Param        teamId  path      string  true  "チームID"
// @Success      200     {object}  response.GoalResponse
// @Failure      403     {object}  response.ErrorResponse
// @Failure      404     {object}  response.ErrorResponse
// @Router       /api/teams/{teamId}/goal [get]
// @Security     BearerAuth
func (ctrl *GoalController) GetGoal(c echo.Context) error {
	uid := c.Get("uid").(string)
	teamId := c.Param("teamId")

	// TODO: ビジネスロジック実装
	_ = uid
	_ = teamId

	return c.JSON(http.StatusOK, response.GoalResponse{})
}

// UpdateGoal 目標更新
// @Summary      目標更新
// @Description  チームの目標を更新する。リーダーのみ更新可能。
// @Tags         goals
// @Accept       json
// @Produce      json
// @Param        teamId  path      string                    true  "チームID"
// @Param        body    body      requests.CreateGoalRequest  true  "更新する目標情報"
// @Success      200     {object}  response.GoalResponse
// @Failure      403     {object}  response.ErrorResponse
// @Failure      404     {object}  response.ErrorResponse
// @Router       /api/teams/{teamId}/goal [put]
// @Security     BearerAuth
func (ctrl *GoalController) UpdateGoal(c echo.Context) error {
	uid := c.Get("uid").(string)
	teamId := c.Param("teamId")

	req := new(requests.CreateGoalRequest)
	if err := c.Bind(req); err != nil {
		return c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error:   "invalid_request",
			Message: "リクエストの形式が不正です",
		})
	}

	// TODO: ビジネスロジック実装
	_ = uid
	_ = teamId

	return c.JSON(http.StatusOK, response.GoalResponse{})
}
