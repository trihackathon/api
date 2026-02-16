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

type GymController struct {
	db *gorm.DB
}

func NewGymController(db *gorm.DB) *GymController {
	return &GymController{db: db}
}

// CreateGymLocation ジム位置登録
// @Summary      ジム位置登録
// @Description  ジムの位置情報を登録する。ジオフェンス半径のデフォルトは100m。
// @Tags         gym
// @Accept       json
// @Produce      json
// @Param        body  body      requests.CreateGymLocationRequest  true  "ジム位置情報"
// @Success      201   {object}  response.GymLocationResponse
// @Failure      400   {object}  response.ErrorResponse
// @Router       /api/gym-locations [post]
// @Security     BearerAuth
func (ctrl *GymController) CreateGymLocation(c echo.Context) error {
	uid := c.Get("uid").(string)

	req := new(requests.CreateGymLocationRequest)
	if err := c.Bind(req); err != nil {
		return c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error:   "invalid_request",
			Message: "リクエストの形式が不正です",
		})
	}

	// バリデーション
	if req.Name == "" || len(req.Name) > 100 {
		return c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error:   "invalid_request",
			Message: "ジム名は1〜100文字で指定してください",
		})
	}
	if req.Latitude < -90 || req.Latitude > 90 {
		return c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error:   "invalid_request",
			Message: "緯度は-90.0〜90.0の範囲で指定してください",
		})
	}
	if req.Longitude < -180 || req.Longitude > 180 {
		return c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error:   "invalid_request",
			Message: "経度は-180.0〜180.0の範囲で指定してください",
		})
	}
	if req.RadiusM != 0 && (req.RadiusM < 50 || req.RadiusM > 500) {
		return c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error:   "invalid_request",
			Message: "ジオフェンス半径は50〜500mの範囲で指定してください",
		})
	}

	// デフォルト値設定
	radiusM := req.RadiusM
	if radiusM == 0 {
		radiusM = 100
	}

	// ジム位置登録
	gymLocation := models.GymLocation{
		ID:        utils.GenerateULID(),
		UserID:    uid,
		Name:      req.Name,
		Latitude:  req.Latitude,
		Longitude: req.Longitude,
		RadiusM:   radiusM,
	}

	if err := ctrl.db.Create(&gymLocation).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Error:   "internal_error",
			Message: "ジム位置の登録に失敗しました",
		})
	}

	return c.JSON(http.StatusCreated, response.GymLocationResponse{
		ID:        gymLocation.ID,
		UserID:    gymLocation.UserID,
		Name:      gymLocation.Name,
		Latitude:  gymLocation.Latitude,
		Longitude: gymLocation.Longitude,
		RadiusM:   gymLocation.RadiusM,
		CreatedAt: gymLocation.CreatedAt.Format(time.RFC3339),
		UpdatedAt: gymLocation.UpdatedAt.Format(time.RFC3339),
	})
}

// GetGymLocations 登録ジム一覧
// @Summary      登録ジム一覧
// @Description  ユーザーが登録したジムの一覧を取得する
// @Tags         gym
// @Produce      json
// @Success      200  {array}   response.GymLocationResponse
// @Router       /api/gym-locations [get]
// @Security     BearerAuth
func (ctrl *GymController) GetGymLocations(c echo.Context) error {
	uid := c.Get("uid").(string)

	// ユーザーの登録ジム一覧を取得
	var gymLocations []models.GymLocation
	if err := ctrl.db.Where("user_id = ?", uid).Order("created_at DESC").Find(&gymLocations).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Error:   "internal_error",
			Message: "ジム一覧の取得に失敗しました",
		})
	}

	// レスポンス作成
	result := make([]response.GymLocationResponse, len(gymLocations))
	for i, gym := range gymLocations {
		result[i] = response.GymLocationResponse{
			ID:        gym.ID,
			UserID:    gym.UserID,
			Name:      gym.Name,
			Latitude:  gym.Latitude,
			Longitude: gym.Longitude,
			RadiusM:   gym.RadiusM,
			CreatedAt: gym.CreatedAt.Format(time.RFC3339),
			UpdatedAt: gym.UpdatedAt.Format(time.RFC3339),
		}
	}

	return c.JSON(http.StatusOK, result)
}

