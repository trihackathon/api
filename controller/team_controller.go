package controller

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/trihackathon/api/requests"
	"github.com/trihackathon/api/response"
	"gorm.io/gorm"
)

type TeamController struct {
	db *gorm.DB
}

func NewTeamController(db *gorm.DB) *TeamController {
	return &TeamController{db: db}
}

// CreateTeam チーム作成
// @Summary      チーム作成
// @Description  チームを作成し、作成者をリーダーとしてメンバーに追加する。1ユーザーが同時に参加できるアクティブチームは1つのみ。
// @Tags         teams
// @Accept       json
// @Produce      json
// @Param        body  body      requests.CreateTeamRequest  true  "チーム情報"
// @Success      201   {object}  response.TeamResponse
// @Failure      400   {object}  response.ErrorResponse
// @Failure      409   {object}  response.ErrorResponse
// @Router       /api/teams [post]
// @Security     BearerAuth
func (ctrl *TeamController) CreateTeam(c echo.Context) error {
	uid := c.Get("uid").(string)

	req := new(requests.CreateTeamRequest)
	if err := c.Bind(req); err != nil {
		return c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error:   "invalid_request",
			Message: "リクエストの形式が不正です",
		})
	}

	// TODO: ビジネスロジック実装
	_ = uid

	return c.JSON(http.StatusCreated, response.TeamResponse{})
}

// GetMyTeam 自分のチーム取得
// @Summary      自分のチーム取得
// @Description  自分が所属するアクティブなチームを返す
// @Tags         teams
// @Produce      json
// @Success      200  {object}  response.TeamResponse
// @Failure      404  {object}  response.ErrorResponse
// @Router       /api/teams/me [get]
// @Security     BearerAuth
func (ctrl *TeamController) GetMyTeam(c echo.Context) error {
	uid := c.Get("uid").(string)

	// TODO: ビジネスロジック実装
	_ = uid

	return c.JSON(http.StatusOK, response.TeamResponse{})
}

// GetTeam チーム詳細取得
// @Summary      チーム詳細取得
// @Description  指定したチームの詳細情報を取得する。チームメンバーのみアクセス可能。
// @Tags         teams
// @Produce      json
// @Param        teamId  path      string  true  "チームID"
// @Success      200     {object}  response.TeamResponse
// @Failure      403     {object}  response.ErrorResponse
// @Failure      404     {object}  response.ErrorResponse
// @Router       /api/teams/{teamId} [get]
// @Security     BearerAuth
func (ctrl *TeamController) GetTeam(c echo.Context) error {
	uid := c.Get("uid").(string)
	teamId := c.Param("teamId")

	// TODO: ビジネスロジック実装
	_ = uid
	_ = teamId

	return c.JSON(http.StatusOK, response.TeamResponse{})
}
