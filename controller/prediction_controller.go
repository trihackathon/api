package controller

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/trihackathon/api/response"
	"gorm.io/gorm"
)

type PredictionController struct {
	db *gorm.DB
}

func NewPredictionController(db *gorm.DB) *PredictionController {
	return &PredictionController{db: db}
}

// GetMyPrediction 自分の失敗予測
// @Summary      自分の失敗予測
// @Description  過去のアクティビティデータから曜日別の成功率を算出し、危険な曜日を警告する。成功率40%未満の曜日を「危険」と判定。
// @Tags         predictions
// @Produce      json
// @Success      200  {object}  response.PredictionResponse
// @Failure      404  {object}  response.ErrorResponse
// @Failure      422  {object}  response.ErrorResponse
// @Router       /api/predictions/me [get]
// @Security     BearerAuth
func (ctrl *PredictionController) GetMyPrediction(c echo.Context) error {
	uid := c.Get("uid").(string)

	// TODO: ビジネスロジック実装
	_ = uid

	return c.JSON(http.StatusOK, response.PredictionResponse{})
}
