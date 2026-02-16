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

type ActivityController struct {
	db *gorm.DB
}

func NewActivityController(db *gorm.DB) *ActivityController {
	return &ActivityController{db: db}
}

// StartRunning ランニング開始
// @Summary      ランニング開始
// @Description  ランニングアクティビティを開始する。同時に進行中にできるアクティビティは1つのみ。チームがactive状態かつexercise_typeがrunningの場合のみ。
// @Tags         activities-running
// @Accept       json
// @Produce      json
// @Param        body  body      requests.StartRunningRequest  true  "開始地点情報"
// @Success      201   {object}  response.ActivityResponse
// @Failure      404   {object}  response.ErrorResponse
// @Failure      409   {object}  response.ErrorResponse
// @Failure      422   {object}  response.ErrorResponse
// @Router       /api/activities/running/start [post]
// @Security     BearerAuth
func (ctrl *ActivityController) StartRunning(c echo.Context) error {
	uid := c.Get("uid").(string)

	req := new(requests.StartRunningRequest)
	if err := c.Bind(req); err != nil {
		return c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error:   "invalid_request",
			Message: "リクエストの形式が不正です",
		})
	}

	// 進行中のアクティビティが既にないかチェック
	var existingActivity models.Activity
	if err := ctrl.db.Where("user_id = ? AND status = ?", uid, "in_progress").
		First(&existingActivity).Error; err == nil {
		return c.JSON(http.StatusConflict, response.ErrorResponse{
			Error:   "activity_already_in_progress",
			Message: "既に進行中のアクティビティがあります",
		})
	}

	// アクティビティを作成（GPS機能のみ、Team不要）
	now := time.Now()
	activityID := utils.GenerateULID()
	activity := models.Activity{
		ID:           activityID,
		UserID:       uid,
		TeamID:       nil, // チーム機能は未実装
		ExerciseType: "running",
		Status:       "in_progress",
		StartedAt:    now,
		DistanceKM:   0,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	if err := ctrl.db.Create(&activity).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Error:   "create_failed",
			Message: "アクティビティの作成に失敗しました",
		})
	}

	// 初期GPSポイントを保存
	initialPoint := models.GPSPoint{
		ID:         utils.GenerateULID(),
		ActivityID: activityID,
		Latitude:   req.Latitude,
		Longitude:  req.Longitude,
		Accuracy:   0, // 初期ポイントは精度不明
		Timestamp:  now,
	}

	if err := ctrl.db.Create(&initialPoint).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Error:   "create_failed",
			Message: "GPSポイントの保存に失敗しました",
		})
	}

	return c.JSON(http.StatusCreated, toActivityResponse(activity, []models.GPSPoint{initialPoint}))
}

