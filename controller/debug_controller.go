package controller

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/trihackathon/api/adapter"
	"github.com/trihackathon/api/requests"
	"github.com/trihackathon/api/response"
)

type DebugController struct {
	fa *adapter.FirebaseAdapter
}

func NewDebugController(fa *adapter.FirebaseAdapter) *DebugController {
	return &DebugController{fa: fa}
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
