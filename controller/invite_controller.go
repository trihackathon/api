package controller

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/trihackathon/api/requests"
	"github.com/trihackathon/api/response"
	"gorm.io/gorm"
)

type InviteController struct {
	db *gorm.DB
}

func NewInviteController(db *gorm.DB) *InviteController {
	return &InviteController{db: db}
}

// CreateInviteCode 招待コード生成
// @Summary      招待コード生成
// @Description  6桁の英数大文字の招待コードを生成する。有効期限は24時間。チーム状態がformingかつメンバー3人未満の場合のみ発行可能。
// @Tags         invite
// @Produce      json
// @Param        teamId  path      string  true  "チームID"
// @Success      201     {object}  response.InviteCodeResponse
// @Failure      403     {object}  response.ErrorResponse
// @Failure      422     {object}  response.ErrorResponse
// @Router       /api/teams/{teamId}/invite [post]
// @Security     BearerAuth
func (ctrl *InviteController) CreateInviteCode(c echo.Context) error {
	uid := c.Get("uid").(string)
	teamId := c.Param("teamId")

	// TODO: ビジネスロジック実装
	_ = uid
	_ = teamId

	return c.JSON(http.StatusCreated, response.InviteCodeResponse{})
}

// JoinTeam 招待コードでチーム参加
// @Summary      招待コードでチーム参加
// @Description  招待コードを使用してチームに参加する。3人揃った場合team_readyがtrueになる。
// @Tags         invite
// @Accept       json
// @Produce      json
// @Param        body  body      requests.JoinTeamRequest  true  "招待コード"
// @Success      200   {object}  response.JoinTeamResponse
// @Failure      400   {object}  response.ErrorResponse
// @Failure      404   {object}  response.ErrorResponse
// @Failure      409   {object}  response.ErrorResponse
// @Failure      410   {object}  response.ErrorResponse
// @Failure      422   {object}  response.ErrorResponse
// @Router       /api/teams/join [post]
// @Security     BearerAuth
func (ctrl *InviteController) JoinTeam(c echo.Context) error {
	uid := c.Get("uid").(string)

	req := new(requests.JoinTeamRequest)
	if err := c.Bind(req); err != nil {
		return c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error:   "invalid_request",
			Message: "リクエストの形式が不正です",
		})
	}

	// TODO: ビジネスロジック実装
	_ = uid

	return c.JSON(http.StatusOK, response.JoinTeamResponse{})
}
