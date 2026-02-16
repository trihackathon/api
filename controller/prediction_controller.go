package controller

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/trihackathon/api/models"
	"github.com/trihackathon/api/response"
	"gorm.io/gorm"
)

type PredictionController struct {
	db *gorm.DB
}

func NewPredictionController(db *gorm.DB) *PredictionController {
	return &PredictionController{db: db}
}

var dayNames = [7]string{"日曜日", "月曜日", "火曜日", "水曜日", "木曜日", "金曜日", "土曜日"}

// GetMyPrediction 自分の失敗予測
// @Summary      自分の失敗予測
// @Description  過去のアクティビティデータから曜日別の成功率を算出し、危険な曜日を警告する。成功率40%未満の曜日を「危険」と判定。
// @Tags         predictions
// @Produce      json
// @Success      200  {object}  response.PredictionResponse
// @Failure      404  {object}  response.ErrorResponse
// @Failure      422  {object}  response.ErrorResponse
// @Router       /api/predictions/me [get]
// @Security     BearerAuth
func (ctrl *PredictionController) GetMyPrediction(c echo.Context) error {
	uid := c.Get("uid").(string)

	// ユーザーがチームに所属しているか確認
	var teamMember models.TeamMember
	if err := ctrl.db.Preload("Team").First(&teamMember, "user_id = ?", uid).Error; err != nil {
		return c.JSON(http.StatusNotFound, response.ErrorResponse{
			Error:   "no_team",
			Message: "チームに所属していません",
		})
	}

	team := teamMember.Team
	if team.StartedAt == nil {
		return c.JSON(http.StatusUnprocessableEntity, response.ErrorResponse{
			Error:   "team_not_active",
			Message: "チームがまだアクティブになっていません",
		})
	}

	// 過去4週間のアクティビティを取得
	analysisPeriodWeeks := 4
	since := time.Now().AddDate(0, 0, -analysisPeriodWeeks*7)

	var activities []models.Activity
	ctrl.db.Where("user_id = ? AND status = ? AND started_at >= ?",
		uid, "completed", since).Find(&activities)

	// 曜日ごとの集計
	// dayActivity[曜日] = アクティビティがあった日数
	// dayTotal[曜日] = その曜日の合計日数
	dayActivity := [7]int{}
	dayTotal := [7]int{}

	// 分析期間中の各日を走査
	now := time.Now()
	for d := since; d.Before(now); d = d.AddDate(0, 0, 1) {
		dow := int(d.Weekday())
		dayTotal[dow]++
	}

	// アクティビティがあった日をカウント（同じ日に複数あっても1回）
	activityDays := make(map[string]bool)
	for _, a := range activities {
		key := a.StartedAt.Format("2006-01-02")
		activityDays[key] = true
	}
	for dateStr := range activityDays {
		t, err := time.Parse("2006-01-02", dateStr)
		if err != nil {
			continue
		}
		dow := int(t.Weekday())
		dayActivity[dow]++
	}

	var dailyStats []response.DailyStat
	var dangerDays []string

	for dow := 0; dow < 7; dow++ {
		var successRate float64
		if dayTotal[dow] > 0 {
			successRate = float64(dayActivity[dow]) / float64(dayTotal[dow])
		}
		isDanger := successRate < 0.4
		if isDanger {
			dangerDays = append(dangerDays, dayNames[dow])
		}

		dailyStats = append(dailyStats, response.DailyStat{
			DayOfWeek:     dow,
			DayName:       dayNames[dow],
			SuccessRate:   successRate,
			ActivityCount: dayActivity[dow],
			IsDanger:      isDanger,
		})
	}

	if dangerDays == nil {
		dangerDays = []string{}
	}

	recommendation := "素晴らしい！全曜日でバランスよく運動できています。"
	if len(dangerDays) > 0 {
		recommendation = fmt.Sprintf("%sが危険です。%sは運動をサボりやすい傾向があります。",
			strings.Join(dangerDays, "と"), strings.Join(dangerDays, "と"))
	}

	return c.JSON(http.StatusOK, response.PredictionResponse{
		UserID:              uid,
		AnalysisPeriodWeeks: analysisPeriodWeeks,
		DailyStats:          dailyStats,
		DangerDays:          dangerDays,
		Recommendation:      recommendation,
	})
}
