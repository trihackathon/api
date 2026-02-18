package response

import (
	"time"

	"github.com/trihackathon/api/models"
)

// NewTeamResponse TeamモデルからTeamResponseを構築する
func NewTeamResponse(team models.Team, members []models.TeamMember, goal *models.Goal) TeamResponse {
	memberResponses := make([]TeamMemberResponse, len(members))
	for i, m := range members {
		name := m.UserID
		if m.User.Name != "" {
			name = m.User.Name
		}
		memberResponses[i] = TeamMemberResponse{
			UserID:   m.UserID,
			Name:     name,
			Role:     m.Role,
			JoinedAt: m.JoinedAt.Format(time.RFC3339),
		}
	}

	var startedAt *string
	if team.StartedAt != nil {
		s := team.StartedAt.Format(time.RFC3339)
		startedAt = &s
	}

	var goalResponse *GoalResponse
	if goal != nil {
		gr := GoalResponse{
			ID:                   goal.ID,
			TeamID:               goal.TeamID,
			ExerciseType:         goal.ExerciseType,
			TargetDistanceKM:     goal.TargetDistanceKM,
			TargetVisitsPerWeek:  goal.TargetVisitsPerWeek,
			TargetMinDurationMin: goal.TargetMinDurationMin,
			CreatedAt:            goal.CreatedAt.Format(time.RFC3339),
			UpdatedAt:            goal.UpdatedAt.Format(time.RFC3339),
		}
		goalResponse = &gr
	}

	return TeamResponse{
		ID:           team.ID,
		Name:         team.Name,
		ExerciseType: team.ExerciseType,
		Strictness:   team.Strictness,
		Status:       team.Status,
		MaxHP:        team.MaxHP,
		CurrentHP:    team.CurrentHP,
		CurrentWeek:  team.CurrentWeek,
		StartedAt:    startedAt,
		Members:      memberResponses,
		Goal:         goalResponse,
		CreatedAt:    team.CreatedAt.Format(time.RFC3339),
		UpdatedAt:    team.UpdatedAt.Format(time.RFC3339),
	}
}

// TeamMemberResponse チームメンバーレスポンス
type TeamMemberResponse struct {
	UserID   string `json:"user_id" example:"firebaseUID123"`
	Name     string `json:"name" example:"山田太郎"`
	Role     string `json:"role" example:"leader"`
	JoinedAt string `json:"joined_at" example:"2026-02-10T09:00:00Z"`
}

// GoalResponse 目標レスポンス
type GoalResponse struct {
	ID                   string   `json:"id" example:"01JARQ3KEXAMPLE00002"`
	TeamID               string   `json:"team_id" example:"01JARQ3KEXAMPLE00001"`
	ExerciseType         string   `json:"exercise_type" example:"running"`
	TargetDistanceKM     *float64 `json:"target_distance_km" example:"15.0"`
	TargetVisitsPerWeek  *int     `json:"target_visits_per_week"`
	TargetMinDurationMin *int     `json:"target_min_duration_min"`
	CreatedAt            string   `json:"created_at" example:"2026-02-10T09:00:00Z"`
	UpdatedAt            string   `json:"updated_at" example:"2026-02-10T09:00:00Z"`
}

// TeamResponse チームレスポンス
type TeamResponse struct {
	ID           string               `json:"id" example:"01JARQ3KEXAMPLE00001"`
	Name         string               `json:"name" example:"朝ランチーム"`
	ExerciseType string               `json:"exercise_type" example:"running"`
	Strictness   string               `json:"strictness" example:"normal"`
	Status       string               `json:"status" example:"forming"`
	MaxHP        int                  `json:"max_hp" example:"100"`
	CurrentHP    int                  `json:"current_hp" example:"100"`
	CurrentWeek  int                  `json:"current_week" example:"0"`
	StartedAt    *string              `json:"started_at"`
	Members      []TeamMemberResponse `json:"members"`
	Goal         *GoalResponse        `json:"goal,omitempty"`
	CreatedAt    string               `json:"created_at" example:"2026-02-10T09:00:00Z"`
	UpdatedAt    string               `json:"updated_at" example:"2026-02-10T09:00:00Z"`
}

// InviteCodeResponse 招待コードレスポンス
type InviteCodeResponse struct {
	Code               string `json:"code" example:"A3K9X2"`
	TeamID             string `json:"team_id" example:"01JARQ3KEXAMPLE00001"`
	TeamName           string `json:"team_name" example:"朝ランチーム"`
	ExerciseType       string `json:"exercise_type" example:"running"`
	ExpiresAt          string `json:"expires_at" example:"2026-02-11T09:00:00Z"`
	CurrentMemberCount int    `json:"current_member_count" example:"1"`
}

