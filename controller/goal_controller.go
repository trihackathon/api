package controller

import (
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/oklog/ulid/v2"
	"github.com/trihackathon/api/models"
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

	// チーム取得
	var team models.Team
	if err := ctrl.db.First(&team, "id = ?", teamId).Error; err != nil {
		return c.JSON(http.StatusNotFound, response.ErrorResponse{
			Error:   "team_not_found",
			Message: "チームが見つかりません",
		})
	}

	// リーダー確認
	var member models.TeamMember
	if err := ctrl.db.First(&member, "team_id = ? AND user_id = ?", teamId, uid).Error; err != nil {
		return c.JSON(http.StatusForbidden, response.ErrorResponse{
			Error:   "not_team_member",
			Message: "チームのメンバーではありません",
		})
	}
	if member.Role != "leader" {
		return c.JSON(http.StatusForbidden, response.ErrorResponse{
			Error:   "not_leader",
			Message: "リーダーのみ目標を設定できます",
		})
	}

	// メンバー3人揃っているか確認
	var memberCount int64
	ctrl.db.Model(&models.TeamMember{}).Where("team_id = ?", teamId).Count(&memberCount)
	if memberCount < 3 {
		return c.JSON(http.StatusUnprocessableEntity, response.ErrorResponse{
			Error:   "team_not_ready",
			Message: "チームメンバーが3人揃っていません",
		})
	}

	// 既に目標が存在するか確認
	var existingGoal models.Goal
	if err := ctrl.db.First(&existingGoal, "team_id = ?", teamId).Error; err == nil {
		return c.JSON(http.StatusConflict, response.ErrorResponse{
			Error:   "goal_already_exists",
			Message: "目標は既に設定されています。更新はPUTを使用してください",
		})
	}

	goal := models.Goal{
		ID:                   ulid.Make().String(),
		TeamID:               teamId,
		ExerciseType:         team.ExerciseType,
		TargetDistanceKM:     req.TargetDistanceKM,
		TargetVisitsPerWeek:  req.TargetVisitsPerWeek,
		TargetMinDurationMin: req.TargetMinDurationMin,
	}

	// トランザクションで目標作成 + チームステータス更新
	if err := ctrl.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&goal).Error; err != nil {
			return err
		}
		now := time.Now()
		if err := tx.Model(&team).Updates(map[string]interface{}{
			"status":       "active",
			"started_at":   now,
			"current_week": 1,
		}).Error; err != nil {
			return err
		}
		return nil
	}); err != nil {
		return c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Error:   "create_failed",
			Message: "目標の設定に失敗しました",
		})
	}

	return c.JSON(http.StatusCreated, newGoalResponse(goal))
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

	// メンバー確認
	var member models.TeamMember
	if err := ctrl.db.First(&member, "team_id = ? AND user_id = ?", teamId, uid).Error; err != nil {
		return c.JSON(http.StatusForbidden, response.ErrorResponse{
			Error:   "not_team_member",
			Message: "チームのメンバーではありません",
		})
	}

	var goal models.Goal
	if err := ctrl.db.First(&goal, "team_id = ?", teamId).Error; err != nil {
		return c.JSON(http.StatusNotFound, response.ErrorResponse{
			Error:   "goal_not_found",
			Message: "目標が設定されていません",
		})
	}

	return c.JSON(http.StatusOK, newGoalResponse(goal))
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

	// リーダー確認
	var member models.TeamMember
	if err := ctrl.db.First(&member, "team_id = ? AND user_id = ?", teamId, uid).Error; err != nil {
		return c.JSON(http.StatusForbidden, response.ErrorResponse{
			Error:   "not_team_member",
			Message: "チームのメンバーではありません",
		})
	}
	if member.Role != "leader" {
		return c.JSON(http.StatusForbidden, response.ErrorResponse{
			Error:   "not_leader",
			Message: "リーダーのみ目標を更新できます",
		})
	}

	var goal models.Goal
	if err := ctrl.db.First(&goal, "team_id = ?", teamId).Error; err != nil {
		return c.JSON(http.StatusNotFound, response.ErrorResponse{
			Error:   "goal_not_found",
			Message: "目標が設定されていません",
		})
	}

	goal.TargetDistanceKM = req.TargetDistanceKM
	goal.TargetVisitsPerWeek = req.TargetVisitsPerWeek
	goal.TargetMinDurationMin = req.TargetMinDurationMin

	if err := ctrl.db.Save(&goal).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Error:   "update_failed",
			Message: "目標の更新に失敗しました",
		})
	}

	return c.JSON(http.StatusOK, newGoalResponse(goal))
}

func newGoalResponse(goal models.Goal) response.GoalResponse {
	return response.GoalResponse{
		ID:                   goal.ID,
		TeamID:               goal.TeamID,
		ExerciseType:         goal.ExerciseType,
		TargetDistanceKM:     goal.TargetDistanceKM,
		TargetVisitsPerWeek:  goal.TargetVisitsPerWeek,
		TargetMinDurationMin: goal.TargetMinDurationMin,
		CreatedAt:            goal.CreatedAt.Format(time.RFC3339),
		UpdatedAt:            goal.UpdatedAt.Format(time.RFC3339),
	}
}
