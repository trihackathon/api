package service

import (
	"fmt"
	"time"

	"github.com/trihackathon/api/models"
	"github.com/trihackathon/api/utils"
	"gorm.io/gorm"
)

type EvaluationService struct {
	db *gorm.DB
}

func NewEvaluationService(db *gorm.DB) *EvaluationService {
	return &EvaluationService{db: db}
}

type EvaluationResult struct {
	EvaluatedTeams int `json:"evaluated_teams"`
	DisbandedTeams int `json:"disbanded_teams"`
}

func (s *EvaluationService) RunWeeklyEvaluation() (*EvaluationResult, error) {
	var teams []models.Team
	if err := s.db.Where("status = ?", "active").Find(&teams).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch active teams: %w", err)
	}

	result := &EvaluationResult{}

	for _, team := range teams {
		if team.StartedAt == nil || team.CurrentWeek == 0 {
			continue
		}

		err := s.evaluateTeam(team)
		if err != nil {
			// Skip this team but continue with others
			fmt.Printf("failed to evaluate team %s: %v\n", team.ID, err)
			continue
		}
		result.EvaluatedTeams++

		// Check if team was disbanded
		var updatedTeam models.Team
		s.db.First(&updatedTeam, "id = ?", team.ID)
		if updatedTeam.Status == "disbanded" {
			result.DisbandedTeams++
		}
	}

	return result, nil
}

func (s *EvaluationService) evaluateTeam(team models.Team) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		// Get goal
		var goal models.Goal
		if err := tx.First(&goal, "team_id = ?", team.ID).Error; err != nil {
			return fmt.Errorf("goal not found: %w", err)
		}

		// Calculate week period
		weekStart := team.StartedAt.AddDate(0, 0, (team.CurrentWeek-1)*7)
		weekEnd := weekStart.AddDate(0, 0, 7)

		// Only evaluate if the week has ended
		if time.Now().Before(weekEnd) {
			return nil
		}

		// Check if already evaluated for this week
		var existingCount int64
		tx.Model(&models.WeeklyEvaluation{}).
			Where("team_id = ? AND week_number = ?", team.ID, team.CurrentWeek).
			Count(&existingCount)
		if existingCount > 0 {
			return nil
		}

		// Get all members
		var members []models.TeamMember
		if err := tx.Where("team_id = ?", team.ID).Find(&members).Error; err != nil {
			return fmt.Errorf("failed to fetch members: %w", err)
		}

		allMet := true
		totalHPChange := 0

		for _, member := range members {
			// Get completed activities for this member in this week (exclude rejected)
			var activities []models.Activity
			tx.Where("user_id = ? AND team_id = ? AND status = ? AND started_at >= ? AND started_at < ? AND (review_status IS NULL OR review_status != ?)",
				member.UserID, team.ID, "completed", weekStart, weekEnd, "rejected").
				Find(&activities)

			var totalDist float64
			var totalVisits int
			var totalDuration int
			var qualifiedVisits int // 滞在時間が目標を満たした訪問回数

			for _, a := range activities {
				totalDist += a.DistanceKM
				totalDuration += a.DurationMin
				if a.ExerciseType == "gym" {
					totalVisits++
					// target_min_duration_min が設定されている場合はその時間以上の訪問のみカウント
					if goal.TargetMinDurationMin != nil {
						if a.DurationMin >= *goal.TargetMinDurationMin {
							qualifiedVisits++
						}
					} else {
						qualifiedVisits++
					}
				}
			}

			// Check if target is met
			targetMet := false
			switch team.ExerciseType {
			case "running":
				if goal.TargetDistanceKM != nil && totalDist >= *goal.TargetDistanceKM {
					targetMet = true
				}
			case "gym":
				// 達成条件: 目標滞在時間を満たした訪問回数が目標回数以上
				if goal.TargetVisitsPerWeek != nil && qualifiedVisits >= *goal.TargetVisitsPerWeek {
					targetMet = true
				}
			}

			// Calculate HP change
			hpChange := 0
			if !targetMet {
				allMet = false
				switch team.Strictness {
				case "relaxed":
					hpChange = -10
				case "normal":
					hpChange = -15
				case "strict":
					hpChange = -25
				default:
					hpChange = -15
				}
			}

			totalHPChange += hpChange

			// Create evaluation record
			eval := models.WeeklyEvaluation{
				ID:               utils.GenerateULID(),
				TeamID:           team.ID,
				UserID:           member.UserID,
				WeekNumber:       team.CurrentWeek,
				TargetMet:        targetMet,
				TotalDistanceKM:  totalDist,
				TotalVisits:      totalVisits,
				TotalDurationMin: totalDuration,
				HPChange:         hpChange,
				EvaluatedAt:      time.Now(),
			}
			if err := tx.Create(&eval).Error; err != nil {
				return fmt.Errorf("failed to create evaluation: %w", err)
			}
		}

		// All members met bonus: +5 per member
		if allMet && len(members) > 0 {
			bonus := 5
			totalHPChange += bonus * len(members)
			// Update each evaluation record with the bonus
			tx.Model(&models.WeeklyEvaluation{}).
				Where("team_id = ? AND week_number = ?", team.ID, team.CurrentWeek).
				Update("hp_change", gorm.Expr("hp_change + ?", bonus))
		}

		// Update team HP
		newHP := team.CurrentHP + totalHPChange
		if newHP < 0 {
			newHP = 0
		}
		if newHP > team.MaxHP {
			newHP = team.MaxHP
		}

		updates := map[string]interface{}{
			"current_hp":   newHP,
			"current_week": team.CurrentWeek + 1,
		}

		// Disband if HP <= 0
		if newHP <= 0 {
			updates["status"] = "disbanded"
		}

		if err := tx.Model(&models.Team{}).Where("id = ?", team.ID).Updates(updates).Error; err != nil {
			return fmt.Errorf("failed to update team: %w", err)
		}

		return nil
	})
}