// FinishRunning ランニング完了
// @Summary      ランニング完了
// @Description  ランニングアクティビティを完了する。GPSポイントから総移動距離を再計算しdistance_kmを確定。
// @Tags         activities-running
// @Accept       json
// @Produce      json
// @Param        activityId  path      string                       true  "アクティビティID"
// @Param        body        body      requests.FinishRunningRequest  true  "終了地点情報"
// @Success      200         {object}  response.ActivityResponse
// @Failure      403         {object}  response.ErrorResponse
// @Failure      404         {object}  response.ErrorResponse
// @Failure      422         {object}  response.ErrorResponse
// @Router       /api/activities/running/{activityId}/finish [post]
// @Security     BearerAuth
func (ctrl *ActivityController) FinishRunning(c echo.Context) error {
	uid := c.Get("uid").(string)
	activityId := c.Param("activityId")

	req := new(requests.FinishRunningRequest)
	if err := c.Bind(req); err != nil {
		return c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error:   "invalid_request",
			Message: "リクエストの形式が不正です",
		})
	}

	// アクティビティを取得して権限チェック
	var activity models.Activity
	if err := ctrl.db.First(&activity, "id = ?", activityId).Error; err != nil {
		return c.JSON(http.StatusNotFound, response.ErrorResponse{
			Error:   "activity_not_found",
			Message: "アクティビティが見つかりません",
		})
	}

	if activity.UserID != uid {
		return c.JSON(http.StatusForbidden, response.ErrorResponse{
			Error:   "not_activity_owner",
			Message: "このアクティビティの所有者ではありません",
		})
	}

	if activity.Status != "in_progress" {
		return c.JSON(http.StatusUnprocessableEntity, response.ErrorResponse{
			Error:   "activity_not_in_progress",
			Message: "アクティビティが進行中ではありません",
		})
	}

	// 終了地点を保存
	now := time.Now()
	endPoint := models.GPSPoint{
		ID:         utils.GenerateULID(),
		ActivityID: activityId,
		Latitude:   req.Latitude,
		Longitude:  req.Longitude,
		Accuracy:   0,
		Timestamp:  now,
	}

	if err := ctrl.db.Create(&endPoint).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Error:   "create_failed",
			Message: "終了地点の保存に失敗しました",
		})
	}

	// 全GPSポイントを取得して距離を再計算（データ整合性のため）
	var allPoints []models.GPSPoint
	if err := ctrl.db.Where("activity_id = ?", activityId).
		Order("timestamp ASC").
		Find(&allPoints).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Error:   "fetch_failed",
			Message: "GPSポイントの取得に失敗しました",
		})
	}

	// 距離を再計算
	totalDistance := calculateTotalDistance(allPoints)

	// アクティビティを完了状態に更新
	activity.Status = "completed"
	activity.EndedAt = &now
	activity.DistanceKM = totalDistance
	activity.DurationMin = int(now.Sub(activity.StartedAt).Minutes())

	if err := ctrl.db.Save(&activity).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Error:   "update_failed",
			Message: "アクティビティの更新に失敗しました",
		})
	}

	// GPSポイントも含めてレスポンス
	return c.JSON(http.StatusOK, toActivityResponse(activity, allPoints))
}

