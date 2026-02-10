package middleware

import (
	"net/http"
	"strings"

	"github.com/trihackathon/api/adapter"

	"github.com/labstack/echo/v4"
)

// FirebaseAuth はFirebase IDトークンを検証するミドルウェア
func FirebaseAuth(fa *adapter.FirebaseAdapter) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Authorizationヘッダーからトークンを取得
			authHeader := c.Request().Header.Get("Authorization")
			if authHeader == "" {
				return c.JSON(http.StatusUnauthorized, map[string]string{
					"error": "Authorization header is required",
				})
			}

			// "Bearer <token>" の形式をパース
			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || parts[0] != "Bearer" {
				return c.JSON(http.StatusUnauthorized, map[string]string{
					"error": "Invalid authorization format. Use: Bearer <token>",
				})
			}

			idToken := parts[1]

			// Firebase でトークンを検証
			token, err := fa.VerifyToken(c.Request().Context(), idToken)
			if err != nil {
				return c.JSON(http.StatusUnauthorized, map[string]string{
					"error": "Invalid or expired token",
				})
			}

			// ユーザー情報をコンテキストにセット（後続のハンドラで使える）
			c.Set("uid", token.UID)
			c.Set("email", token.Claims["email"])
			c.Set("token", token)

			return next(c)
		}
	}
}