// DeleteGymLocation ジム位置削除
// @Summary      ジム位置削除
// @Description  登録したジムの位置情報を削除する。所有者のみ削除可能。
// @Tags         gym
// @Param        locationId  path  string  true  "ジム位置ID"
// @Success      204  "No Content"
// @Failure      403  {object}  response.ErrorResponse
// @Failure      404  {object}  response.ErrorResponse
// @Router       /api/gym-locations/{locationId} [delete]
// @Security     BearerAuth
func (ctrl *GymController) DeleteGymLocation(c echo.Context) error {
	uid := c.Get("uid").(string)
	locationId := c.Param("locationId")

	// ジム位置を取得
	var gymLocation models.GymLocation
	if err := ctrl.db.Where("id = ?", locationId).First(&gymLocation).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.JSON(http.StatusNotFound, response.ErrorResponse{
				Error:   "location_not_found",
				Message: "ジム位置が見つかりません",
			})
		}
		return c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Error:   "internal_error",
			Message: "ジム位置の取得に失敗しました",
		})
	}

	// 所有者チェック
	if gymLocation.UserID != uid {
		return c.JSON(http.StatusForbidden, response.ErrorResponse{
			Error:   "not_location_owner",
			Message: "このジム位置の所有者ではありません",
		})
	}

	// 削除
	if err := ctrl.db.Delete(&gymLocation).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Error:   "internal_error",
			Message: "ジム位置の削除に失敗しました",
		})
	}

	return c.NoContent(http.StatusNoContent)
}

// GymCheckin ジムチェックイン
// @Summary      ジムチェックイン
// @Description  ジムにチェックインする。現在位置とジムの距離を検証し、radius_m以内であることを確認。
// @Tags         gym
// @Accept       json
// @Produce      json
// @Param        body  body      requests.GymCheckinRequest  true  "チェックイン情報"
// @Success      201   {object}  response.ActivityResponse
// @Failure      404   {object}  response.ErrorResponse
// @Failure      409   {object}  response.ErrorResponse
// @Failure      422   {object}  response.ErrorResponse
// @Router       /api/activities/gym/checkin [post]
// @Security     BearerAuth
func (ctrl *GymController) GymCheckin(c echo.Context) error {
	uid := c.Get("uid").(string)

	req := new(requests.GymCheckinRequest)
	if err := c.Bind(req); err != nil {
		return c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error:   "invalid_request",
			Message: "リクエストの形式が不正です",
		})
	}

	// デバッグ: リクエスト内容をログ出力
	c.Logger().Infof("チェックインリクエスト: gym_location_id=%s, lat=%f, lon=%f", 
		req.GymLocationID, req.Latitude, req.Longitude)

	// バリデーション
	if req.Latitude < -90 || req.Latitude > 90 {
		return c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error:   "invalid_request",
			Message: "緯度は-90.0〜90.0の範囲で指定してください",
		})
	}
	if req.Longitude < -180 || req.Longitude > 180 {
		return c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error:   "invalid_request",
			Message: "経度は-180.0〜180.0の範囲で指定してください",
		})
	}

	// ジム位置を取得
	var gymLocation models.GymLocation
	if err := ctrl.db.Where("id = ?", req.GymLocationID).First(&gymLocation).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.JSON(http.StatusNotFound, response.ErrorResponse{
				Error:   "location_not_found",
				Message: "ジム位置が見つかりません",
			})
		}
		return c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Error:   "internal_error",
			Message: "ジム位置の取得に失敗しました",
		})
	}

	// ユーザーのアクティブなチームを取得（オプショナル）
	var team models.Team
	var teamID *string
	err := ctrl.db.
		Joins("JOIN team_members ON teams.id = team_members.team_id").
		Where("team_members.user_id = ?", uid).
		Where("teams.status IN ?", []string{"forming", "active"}).
		First(&team).Error
	
	if err == nil {
		// チームが見つかった場合、検証を行う
		// チームがactiveか確認
		if team.Status != "active" {
			return c.JSON(http.StatusUnprocessableEntity, response.ErrorResponse{
				Error:   "team_not_active",
				Message: "チームがアクティブではありません",
			})
		}

		// チームの運動タイプがジムか確認
		if team.ExerciseType != "gym" {
			return c.JSON(http.StatusUnprocessableEntity, response.ErrorResponse{
				Error:   "exercise_type_mismatch",
				Message: "チームの運動タイプがジムではありません",
			})
		}
		
		teamID = &team.ID
	} else if err != gorm.ErrRecordNotFound {
		// チームが見つからない以外のエラー
		return c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Error:   "internal_error",
			Message: "チームの取得に失敗しました",
		})
	}
	// チームが見つからない場合（err == gorm.ErrRecordNotFound）は、teamID = nil のまま続行

	// 進行中のアクティビティがないか確認
	var existingActivity models.Activity
	err = ctrl.db.Where("user_id = ? AND status = ?", uid, "in_progress").First(&existingActivity).Error
	if err == nil {
		return c.JSON(http.StatusConflict, response.ErrorResponse{
			Error:   "activity_in_progress",
			Message: "進行中のアクティビティがあります",
		})
	} else if err != gorm.ErrRecordNotFound {
		return c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Error:   "internal_error",
			Message: "アクティビティの確認に失敗しました",
		})
	}

	// 現在位置とジム位置の距離を計算
	distanceKM := utils.Haversine(req.Latitude, req.Longitude, gymLocation.Latitude, gymLocation.Longitude)
	distanceM := distanceKM * 1000

	// ジオフェンス半径内か確認
	if distanceM > float64(gymLocation.RadiusM) {
		return c.JSON(http.StatusUnprocessableEntity, response.ErrorResponse{
			Error:   "too_far_from_gym",
			Message: "ジムから離れすぎています（ジオフェンス半径外）",
		})
	}

	// アクティビティを作成
	now := time.Now()
	activity := models.Activity{
		ID:            utils.GenerateULID(),
		UserID:        uid,
		TeamID:        teamID, // チームがない場合はnil
		ExerciseType:  "gym",
		Status:        "in_progress",
		StartedAt:     now,
		GymLocationID: &gymLocation.ID,
		AutoDetected:  req.AutoDetected,
		DurationMin:   0,
	}

	if err := ctrl.db.Create(&activity).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Error:   "internal_error",
			Message: "アクティビティの作成に失敗しました",
		})
	}

	// レスポンス用のteamID（文字列）
	responseTeamID := ""
	if teamID != nil {
		responseTeamID = *teamID
	}

	return c.JSON(http.StatusCreated, response.ActivityResponse{
		ID:              activity.ID,
		UserID:          activity.UserID,
		TeamID:          responseTeamID,
		ExerciseType:    activity.ExerciseType,
		Status:          activity.Status,
		StartedAt:       activity.StartedAt.Format(time.RFC3339),
		EndedAt:         nil,
		GymLocationID:   activity.GymLocationID,
		GymLocationName: &gymLocation.Name,
		AutoDetected:    activity.AutoDetected,
		DurationMin:     activity.DurationMin,
		CreatedAt:       activity.CreatedAt.Format(time.RFC3339),
		UpdatedAt:       activity.UpdatedAt.Format(time.RFC3339),
	})
}

