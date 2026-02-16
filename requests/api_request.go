package requests

// CreateTeamRequest チーム作成リクエスト
type CreateTeamRequest struct {
	Name         string `json:"name" example:"朝ランチーム"`
	ExerciseType string `json:"exercise_type" example:"running"`
	Strictness   string `json:"strictness" example:"normal"`
}

// JoinTeamRequest 招待コードでチーム参加リクエスト
type JoinTeamRequest struct {
	Code string `json:"code" example:"A3K9X2"`
}

// CreateGoalRequest 目標設定リクエスト
type CreateGoalRequest struct {
	TargetDistanceKM     *float64 `json:"target_distance_km" example:"15.0"`
	TargetVisitsPerWeek  *int     `json:"target_visits_per_week" example:"3"`
	TargetMinDurationMin *int     `json:"target_min_duration_min" example:"60"`
}

// StartRunningRequest ランニング開始リクエスト
type StartRunningRequest struct {
	Latitude  float64 `json:"latitude" example:"35.6812362"`
	Longitude float64 `json:"longitude" example:"139.7671248"`
}

// FinishRunningRequest ランニング完了リクエスト
type FinishRunningRequest struct {
	Latitude  float64 `json:"latitude" example:"35.6815"`
	Longitude float64 `json:"longitude" example:"139.7675"`
}

// GPSPointRequest GPSポイント
type GPSPointRequest struct {
	Latitude  float64 `json:"latitude" example:"35.6812"`
	Longitude float64 `json:"longitude" example:"139.7671"`
	Accuracy  float64 `json:"accuracy" example:"5.0"`
	Timestamp string  `json:"timestamp" example:"2026-02-10T07:01:00Z"`
}

// SendGPSPointsRequest GPSポイント送信リクエスト
type SendGPSPointsRequest struct {
	Points []GPSPointRequest `json:"points"`
}

// CreateGymLocationRequest ジム位置登録リクエスト
type CreateGymLocationRequest struct {
	Name      string  `json:"name" example:"エニタイムフィットネス 渋谷店"`
	Latitude  float64 `json:"latitude" example:"35.6580"`
	Longitude float64 `json:"longitude" example:"139.7016"`
	RadiusM   int     `json:"radius_m" example:"100"`
}

// GymCheckinRequest ジムチェックインリクエスト
type GymCheckinRequest struct {
	GymLocationID string  `json:"gym_location_id" example:"01JARQ3KEXAMPLE00004"`
	Latitude      float64 `json:"latitude" example:"35.6581"`
	Longitude     float64 `json:"longitude" example:"139.7017"`
	AutoDetected  bool    `json:"auto_detected" example:"false"`
}

// GymCheckoutRequest ジムチェックアウトリクエスト
type GymCheckoutRequest struct {
	Latitude  float64 `json:"latitude" example:"35.6581"`
	Longitude float64 `json:"longitude" example:"139.7017"`
}
