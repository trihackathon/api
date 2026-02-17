package controller

import (
	"fmt"
	"log"
	"net/http"
	"strings"
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

	// ユーザーが users テーブルに存在するか確認（TeamMember の外部キー制約のため必須）
	var user models.User
	if err := ctrl.db.First(&user, "id = ?", uid).Error; err != nil {
		return c.JSON(http.StatusUnprocessableEntity, response.ErrorResponse{
			Error:   "user_not_registered",
			Message: "プロフィールが未登録です。先にアカウント設定（新規登録）を完了してください",
		})
	}

	// メンバー追加とチームステータス更新をトランザクションで実行
	var team models.Team
	var members []models.TeamMember
	var teamReady bool

	err = ctrl.db.Transaction(func(tx *gorm.DB) error {
		// メンバー追加
		memberID := utils.GenerateULID()
		member := models.TeamMember{
			ID:     memberID,
			TeamID: inviteCode.TeamID,
			UserID: uid,
			Role:   "member",
		}
		if err := tx.Create(&member).Error; err != nil {
			return err
		}

		// メンバー数を確認
		var memberCount int64
		if err := tx.Model(&models.TeamMember{}).Where("team_id = ?", inviteCode.TeamID).Count(&memberCount).Error; err != nil {
			return err
		}

		// 3人揃ったらステータスをactiveに更新（started_at, current_weekも設定）
		if memberCount >= 3 {
			now := time.Now()
			if err := tx.Model(&models.Team{}).Where("id = ?", inviteCode.TeamID).Updates(map[string]interface{}{
				"status":       "active",
				"started_at":   now,
				"current_week": 1,
			}).Error; err != nil {
				return err
			}
			teamReady = true
		}

		return nil
	})

	if err != nil {
		log.Printf("[JoinTeam] transaction error: %v", err)
		// ユーザーが存在しない場合のエラーメッセージを改善
		errStr := err.Error()
		if strings.Contains(errStr, "foreign key") || strings.Contains(errStr, "violates foreign key") {
			return c.JSON(http.StatusUnprocessableEntity, response.ErrorResponse{
				Error:   "user_not_registered",
				Message: "プロフィールが未登録です。先にアカウント設定を完了してください",
			})
		}
		if strings.Contains(errStr, "unique") || strings.Contains(errStr, "duplicate") {
			return c.JSON(http.StatusConflict, response.ErrorResponse{
				Error:   "already_in_team",
				Message: "既にこのチームに参加しています",
			})
		}
		return c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Error:   "join_failed",
			Message: fmt.Sprintf("チームへの参加に失敗しました: %v", err),
		})
	}

	// レスポンス構築
	ctrl.db.First(&team, "id = ?", inviteCode.TeamID)
	ctrl.db.Preload("User").Where("team_id = ?", team.ID).Find(&members)

	return c.JSON(http.StatusOK, response.JoinTeamResponse{
		Team:      response.NewTeamResponse(team, members),
		TeamReady: teamReady,
	})
}