// GymCheckout ジムチェックアウト
// @Summary      ジムチェックアウト
// @Description  ジムからチェックアウトする。duration_minをended_at - started_atから算出。
// @Tags         gym
// @Accept       json
// @Produce      json
// @Param        activityId  path      string                       true  "アクティビティID"
// @Param        body        body      requests.GymCheckoutRequest  true  "チェックアウト情報"
// @Success      200         {object}  response.ActivityResponse
// @Failure      403         {object}  response.ErrorResponse
// @Failure      404         {object}  response.ErrorResponse
// @Failure      422         {object}  response.ErrorResponse
// @Router       /api/activities/gym/{activityId}/checkout [post]
// @Security     BearerAuth
func (ctrl *GymController) GymCheckout(c echo.Context) error {
	uid := c.Get("uid").(string)
	activityId := c.Param("activityId")

	req := new(requests.GymCheckoutRequest)
	if err := c.Bind(req); err != nil {
		return c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error:   "invalid_request",
			Message: "リクエストの形式が不正です",
		})
	}

	// バリデーション
	if req.Latitude < -90 || req.Latitude > 90 {
		return c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error:   "invalid_request",
			Message: "緯度は-90.0〜90.0の範囲で指定してください",
		})
	}
	if req.Longitude < -180 || req.Longitude > 180 {
		return c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error:   "invalid_request",
			Message: "経度は-180.0〜180.0の範囲で指定してください",
		})
	}

	// アクティビティを取得
	var activity models.Activity
	if err := ctrl.db.Where("id = ?", activityId).First(&activity).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.JSON(http.StatusNotFound, response.ErrorResponse{
				Error:   "activity_not_found",
				Message: "アクティビティが見つかりません",
			})
		}
		return c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Error:   "internal_error",
			Message: "アクティビティの取得に失敗しました",
		})
	}

	// 所有者確認
	if activity.UserID != uid {
		return c.JSON(http.StatusForbidden, response.ErrorResponse{
			Error:   "not_activity_owner",
			Message: "このアクティビティの所有者ではありません",
		})
	}

	// 進行中か確認
	if activity.Status != "in_progress" {
		return c.JSON(http.StatusUnprocessableEntity, response.ErrorResponse{
			Error:   "activity_not_in_progress",
			Message: "アクティビティが進行中ではありません",
		})
	}

	// チェックアウト処理
	now := time.Now()
	durationMin := int(now.Sub(activity.StartedAt).Minutes())

	activity.EndedAt = &now
	activity.Status = "completed"
	activity.DurationMin = durationMin

	if err := ctrl.db.Save(&activity).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Error:   "internal_error",
			Message: "アクティビティの更新に失敗しました",
		})
	}

	// ジム位置名を取得
	var gymLocationName *string
	if activity.GymLocationID != nil {
		var gymLocation models.GymLocation
		if err := ctrl.db.Where("id = ?", *activity.GymLocationID).First(&gymLocation).Error; err == nil {
			gymLocationName = &gymLocation.Name
		}
	}

	// レスポンス用のteamID
	responseTeamID := ""
	if activity.TeamID != nil {
		responseTeamID = *activity.TeamID
	}

	endedAtStr := activity.EndedAt.Format(time.RFC3339)
	return c.JSON(http.StatusOK, response.ActivityResponse{
		ID:              activity.ID,
		UserID:          activity.UserID,
		TeamID:          responseTeamID,
		ExerciseType:    activity.ExerciseType,
		Status:          activity.Status,
		StartedAt:       activity.StartedAt.Format(time.RFC3339),
		EndedAt:         &endedAtStr,
		GymLocationID:   activity.GymLocationID,
		GymLocationName: gymLocationName,
		AutoDetected:    activity.AutoDetected,
		DurationMin:     activity.DurationMin,
		CreatedAt:       activity.CreatedAt.Format(time.RFC3339),
		UpdatedAt:       activity.UpdatedAt.Format(time.RFC3339),
	})
}

