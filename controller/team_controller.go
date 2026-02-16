package controller

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/trihackathon/api/models"
	"github.com/trihackathon/api/requests"
	"github.com/trihackathon/api/response"
	"github.com/trihackathon/api/utils"
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

	// バリデーション
	if req.Name == "" {
		return c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error:   "invalid_request",
			Message: "name は必須です",
		})
	}
	if req.ExerciseType != "running" && req.ExerciseType != "gym" {
		return c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error:   "invalid_request",
			Message: "exercise_type は running または gym を指定してください",
		})
	}
	if req.Strictness == "" {
		req.Strictness = "normal"
	}

	// ユーザーが既にアクティブチームに所属しているか確認
	var existingMember models.TeamMember
	err := ctrl.db.
		Joins("JOIN teams ON teams.id = team_members.team_id").
		Where("team_members.user_id = ? AND teams.status IN ?", uid, []string{"forming", "active"}).
		First(&existingMember).Error
	if err == nil {
		return c.JSON(http.StatusConflict, response.ErrorResponse{
			Error:   "already_in_team",
			Message: "既にアクティブなチームに所属しています",
		})
	}

	teamID := utils.GenerateULID()
	memberID := utils.GenerateULID()

	team := models.Team{
		ID:           teamID,
		Name:         req.Name,
		ExerciseType: req.ExerciseType,
		Strictness:   req.Strictness,
		Status:       "forming",
		MaxHP:        100,
		CurrentHP:    100,
		CurrentWeek:  0,
	}

	member := models.TeamMember{
		ID:     memberID,
		TeamID: teamID,
		UserID: uid,
		Role:   "leader",
	}

	// トランザクションでチームとメンバーを作成
	if err := ctrl.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&team).Error; err != nil {
			return err
		}
		if err := tx.Create(&member).Error; err != nil {
			return err
		}
		return nil
	}); err != nil {
		return c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Error:   "create_failed",
			Message: "チームの作成に失敗しました",
		})
	}

	// レスポンス用にメンバー情報を取得
	var members []models.TeamMember
	ctrl.db.Preload("User").Where("team_id = ?", teamID).Find(&members)

	return c.JSON(http.StatusCreated, response.NewTeamResponse(team, members))
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

	// 自分が所属するアクティブチームを検索
	var member models.TeamMember
	err := ctrl.db.
		Joins("JOIN teams ON teams.id = team_members.team_id").
		Where("team_members.user_id = ? AND teams.status IN ?", uid, []string{"forming", "active"}).
		First(&member).Error
	if err != nil {
		return c.JSON(http.StatusNotFound, response.ErrorResponse{
			Error:   "team_not_found",
			Message: "所属するアクティブなチームが見つかりません",
		})
	}

	var team models.Team
	ctrl.db.First(&team, "id = ?", member.TeamID)

	var members []models.TeamMember
	ctrl.db.Preload("User").Where("team_id = ?", team.ID).Find(&members)

	return c.JSON(http.StatusOK, response.NewTeamResponse(team, members))
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

	var team models.Team
	if err := ctrl.db.First(&team, "id = ?", teamId).Error; err != nil {
		return c.JSON(http.StatusNotFound, response.ErrorResponse{
			Error:   "team_not_found",
			Message: "チームが見つかりません",
		})
	}

	// メンバーか確認
	var member models.TeamMember
	if err := ctrl.db.Where("team_id = ? AND user_id = ?", teamId, uid).First(&member).Error; err != nil {
		return c.JSON(http.StatusForbidden, response.ErrorResponse{
			Error:   "not_team_member",
			Message: "このチームのメンバーではありません",
		})
	}

	var members []models.TeamMember
	ctrl.db.Preload("User").Where("team_id = ?", teamId).Find(&members)

	return c.JSON(http.StatusOK, response.NewTeamResponse(team, members))
}
