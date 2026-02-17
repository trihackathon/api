package controller

import (
	"net/http"
	"os"

	"github.com/labstack/echo/v4"
	"github.com/trihackathon/api/response"
	"github.com/trihackathon/api/service"
)

type CronController struct {
	evaluationService *service.EvaluationService
}

func NewCronController(evaluationService *service.EvaluationService) *CronController {
	return &CronController{evaluationService: evaluationService}
}

// RunWeeklyEvaluation 週次評価実行
// @Summary      週次評価実行
// @Description  全activeチームの週次評価を実行し、HP更新・disbanded処理を行う
// @Tags         cron
// @Produce      json
// @Param        X-Cron-Secret  header  string  true  "Cronシークレットキー"
// @Success      200  {object}  map[string]interface{}
// @Failure      401  {object}  response.ErrorResponse
// @Failure      500  {object}  response.ErrorResponse
// @Router       /cron/weekly-evaluation [post]
func (ctrl *CronController) RunWeeklyEvaluation(c echo.Context) error {
	// APIキー認証
	secret := c.Request().Header.Get("X-Cron-Secret")
	expectedSecret := os.Getenv("CRON_SECRET")
	if expectedSecret == "" || secret != expectedSecret {
		return c.JSON(http.StatusUnauthorized, response.ErrorResponse{
			Error:   "unauthorized",
			Message: "Invalid or missing cron secret",
		})
	}

	result, err := ctrl.evaluationService.RunWeeklyEvaluation()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Error:   "evaluation_failed",
			Message: err.Error(),
		})
	}

	return c.JSON(http.StatusOK, result)
}
