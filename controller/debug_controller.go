package controller

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/trihackathon/api/adapter"
	"github.com/trihackathon/api/models"
	"github.com/trihackathon/api/requests"
	"github.com/trihackathon/api/response"
	"gorm.io/gorm"
)

type DebugController struct {
	fa *adapter.FirebaseAdapter
	db *gorm.DB
}

func NewDebugController(fa *adapter.FirebaseAdapter, db *gorm.DB) *DebugController {
	return &DebugController{fa: fa, db: db}
}

// @Summary Health check
// @Tags debug
// @Success 200 {object} map[string]string "OK! API is healthy"
// @Router /debug/health [get]
func (ctrl *DebugController) Health(ctx echo.Context) error {
	return ctx.JSON(http.StatusOK, map[string]string{
		"status": "healthy",
	})
}

// @Summary List endpoints
// @Tags debug
// @Success 200 {object} map[string]interface{} "List of available endpoints"
// @Router /debug/endpoints [get]
func (ctrl *DebugController) Endpoints(ctx echo.Context) error {
	return ctx.JSON(http.StatusOK, map[string]interface{}{
		"endpoints": []string{
			"GET /debug/health",
			"GET /debug/endpoints",
			"POST /debug/echo",
			"GET /debug/token?uid=xxx",
		},
	})
}

// @Summary Echo
// @Tags debug
// @Param message body requests.DebugEchoRequest true "メッセージ"
// @Success 200 {object} response.DebugEchoResponse "OK! Echo is healthy"
// @Router /debug/echo [post]
func (ctrl *DebugController) Echo(ctx echo.Context) error {
	req := new(requests.DebugEchoRequest)
	if err := ctx.Bind(req); err != nil {
		return ctx.JSON(http.StatusBadRequest, map[string]interface{}{
			"message": "Failed to bind request",
		})
	}
	return ctx.JSON(http.StatusOK, &response.DebugEchoResponse{Message: req.Message})
}

// Token デバッグ用IDトークン生成
// @Summary デバッグ用IDトークン生成
// @Description Firebase カスタムトークンを生成し、IDトークンに交換して返す（開発環境専用）
// @Tags debug
// @Produce json
// @Param uid query string true "Firebase UID"
// @Success 200 {object} response.DebugTokenResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /debug/token [get]
func (ctrl *DebugController) Token(ctx echo.Context) error {
	if os.Getenv("FLAVOR") != "dev" {
		return ctx.JSON(http.StatusForbidden, response.ErrorResponse{
			Error:   "forbidden",
			Message: "このエンドポイントは開発環境でのみ使用できます",
		})
	}

	uid := ctx.QueryParam("uid")
	if uid == "" {
		return ctx.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error:   "missing_uid",
			Message: "uid クエリパラメータが必要です",
		})
	}

	// カスタムトークン生成
	customToken, err := ctrl.fa.CreateCustomToken(ctx.Request().Context(), uid)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Error:   "custom_token_failed",
			Message: fmt.Sprintf("カスタムトークンの生成に失敗しました: %v", err),
		})
	}

	// カスタムトークンをIDトークンに交換
	apiKey := os.Getenv("FIREBASE_API_KEY")
	if apiKey == "" {
		return ctx.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Error:   "missing_api_key",
			Message: "FIREBASE_API_KEY が設定されていません",
		})
	}

	idToken, err := exchangeCustomTokenForIDToken(customToken, apiKey)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Error:   "token_exchange_failed",
			Message: fmt.Sprintf("IDトークンの取得に失敗しました: %v", err),
		})
	}

	return ctx.JSON(http.StatusOK, response.DebugTokenResponse{
		IDToken: idToken,
		UID:     uid,
	})
}