// JoinTeamResponse チーム参加レスポンス
type JoinTeamResponse struct {
	Team      TeamResponse `json:"team"`
	TeamReady bool         `json:"team_ready" example:"false"`
}

// GPSPointResponse GPSポイントレスポンス
type GPSPointResponse struct {
	Latitude  float64 `json:"latitude" example:"35.6812"`
	Longitude float64 `json:"longitude" example:"139.7671"`
	Accuracy  float64 `json:"accuracy" example:"5.0"`
	Timestamp string  `json:"timestamp" example:"2026-02-10T07:00:00Z"`
}

// ActivityResponse アクティビティレスポンス
type ActivityResponse struct {
	ID              string             `json:"id" example:"01JARQ3KEXAMPLE00003"`
	UserID          string             `json:"user_id" example:"firebaseUID123"`
	UserName        string             `json:"user_name" example:"山田太郎"`
	TeamID          string             `json:"team_id" example:"01JARQ3KEXAMPLE00001"`
	ExerciseType    string             `json:"exercise_type" example:"running"`
	Status          string             `json:"status" example:"in_progress"`
	ReviewStatus    string             `json:"review_status" example:"pending"`
	StartedAt       string             `json:"started_at" example:"2026-02-10T07:00:00Z"`
	EndedAt         *string            `json:"ended_at"`
	DistanceKM      float64            `json:"distance_km,omitempty" example:"5.234"`
	GymLocationID   *string            `json:"gym_location_id,omitempty"`
	GymLocationName *string            `json:"gym_location_name,omitempty"`
	AutoDetected    bool               `json:"auto_detected,omitempty" example:"false"`
	DurationMin     int                `json:"duration_min" example:"35"`
	GPSPoints       []GPSPointResponse `json:"gps_points,omitempty"`
	CreatedAt       string             `json:"created_at" example:"2026-02-10T07:00:00Z"`
	UpdatedAt       string             `json:"updated_at" example:"2026-02-10T07:00:00Z"`
}

// ActivityReviewResponse レビューレスポンス
type ActivityReviewResponse struct {
	ID           string `json:"id" example:"01JARQ3KEXAMPLE00020"`
	ActivityID   string `json:"activity_id" example:"01JARQ3KEXAMPLE00003"`
	ReviewerID   string `json:"reviewer_id" example:"firebaseUID456"`
	ReviewerName string `json:"reviewer_name" example:"佐藤花子"`
	Status       string `json:"status" example:"approved"`
	Comment      string `json:"comment" example:"いいペースですね！"`
	CreatedAt    string `json:"created_at" example:"2026-02-10T12:00:00Z"`
}

// SendGPSPointsResponse GPSポイント送信レスポンス
type SendGPSPointsResponse struct {
	SavedCount        int     `json:"saved_count" example:"2"`
	CurrentDistanceKM float64 `json:"current_distance_km" example:"3.456"`
}

// GymLocationResponse ジム位置レスポンス
type GymLocationResponse struct {
	ID        string  `json:"id" example:"01JARQ3KEXAMPLE00004"`
	UserID    string  `json:"user_id" example:"firebaseUID123"`
	Name      string  `json:"name" example:"エニタイムフィットネス 渋谷店"`
	Latitude  float64 `json:"latitude" example:"35.6580"`
	Longitude float64 `json:"longitude" example:"139.7016"`
	RadiusM   int     `json:"radius_m" example:"100"`
	CreatedAt string  `json:"created_at" example:"2026-02-10T09:00:00Z"`
	UpdatedAt string  `json:"updated_at" example:"2026-02-10T09:00:00Z"`
}

// HPChangeEntry HP変動エントリ
type HPChangeEntry struct {
	UserID    string `json:"user_id" example:"firebaseUID123"`
	UserName  string `json:"user_name" example:"山田太郎"`
	HPChange  int    `json:"hp_change" example:"0"`
	TargetMet bool   `json:"target_met" example:"true"`
}

// WeekHPHistory 週別HP履歴
type WeekHPHistory struct {
	Week    int             `json:"week" example:"1"`
	HPStart int             `json:"hp_start" example:"100"`
	HPEnd   int             `json:"hp_end" example:"100"`
	Changes []HPChangeEntry `json:"changes"`
}

// MemberProgress メンバー進捗
type MemberProgress struct {
	UserID                 string   `json:"user_id" example:"firebaseUID123"`
	UserName               string   `json:"user_name" example:"山田太郎"`
	CurrentWeekDistanceKM  *float64 `json:"current_week_distance_km" example:"12.5"`
	CurrentWeekVisits      *int     `json:"current_week_visits"`
	CurrentWeekDurationMin *int     `json:"current_week_duration_min"`
	TargetProgressPercent  float64  `json:"target_progress_percent" example:"83.3"`
}

