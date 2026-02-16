package controller

import (
	"math"
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/trihackathon/api/models"
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

	// メンバー確認
	var member models.TeamMember
	if err := ctrl.db.First(&member, "team_id = ? AND user_id = ?", teamId, uid).Error; err != nil {
		return c.JSON(http.StatusForbidden, response.ErrorResponse{
			Error:   "not_team_member",
			Message: "チームのメンバーではありません",
		})
	}

	query := ctrl.db.Preload("User").Where("team_id = ?", teamId)

	// weekクエリパラメータで絞り込み
	if weekStr := c.QueryParam("week"); weekStr != "" {
		week, err := strconv.Atoi(weekStr)
		if err == nil {
			query = query.Where("week_number = ?", week)
		}
	}

	var evaluations []models.WeeklyEvaluation
	query.Order("week_number ASC, user_id ASC").Find(&evaluations)

	results := make([]response.WeeklyEvaluationResponse, len(evaluations))
	for i, e := range evaluations {
		results[i] = response.WeeklyEvaluationResponse{
			ID:               e.ID,
			TeamID:           e.TeamID,
			UserID:           e.UserID,
			UserName:         e.User.Name,
			WeekNumber:       e.WeekNumber,
			TargetMet:        e.TargetMet,
			TotalDistanceKM:  e.TotalDistanceKM,
			TotalVisits:      e.TotalVisits,
			TotalDurationMin: e.TotalDurationMin,
			HPChange:         e.HPChange,
			EvaluatedAt:      e.EvaluatedAt.Format(time.RFC3339),
		}
	}

	return c.JSON(http.StatusOK, results)
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

	// チーム取得
	var team models.Team
	if err := ctrl.db.First(&team, "id = ?", teamId).Error; err != nil {
		return c.JSON(http.StatusNotFound, response.ErrorResponse{
			Error:   "team_not_found",
			Message: "チームが見つかりません",
		})
	}

	// メンバー確認
	var member models.TeamMember
	if err := ctrl.db.First(&member, "team_id = ? AND user_id = ?", teamId, uid).Error; err != nil {
		return c.JSON(http.StatusForbidden, response.ErrorResponse{
			Error:   "not_team_member",
			Message: "チームのメンバーではありません",
		})
	}

	if team.StartedAt == nil || team.CurrentWeek == 0 {
		return c.JSON(http.StatusOK, response.CurrentWeekEvaluationResponse{
			TeamID:     teamId,
			WeekNumber: 0,
			Members:    []response.CurrentWeekMemberProgress{},
		})
	}

	// 目標取得
	var goal models.Goal
	ctrl.db.First(&goal, "team_id = ?", teamId)

	// 今週の期間
	weekStart := team.StartedAt.AddDate(0, 0, (team.CurrentWeek-1)*7)
	weekEnd := weekStart.AddDate(0, 0, 7)
	now := time.Now()
	daysRemaining := int(math.Ceil(weekEnd.Sub(now).Hours() / 24))
	if daysRemaining < 0 {
		daysRemaining = 0
	}

	// メンバー一覧
	var members []models.TeamMember
	ctrl.db.Preload("User").Where("team_id = ?", teamId).Find(&members)

	var memberProgresses []response.CurrentWeekMemberProgress
	for _, m := range members {
		// 今週のアクティビティ取得
		var activities []models.Activity
		ctrl.db.Where("user_id = ? AND team_id = ? AND status = ? AND started_at >= ? AND started_at < ?",
			m.UserID, teamId, "completed", weekStart, weekEnd).Order("started_at ASC").Find(&activities)

		var totalDist float64
		var totalVisits int
		var totalDuration int
		var actSummaries []response.WeekActivitySummary

		for _, a := range activities {
			totalDist += a.DistanceKM
			totalDuration += a.DurationMin
			if a.ExerciseType == "gym" {
				totalVisits++
			}
			actSummaries = append(actSummaries, response.WeekActivitySummary{
				ID:          a.ID,
				Date:        a.StartedAt.Format("2006-01-02"),
				DistanceKM:  a.DistanceKM,
				DurationMin: a.DurationMin,
			})
		}
		if actSummaries == nil {
			actSummaries = []response.WeekActivitySummary{}
		}

		var progressPercent float64
		switch team.ExerciseType {
		case "running":
			if goal.TargetDistanceKM != nil && *goal.TargetDistanceKM > 0 {
				progressPercent = (totalDist / *goal.TargetDistanceKM) * 100
			}
		case "gym":
			if goal.TargetVisitsPerWeek != nil && *goal.TargetVisitsPerWeek > 0 {
				progressPercent = (float64(totalVisits) / float64(*goal.TargetVisitsPerWeek)) * 100
			}
		}
		if progressPercent > 100 {
			progressPercent = 100
		}

		onTrack := progressPercent >= 100 || (daysRemaining > 0 && progressPercent > 0)

		memberProgresses = append(memberProgresses, response.CurrentWeekMemberProgress{
			UserID:                m.UserID,
			UserName:              m.User.Name,
			TotalDistanceKM:       totalDist,
			TotalVisits:           totalVisits,
			TotalDurationMin:      totalDuration,
			TargetProgressPercent: progressPercent,
			OnTrack:               onTrack,
			ActivitiesThisWeek:    actSummaries,
		})
	}

	if memberProgresses == nil {
		memberProgresses = []response.CurrentWeekMemberProgress{}
	}

	return c.JSON(http.StatusOK, response.CurrentWeekEvaluationResponse{
		TeamID:        teamId,
		WeekNumber:    team.CurrentWeek,
		WeekStart:     weekStart.Format(time.RFC3339),
		WeekEnd:       weekEnd.Add(-time.Second).Format(time.RFC3339),
		DaysRemaining: daysRemaining,
		Members:       memberProgresses,
	})
}