// GetGymActivity ジム記録詳細
// @Summary      ジム記録詳細
// @Description  指定したジムアクティビティの詳細情報を取得する
// @Tags         gym
// @Produce      json
// @Param        activityId  path      string  true  "アクティビティID"
// @Success      200         {object}  response.ActivityResponse
// @Failure      403         {object}  response.ErrorResponse
// @Failure      404         {object}  response.ErrorResponse
// @Router       /api/activities/gym/{activityId} [get]
// @Security     BearerAuth
func (ctrl *GymController) GetGymActivity(c echo.Context) error {
	uid := c.Get("uid").(string)
	activityId := c.Param("activityId")

	// アクティビティを取得
	var activity models.Activity
	if err := ctrl.db.Where("id = ?", activityId).First(&activity).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.JSON(http.StatusNotFound, response.ErrorResponse{
				Error:   "activity_not_found",
				Message: "アクティビティが見つかりません",
			})
		}
		return c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Error:   "internal_error",
			Message: "アクティビティの取得に失敗しました",
		})
	}

	// 所有者確認（または同じチームのメンバーか確認）
	if activity.UserID != uid {
		// チームメンバーかチェック
		if activity.TeamID != nil {
			var count int64
			ctrl.db.Model(&models.TeamMember{}).
				Where("team_id = ? AND user_id = ?", *activity.TeamID, uid).
				Count(&count)
			if count == 0 {
				return c.JSON(http.StatusForbidden, response.ErrorResponse{
					Error:   "not_activity_owner",
					Message: "このアクティビティの所有者ではありません",
				})
			}
		} else {
			return c.JSON(http.StatusForbidden, response.ErrorResponse{
				Error:   "not_activity_owner",
				Message: "このアクティビティの所有者ではありません",
			})
		}
	}

	// ジム位置名を取得
	var gymLocationName *string
	if activity.GymLocationID != nil {
		var gymLocation models.GymLocation
		if err := ctrl.db.Where("id = ?", *activity.GymLocationID).First(&gymLocation).Error; err == nil {
			gymLocationName = &gymLocation.Name
		}
	}

	// レスポンス作成
	var endedAtStr *string
	if activity.EndedAt != nil {
		str := activity.EndedAt.Format(time.RFC3339)
		endedAtStr = &str
	}

	teamID := ""
	if activity.TeamID != nil {
		teamID = *activity.TeamID
	}

	return c.JSON(http.StatusOK, response.ActivityResponse{
		ID:              activity.ID,
		UserID:          activity.UserID,
		TeamID:          teamID,
		ExerciseType:    activity.ExerciseType,
		Status:          activity.Status,
		StartedAt:       activity.StartedAt.Format(time.RFC3339),
		EndedAt:         endedAtStr,
		GymLocationID:   activity.GymLocationID,
		GymLocationName: gymLocationName,
		AutoDetected:    activity.AutoDetected,
		DurationMin:     activity.DurationMin,
		CreatedAt:       activity.CreatedAt.Format(time.RFC3339),
		UpdatedAt:       activity.UpdatedAt.Format(time.RFC3339),
	})
}