func exchangeCustomTokenForIDToken(customToken, apiKey string) (string, error) {
	url := fmt.Sprintf("https://identitytoolkit.googleapis.com/v1/accounts:signInWithCustomToken?key=%s", apiKey)

	body := fmt.Sprintf(`{"token":"%s","returnSecureToken":true}`, customToken)
	resp, err := http.Post(url, "application/json", strings.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("リクエスト失敗: %w", err)
	}
	defer resp.Body.Close()

	var result struct {
		IDToken string `json:"idToken"`
		Error   *struct {
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("レスポンスのパースに失敗: %w", err)
	}
	if result.Error != nil {
		return "", fmt.Errorf("Firebase エラー: %s", result.Error.Message)
	}
	if result.IDToken == "" {
		return "", fmt.Errorf("IDトークンが空です")
	}
	return result.IDToken, nil
}

// CleanupDisbandedTeams 解散済みチームのメンバーと投票を削除
// @Summary 解散済みチームのクリーンアップ
// @Description 解散済み（disbanded）チームのteam_membersとdisband_votesレコードを削除（開発環境専用）
// @Tags debug
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 403 {object} response.ErrorResponse
// @Router /debug/cleanup-disbanded-teams [post]
func (ctrl *DebugController) CleanupDisbandedTeams(ctx echo.Context) error {
	if os.Getenv("FLAVOR") != "dev" {
		return ctx.JSON(http.StatusForbidden, response.ErrorResponse{
			Error:   "forbidden",
			Message: "このエンドポイントは開発環境でのみ使用できます",
		})
	}

	// 解散済みチームのIDを取得
	var disbandedTeams []models.Team
	if err := ctrl.db.Where("status = ?", "disbanded").Find(&disbandedTeams).Error; err != nil {
		return ctx.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Error:   "query_failed",
			Message: fmt.Sprintf("チームの取得に失敗: %v", err),
		})
	}

	teamIDs := make([]string, len(disbandedTeams))
	for i, team := range disbandedTeams {
		teamIDs[i] = team.ID
	}

	if len(teamIDs) == 0 {
		return ctx.JSON(http.StatusOK, map[string]interface{}{
			"message":         "解散済みチームはありません",
			"deleted_votes":   0,
			"deleted_members": 0,
		})
	}

	// トランザクションでクリーンアップ
	var deletedVotes int64
	var deletedMembers int64

	err := ctrl.db.Transaction(func(tx *gorm.DB) error {
		// 解散投票を削除
		result := tx.Where("team_id IN ?", teamIDs).Delete(&models.DisbandVote{})
		if result.Error != nil {
			return result.Error
		}
		deletedVotes = result.RowsAffected

		// チームメンバーを削除
		result = tx.Where("team_id IN ?", teamIDs).Delete(&models.TeamMember{})
		if result.Error != nil {
			return result.Error
		}
		deletedMembers = result.RowsAffected

		return nil
	})

	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Error:   "cleanup_failed",
			Message: fmt.Sprintf("クリーンアップに失敗: %v", err),
		})
	}

	return ctx.JSON(http.StatusOK, map[string]interface{}{
		"message":         "クリーンアップ完了",
		"disbanded_teams": len(teamIDs),
		"deleted_votes":   deletedVotes,
		"deleted_members": deletedMembers,
		"team_ids":        teamIDs,
	})
}

// GetUserTeamStatus 現在のユーザーのチーム所属状況を確認
// @Summary ユーザーのチーム所属状況確認
// @Description 現在のユーザーが所属しているチーム、メンバーレコード、チームステータスを確認（開発環境専用）
// @Tags debug
// @Produce json
// @Param uid query string true "ユーザーID（Firebase UID）"
// @Success 200 {object} map[string]interface{}
// @Failure 403 {object} response.ErrorResponse
// @Router /debug/user-team-status [get]
func (ctrl *DebugController) GetUserTeamStatus(ctx echo.Context) error {
	if os.Getenv("FLAVOR") != "dev" {
		return ctx.JSON(http.StatusForbidden, response.ErrorResponse{
			Error:   "forbidden",
			Message: "このエンドポイントは開発環境でのみ使用できます",
		})
	}

	uid := ctx.QueryParam("uid")
	if uid == "" {
		return ctx.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error:   "missing_uid",
			Message: "uid クエリパラメータが必要です",
		})
	}

	// ユーザーのすべてのチームメンバーレコードを取得
	var members []models.TeamMember
	if err := ctrl.db.Preload("Team").Where("user_id = ?", uid).Find(&members).Error; err != nil {
		return ctx.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Error:   "query_failed",
			Message: fmt.Sprintf("クエリ失敗: %v", err),
		})
	}

	type MemberInfo struct {
		MemberID   string `json:"member_id"`
		TeamID     string `json:"team_id"`
		TeamName   string `json:"team_name"`
		TeamStatus string `json:"team_status"`
		Role       string `json:"role"`
		JoinedAt   string `json:"joined_at"`
	}

	memberInfos := make([]MemberInfo, len(members))
	for i, m := range members {
		memberInfos[i] = MemberInfo{
			MemberID:   m.ID,
			TeamID:     m.TeamID,
			TeamName:   m.Team.Name,
			TeamStatus: m.Team.Status,
			Role:       m.Role,
			JoinedAt:   m.JoinedAt.Format("2006-01-02 15:04:05"),
		}
	}

	// アクティブチーム（forming/active）に所属しているか
	var activeMembers []models.TeamMember
	ctrl.db.
		Joins("JOIN teams ON teams.id = team_members.team_id").
		Where("team_members.user_id = ? AND teams.status IN ?", uid, []string{"forming", "active"}).
		Find(&activeMembers)

	return ctx.JSON(http.StatusOK, map[string]interface{}{
		"user_id":             uid,
		"total_memberships":   len(members),
		"active_memberships":  len(activeMembers),
		"all_memberships":     memberInfos,
		"can_create_new_team": len(activeMembers) == 0,
	})
}