// SendGPSPoints GPSポイント送信（バッチ）
// @Summary      GPSポイント送信（バッチ）
// @Description  バックグラウンドで蓄積したGPSデータをバッチ送信する。精度が50mを超えるポイントは距離計算から除外（保存はする）。
// @Tags         activities-running
// @Accept       json
// @Produce      json
// @Param        activityId  path      string                         true  "アクティビティID"
// @Param        body        body      requests.SendGPSPointsRequest  true  "GPSポイントデータ"
// @Success      200         {object}  response.SendGPSPointsResponse
// @Failure      400         {object}  response.ErrorResponse
// @Failure      403         {object}  response.ErrorResponse
// @Failure      404         {object}  response.ErrorResponse
// @Failure      422         {object}  response.ErrorResponse
// @Router       /api/activities/running/{activityId}/gps [post]
// @Security     BearerAuth
func (ctrl *ActivityController) SendGPSPoints(c echo.Context) error {
	uid := c.Get("uid").(string)
	activityId := c.Param("activityId")

	req := new(requests.SendGPSPointsRequest)
	if err := c.Bind(req); err != nil {
		return c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error:   "invalid_request",
			Message: "リクエストの形式が不正です",
		})
	}

	// アクティビティを取得して権限チェック
	var activity models.Activity
	if err := ctrl.db.First(&activity, "id = ?", activityId).Error; err != nil {
		return c.JSON(http.StatusNotFound, response.ErrorResponse{
			Error:   "activity_not_found",
			Message: "アクティビティが見つかりません",
		})
	}

	if activity.UserID != uid {
		return c.JSON(http.StatusForbidden, response.ErrorResponse{
			Error:   "not_activity_owner",
			Message: "このアクティビティの所有者ではありません",
		})
	}

	if activity.Status != "in_progress" {
		return c.JSON(http.StatusUnprocessableEntity, response.ErrorResponse{
			Error:   "activity_not_in_progress",
			Message: "アクティビティが進行中ではありません",
		})
	}

	// 最後に有効だったGPSポイントを取得（accuracy <= 50）
	var lastValidPoint models.GPSPoint
	hasLastPoint := false
	if err := ctrl.db.Where("activity_id = ? AND (accuracy IS NULL OR accuracy <= 50)", activityId).
		Order("timestamp DESC").
		First(&lastValidPoint).Error; err == nil {
		hasLastPoint = true
	}

	// 新規ポイントを保存
	savedPoints := []models.GPSPoint{}
	for _, reqPoint := range req.Points {
		timestamp, err := time.Parse(time.RFC3339, reqPoint.Timestamp)
		if err != nil {
			continue // 不正なタイムスタンプはスキップ
		}

		// PWA重複防止: ClientIDが指定されている場合、既に存在するかチェック
		if reqPoint.ClientID != nil && *reqPoint.ClientID != "" {
			var existing models.GPSPoint
			if err := ctrl.db.Where("client_id = ?", *reqPoint.ClientID).First(&existing).Error; err == nil {
				// 既に存在する場合はスキップ
				continue
			}
		}

		point := models.GPSPoint{
			ID:         utils.GenerateULID(),
			ActivityID: activityId,
			ClientID:   reqPoint.ClientID,
			Latitude:   reqPoint.Latitude,
			Longitude:  reqPoint.Longitude,
			Accuracy:   reqPoint.Accuracy,
			Timestamp:  timestamp,
		}

		if err := ctrl.db.Create(&point).Error; err != nil {
			continue // エラーがあってもスキップして継続（重複エラーも含む）
		}

		savedPoints = append(savedPoints, point)
	}

	// 差分距離を計算（最適化案A）
	additionalDistance := 0.0
	prevPoint := &lastValidPoint
	usePrev := hasLastPoint

	for i := range savedPoints {
		point := &savedPoints[i]

		// 精度50m超えるポイントは距離計算から除外
		if point.Accuracy > 50.0 {
			continue
		}

		if usePrev {
			dist := utils.Haversine(
				prevPoint.Latitude, prevPoint.Longitude,
				point.Latitude, point.Longitude,
			)

			// 1km超える区間は異常値として除外
			if dist <= 1.0 {
				additionalDistance += dist
			}
		}

		prevPoint = point
		usePrev = true
	}

	// 累積距離を更新（差分を追加）
	if err := ctrl.db.Model(&activity).
		UpdateColumn("distance_km", gorm.Expr("distance_km + ?", additionalDistance)).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Error:   "update_failed",
			Message: "距離の更新に失敗しました",
		})
	}

	// 更新後の距離を取得
	ctrl.db.First(&activity, "id = ?", activityId)

	return c.JSON(http.StatusOK, response.SendGPSPointsResponse{
		SavedCount:        len(savedPoints),
		CurrentDistanceKM: activity.DistanceKM,
	})
}

// GetRunningActivity ランニング記録詳細
// @Summary      ランニング記録詳細
// @Description  指定したランニングアクティビティの詳細情報（GPSポイント含む）を取得する
// @Tags         activities-running
// @Produce      json
// @Param        activityId  path      string  true  "アクティビティID"
// @Success      200         {object}  response.ActivityResponse
// @Failure      403         {object}  response.ErrorResponse
// @Failure      404         {object}  response.ErrorResponse
// @Router       /api/activities/running/{activityId} [get]
// @Security     BearerAuth
func (ctrl *ActivityController) GetRunningActivity(c echo.Context) error {
	uid := c.Get("uid").(string)
	activityId := c.Param("activityId")

	var activity models.Activity
	if err := ctrl.db.First(&activity, "id = ?", activityId).Error; err != nil {
		return c.JSON(http.StatusNotFound, response.ErrorResponse{
			Error:   "activity_not_found",
			Message: "アクティビティが見つかりません",
		})
	}

	// 権限チェック（自分のアクティビティのみ）
	if activity.UserID != uid {
		return c.JSON(http.StatusForbidden, response.ErrorResponse{
			Error:   "not_activity_owner",
			Message: "このアクティビティの所有者ではありません",
		})
	}

	// GPSポイントを取得
	var gpsPoints []models.GPSPoint
	ctrl.db.Where("activity_id = ?", activityId).
		Order("timestamp ASC").
		Find(&gpsPoints)

	return c.JSON(http.StatusOK, toActivityResponse(activity, gpsPoints))
}

