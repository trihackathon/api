package controller

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/trihackathon/api/response"
	"gorm.io/gorm"
)

type TeamStatusController struct {
	db *gorm.DB
}

func NewTeamStatusController(db *gorm.DB) *TeamStatusController {
	return &TeamStatusController{db: db}
}

// GetTeamStatus チームHP・状態取得
// @Summary      チームHP・状態取得
// @Description  チームのHP、ステータス、HP履歴、メンバーの進捗状況を取得する
// @Tags         team-status
// @Produce      json
// @Param        teamId  path      string  true  "チームID"
// @Success      200     {object}  response.TeamStatusResponse
// @Failure      403     {object}  response.ErrorResponse
// @Failure      404     {object}  response.ErrorResponse
// @Router       /api/teams/{teamId}/status [get]
// @Security     BearerAuth
func (ctrl *TeamStatusController) GetTeamStatus(c echo.Context) error {
	uid := c.Get("uid").(string)
	teamId := c.Param("teamId")

	// TODO: ビジネスロジック実装
	_ = uid
	_ = teamId

	return c.JSON(http.StatusOK, response.TeamStatusResponse{})
}