// TeamStatusResponse チームHP・状態レスポンス
type TeamStatusResponse struct {
	TeamID          string           `json:"team_id" example:"01JARQ3KEXAMPLE00001"`
	Status          string           `json:"status" example:"active"`
	CurrentHP       int              `json:"current_hp" example:"85"`
	MaxHP           int              `json:"max_hp" example:"100"`
	CurrentWeek     int              `json:"current_week" example:"3"`
	StartedAt       string           `json:"started_at" example:"2026-01-20T00:00:00Z"`
	HPHistory       []WeekHPHistory  `json:"hp_history"`
	MembersProgress []MemberProgress `json:"members_progress"`
}

// WeeklyEvaluationResponse 週次評価レスポンス
type WeeklyEvaluationResponse struct {
	ID               string  `json:"id" example:"01JARQ3KEXAMPLE00010"`
	TeamID           string  `json:"team_id" example:"01JARQ3KEXAMPLE00001"`
	UserID           string  `json:"user_id" example:"firebaseUID123"`
	UserName         string  `json:"user_name" example:"山田太郎"`
	WeekNumber       int     `json:"week_number" example:"1"`
	TargetMet        bool    `json:"target_met" example:"true"`
	TotalDistanceKM  float64 `json:"total_distance_km" example:"16.5"`
	TotalVisits      int     `json:"total_visits" example:"0"`
	TotalDurationMin int     `json:"total_duration_min" example:"0"`
	HPChange         int     `json:"hp_change" example:"0"`
	EvaluatedAt      string  `json:"evaluated_at" example:"2026-01-27T00:00:00Z"`
}

// WeekActivitySummary 週間アクティビティサマリー
type WeekActivitySummary struct {
	ID          string  `json:"id" example:"01JARQ3KEXAMPLE00003"`
	Date        string  `json:"date" example:"2026-02-04"`
	DistanceKM  float64 `json:"distance_km,omitempty" example:"5.2"`
	DurationMin int     `json:"duration_min" example:"35"`
}

// CurrentWeekMemberProgress 今週のメンバー進捗
type CurrentWeekMemberProgress struct {
	UserID                string                `json:"user_id" example:"firebaseUID123"`
	UserName              string                `json:"user_name" example:"山田太郎"`
	TotalDistanceKM       float64               `json:"total_distance_km" example:"12.5"`
	TotalVisits           int                   `json:"total_visits" example:"0"`
	QualifiedVisits       int                   `json:"qualified_visits" example:"0"` // 滞在時間目標を満たした訪問回数
	TotalDurationMin      int                   `json:"total_duration_min" example:"0"`
	TargetProgressPercent float64               `json:"target_progress_percent" example:"83.3"`
	OnTrack               bool                  `json:"on_track" example:"true"`
	TargetMultiplier      float64               `json:"target_multiplier" example:"1.0"` // 1.0=通常, 1.5=前週未達成ペナルティ
	ActivitiesThisWeek    []WeekActivitySummary `json:"activities_this_week"`
}

// CurrentWeekEvaluationResponse 今週の進捗レスポンス
type CurrentWeekEvaluationResponse struct {
	TeamID        string                      `json:"team_id" example:"01JARQ3KEXAMPLE00001"`
	WeekNumber    int                         `json:"week_number" example:"3"`
	WeekStart     string                      `json:"week_start" example:"2026-02-03T00:00:00Z"`
	WeekEnd       string                      `json:"week_end" example:"2026-02-09T23:59:59Z"`
	DaysRemaining int                         `json:"days_remaining" example:"0"`
	Members       []CurrentWeekMemberProgress `json:"members"`
}

// DailyStat 日別統計
type DailyStat struct {
	DayOfWeek     int     `json:"day_of_week" example:"0"`
	DayName       string  `json:"day_name" example:"日曜日"`
	SuccessRate   float64 `json:"success_rate" example:"0.75"`
	ActivityCount int     `json:"activity_count" example:"4"`
	IsDanger      bool    `json:"is_danger" example:"false"`
}

// DisbandVoteResponse 解散投票レスポンス
type DisbandVoteResponse struct {
	TeamID     string   `json:"team_id" example:"01JARQ3KEXAMPLE00001"`
	TotalCount int      `json:"total_count" example:"3"`
	VotedCount int      `json:"voted_count" example:"1"`
	VotedUsers []string `json:"voted_users"`
	Disbanded  bool     `json:"disbanded" example:"false"`
}

// PredictionResponse 失敗予測レスポンス
type PredictionResponse struct {
	UserID              string      `json:"user_id" example:"firebaseUID123"`
	AnalysisPeriodWeeks int         `json:"analysis_period_weeks" example:"4"`
	DailyStats          []DailyStat `json:"daily_stats"`
	DangerDays          []string    `json:"danger_days"`
	Recommendation      string      `json:"recommendation" example:"月曜日が危険です。月曜日は運動をサボりやすい傾向があります。"`
}