// GetMyActivities 自分のアクティビティ一覧
// @Summary      自分のアクティビティ一覧
// @Description  自分のアクティビティ一覧を取得する
// @Tags         activities
// @Produce      json
// @Success      200  {array}   response.ActivityResponse
// @Failure      404  {object}  response.ErrorResponse
// @Router       /api/activities [get]
// @Security     BearerAuth
func (ctrl *ActivityController) GetMyActivities(c echo.Context) error {
	uid := c.Get("uid").(string)

	var activities []models.Activity
	if err := ctrl.db.Where("user_id = ?", uid).
		Order("started_at DESC").
		Find(&activities).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Error:   "fetch_failed",
			Message: "アクティビティの取得に失敗しました",
		})
	}

	responses := make([]response.ActivityResponse, len(activities))
	for i, activity := range activities {
		responses[i] = toActivityResponse(activity, nil) // GPSポイントは含めない
	}

	return c.JSON(http.StatusOK, responses)
}

// GetTeamActivities チーム全体のアクティビティ一覧
// @Summary      チーム全体のアクティビティ一覧
// @Description  チーム全体のアクティビティ一覧を取得する
// @Tags         activities
// @Produce      json
// @Param        teamId  path      string  true  "チームID"
// @Success      200     {array}   response.ActivityResponse
// @Failure      403     {object}  response.ErrorResponse
// @Failure      404     {object}  response.ErrorResponse
// @Router       /api/teams/{teamId}/activities [get]
// @Security     BearerAuth
func (ctrl *ActivityController) GetTeamActivities(c echo.Context) error {
	// TODO: チーム機能未実装
	return c.JSON(http.StatusNotImplemented, response.ErrorResponse{
		Error:   "not_implemented",
		Message: "チーム機能は未実装です",
	})
}

// Helper functions

// calculateTotalDistance 全GPSポイントから総距離を計算
func calculateTotalDistance(points []models.GPSPoint) float64 {
	if len(points) < 2 {
		return 0
	}

	totalDistance := 0.0
	var prevPoint *models.GPSPoint

	for i := range points {
		point := &points[i]

		// 精度50m超えるポイントは距離計算から除外
		if point.Accuracy > 50.0 {
			continue
		}

		if prevPoint != nil {
			dist := utils.Haversine(
				prevPoint.Latitude, prevPoint.Longitude,
				point.Latitude, point.Longitude,
			)

			// 1km超える区間は異常値として除外
			if dist <= 1.0 {
				totalDistance += dist
			}
		}

		prevPoint = point
	}

	return totalDistance
}

// toActivityResponse ActivityモデルからレスポンスDTOに変換
func toActivityResponse(activity models.Activity, gpsPoints []models.GPSPoint) response.ActivityResponse {
	resp := response.ActivityResponse{
		ID:           activity.ID,
		UserID:       activity.UserID,
		TeamID:       "", // TeamIDはnullableなので空文字列を返す
		ExerciseType: activity.ExerciseType,
		Status:       activity.Status,
		StartedAt:    activity.StartedAt.Format(time.RFC3339),
		DistanceKM:   activity.DistanceKM,
		DurationMin:  activity.DurationMin,
		CreatedAt:    activity.CreatedAt.Format(time.RFC3339),
		UpdatedAt:    activity.UpdatedAt.Format(time.RFC3339),
	}

	if activity.TeamID != nil {
		resp.TeamID = *activity.TeamID
	}

	if activity.EndedAt != nil {
		endedStr := activity.EndedAt.Format(time.RFC3339)
		resp.EndedAt = &endedStr
	}

	if activity.GymLocationID != nil {
		resp.GymLocationID = activity.GymLocationID
		resp.AutoDetected = activity.AutoDetected
	}

	if len(gpsPoints) > 0 {
		resp.GPSPoints = make([]response.GPSPointResponse, len(gpsPoints))
		for i, point := range gpsPoints {
			resp.GPSPoints[i] = response.GPSPointResponse{
				Latitude:  point.Latitude,
				Longitude: point.Longitude,
				Accuracy:  point.Accuracy,
				Timestamp: point.Timestamp.Format(time.RFC3339),
			}
		}
	}

	return resp
}
