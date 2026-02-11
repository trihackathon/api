package main

import (
	"net/http"
	"os"

	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	echoSwagger "github.com/swaggo/echo-swagger"

	"github.com/trihackathon/api/adapter"
	"github.com/trihackathon/api/controller"
	_ "github.com/trihackathon/api/docs" // Swagger docs
	"github.com/trihackathon/api/driver"
	"github.com/trihackathon/api/middleware"
)

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
func main() {
	// .envファイルを読み込む
	if err := godotenv.Load(); err != nil {
		// .envファイルが存在しない場合は警告のみ（本番環境など）
		// log.Printf("Warning: .env file not found: %v", err)
	}
	e := echo.New()

	// DB接続
	db := driver.NewDB()

	// Swagger UI
	e.GET("/swagger/*", echoSwagger.WrapHandler)

	// Firebase初期化
	fa := adapter.NewFirebaseAdapter()

	// コントローラー初期化
	debugController := controller.NewDebugController(fa)
	userController := controller.NewUserController(db)

	// 認証不要のルート
	e.GET("/debug/health", debugController.Health)
	e.GET("/debug/token", debugController.Token)

	// 認証必須のルートグループ
	api := e.Group("/api")
	api.Use(middleware.FirebaseAuth(fa))

	api.GET("/users/me", userController.GetMe)
	api.POST("/users/me", userController.CreateMe)
	api.PUT("/users/me", userController.UpdateMe)

	// Health check endpoint
	// @Summary      Health check
	// @Description  Returns a simple health check message
	// @Tags         health
	// @Accept       json
	// @Produce      json
	// @Success      200  {string}  string  "Hello, World!"
	// @Router       / [get]
	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "Hello, World!")
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	e.Start(":" + port)
}
