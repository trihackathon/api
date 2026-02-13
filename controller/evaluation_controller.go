package controller

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/trihackathon/api/response"
	"gorm.io/gorm"
)

type EvaluationController struct {
	db *gorm.DB
}

func NewEvaluationController(db *gorm.DB) *EvaluationController {
	return &EvaluationController{db: db}
}

// GetEvaluations 週次評価一覧
// @Summary      週次評価一覧
// @Description  チームの週次評価一覧を取得する。weekクエリパラメータで特定の週のみ取得可能。
// @Tags         evaluations
// @Produce      json
// @Param        teamId  path      string  true   "チームID"
// @Param        week    query     int     false  "特定の週番号"
// @Success      200     {array}   response.WeeklyEvaluationResponse
// @Failure      403     {object}  response.ErrorResponse
// @Failure      404     {object}  response.ErrorResponse
// @Router       /api/teams/{teamId}/evaluations [get]
// @Security     BearerAuth
func (ctrl *EvaluationController) GetEvaluations(c echo.Context) error {
	uid := c.Get("uid").(string)
	teamId := c.Param("teamId")

	// TODO: ビジネスロジック実装
	_ = uid
	_ = teamId

	return c.JSON(http.StatusOK, []response.WeeklyEvaluationResponse{})
}

// GetCurrentWeekEvaluation 今週の進捗状況
// @Summary      今週の進捗状況
// @Description  現在の週のリアルタイム進捗を返す（まだ週次評価が確定していない状態）
// @Tags         evaluations
// @Produce      json
// @Param        teamId  path      string  true  "チームID"
// @Success      200     {object}  response.CurrentWeekEvaluationResponse
// @Failure      403     {object}  response.ErrorResponse
// @Failure      404     {object}  response.ErrorResponse
// @Router       /api/teams/{teamId}/evaluations/current [get]
// @Security     BearerAuth
func (ctrl *EvaluationController) GetCurrentWeekEvaluation(c echo.Context) error {
	uid := c.Get("uid").(string)
	teamId := c.Param("teamId")

	// TODO: ビジネスロジック実装
	_ = uid
	_ = teamId

	return c.JSON(http.StatusOK, response.CurrentWeekEvaluationResponse{})
}
