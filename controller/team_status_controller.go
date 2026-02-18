package controller

import (
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/trihackathon/api/models"
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

	// 目標取得
	var goal models.Goal
	ctrl.db.First(&goal, "team_id = ?", teamId)

	// メンバー一覧取得
	var members []models.TeamMember
	ctrl.db.Preload("User").Where("team_id = ?", teamId).Find(&members)

	// HP履歴: 週次評価から構築
	var evaluations []models.WeeklyEvaluation
	ctrl.db.Preload("User").Where("team_id = ?", teamId).Order("week_number ASC").Find(&evaluations)

	hpHistory := buildHPHistory(evaluations, team.MaxHP)

	// メンバー進捗: 今週のアクティビティ集計
	membersProgress := ctrl.buildMembersProgress(teamId, members, team, goal)

	startedAt := ""
	if team.StartedAt != nil {
		startedAt = team.StartedAt.Format(time.RFC3339)
	}

	return c.JSON(http.StatusOK, response.TeamStatusResponse{
		TeamID:          team.ID,
		Status:          team.Status,
		CurrentHP:       team.CurrentHP,
		MaxHP:           team.MaxHP,
		CurrentWeek:     team.CurrentWeek,
		StartedAt:       startedAt,
		HPHistory:       hpHistory,
		MembersProgress: membersProgress,
	})
}

func buildHPHistory(evaluations []models.WeeklyEvaluation, maxHP int) []response.WeekHPHistory {
	if len(evaluations) == 0 {
		return []response.WeekHPHistory{}
	}

	// 週ごとにグループ化
	weekMap := make(map[int][]models.WeeklyEvaluation)
	for _, e := range evaluations {
		weekMap[e.WeekNumber] = append(weekMap[e.WeekNumber], e)
	}

	var history []response.WeekHPHistory
	hpStart := maxHP
	for week := 1; week <= len(weekMap); week++ {
		evals, ok := weekMap[week]
		if !ok {
			continue
		}
		totalChange := 0
		var changes []response.HPChangeEntry
		for _, e := range evals {
			totalChange += e.HPChange
			changes = append(changes, response.HPChangeEntry{
				UserID:    e.UserID,
				UserName:  e.User.Name,
				HPChange:  e.HPChange,
				TargetMet: e.TargetMet,
			})
		}
		hpEnd := hpStart + totalChange
		if hpEnd < 0 {
			hpEnd = 0
		}
		history = append(history, response.WeekHPHistory{
			Week:    week,
			HPStart: hpStart,
			HPEnd:   hpEnd,
			Changes: changes,
		})
		hpStart = hpEnd
	}

	return history
}

func (ctrl *TeamStatusController) buildMembersProgress(teamId string, members []models.TeamMember, team models.Team, goal models.Goal) []response.MemberProgress {
	progress := make([]response.MemberProgress, 0, len(members))

	if team.StartedAt == nil || team.CurrentWeek == 0 {
		for _, m := range members {
			progress = append(progress, response.MemberProgress{
				UserID:   m.UserID,
				UserName: m.User.Name,
			})
		}
		return progress
	}

	// 今週の期間を算出
	weekStart := team.StartedAt.AddDate(0, 0, (team.CurrentWeek-1)*7)
	weekEnd := weekStart.AddDate(0, 0, 7)

	for _, m := range members {
		// 今週のアクティビティ集計（gymはqualifiedVisitsが必要なため別途取得）
		var activities []models.Activity
		ctrl.db.Where("user_id = ? AND team_id = ? AND status = ? AND started_at >= ? AND started_at < ? AND (review_status IS NULL OR review_status != ?)",
			m.UserID, teamId, "completed", weekStart, weekEnd, "rejected").Find(&activities)

		var totalDist float64
		var totalVisits int
		var totalDuration int
		var qualifiedVisits int
		for _, a := range activities {
			totalDist += a.DistanceKM
			totalDuration += a.DurationMin
			if a.ExerciseType == "gym" {
				totalVisits++
				if goal.TargetMinDurationMin != nil {
					if a.DurationMin >= *goal.TargetMinDurationMin {
						qualifiedVisits++
					}
				} else {
					qualifiedVisits++
				}
			}
		}

		multiplier := m.TargetMultiplier
		if multiplier <= 0 {
			multiplier = 1.0
		}

		var progressPercent float64
		var distPtr *float64
		var visitsPtr *int
		var durationPtr *int

		switch team.ExerciseType {
		case "running":
			distPtr = &totalDist
			if goal.TargetDistanceKM != nil && *goal.TargetDistanceKM > 0 {
				effectiveTarget := *goal.TargetDistanceKM * multiplier
				progressPercent = (totalDist / effectiveTarget) * 100
			}
		case "gym":
			visitsPtr = &totalVisits
			durationPtr = &totalDuration
			if goal.TargetVisitsPerWeek != nil && *goal.TargetVisitsPerWeek > 0 {
				effectiveTarget := float64(*goal.TargetVisitsPerWeek) * multiplier
				visitCount := totalVisits
				if goal.TargetMinDurationMin != nil {
					visitCount = qualifiedVisits
				}
				progressPercent = (float64(visitCount) / effectiveTarget) * 100
			}
		}
		if progressPercent > 100 {
			progressPercent = 100
		}

		progress = append(progress, response.MemberProgress{
			UserID:                 m.UserID,
			UserName:               m.User.Name,
			CurrentWeekDistanceKM:  distPtr,
			CurrentWeekVisits:      visitsPtr,
			CurrentWeekDurationMin: durationPtr,
			TargetProgressPercent:  progressPercent,
		})
	}

	return progress
}
