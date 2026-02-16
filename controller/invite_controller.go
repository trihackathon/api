package controller

import (
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/trihackathon/api/models"
	"github.com/trihackathon/api/requests"
	"github.com/trihackathon/api/response"
	"github.com/trihackathon/api/utils"
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

	// チーム存在確認
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

	// チーム状態がformingか確認
	if team.Status != "forming" {
		return c.JSON(http.StatusUnprocessableEntity, response.ErrorResponse{
			Error:   "team_not_forming",
			Message: "チームはメンバー募集中ではありません",
		})
	}

	// メンバー数確認
	var memberCount int64
	ctrl.db.Model(&models.TeamMember{}).Where("team_id = ?", teamId).Count(&memberCount)
	if memberCount >= 3 {
		return c.JSON(http.StatusUnprocessableEntity, response.ErrorResponse{
			Error:   "team_full",
			Message: "チームは満員です",
		})
	}

	code := utils.GenerateInviteCode()
	inviteCode := models.InviteCode{
		Code:      code,
		TeamID:    teamId,
		CreatedBy: uid,
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}

	if err := ctrl.db.Create(&inviteCode).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Error:   "create_failed",
			Message: "招待コードの生成に失敗しました",
		})
	}

	return c.JSON(http.StatusCreated, response.InviteCodeResponse{
		Code:               code,
		TeamID:             team.ID,
		TeamName:           team.Name,
		ExerciseType:       team.ExerciseType,
		ExpiresAt:          inviteCode.ExpiresAt.Format(time.RFC3339),
		CurrentMemberCount: int(memberCount),
	})
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

	// コード存在確認
	var inviteCode models.InviteCode
	if err := ctrl.db.First(&inviteCode, "code = ?", req.Code).Error; err != nil {
		return c.JSON(http.StatusNotFound, response.ErrorResponse{
			Error:   "code_not_found",
			Message: "招待コードが見つかりません",
		})
	}

	// 有効期限チェック
	if time.Now().After(inviteCode.ExpiresAt) {
		return c.JSON(http.StatusGone, response.ErrorResponse{
			Error:   "code_expired",
			Message: "招待コードの有効期限が切れています",
		})
	}

	// 既にアクティブチーム所属チェック
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

	// チーム満員チェック
	var memberCount int64
	ctrl.db.Model(&models.TeamMember{}).Where("team_id = ?", inviteCode.TeamID).Count(&memberCount)
	if memberCount >= 3 {
		return c.JSON(http.StatusUnprocessableEntity, response.ErrorResponse{
			Error:   "team_full",
			Message: "チームは満員です",
		})
	}

	// メンバー追加
	memberID := utils.GenerateULID()
	member := models.TeamMember{
		ID:     memberID,
		TeamID: inviteCode.TeamID,
		UserID: uid,
		Role:   "member",
	}
	if err := ctrl.db.Create(&member).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Error:   "join_failed",
			Message: "チームへの参加に失敗しました",
		})
	}

	// レスポンス構築
	var team models.Team
	ctrl.db.First(&team, "id = ?", inviteCode.TeamID)

	var members []models.TeamMember
	ctrl.db.Preload("User").Where("team_id = ?", team.ID).Find(&members)

	teamReady := len(members) >= 3

	return c.JSON(http.StatusOK, response.JoinTeamResponse{
		Team:      response.NewTeamResponse(team, members),
		TeamReady: teamReady,
	})
}
