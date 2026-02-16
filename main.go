package main

import (
	"net/http"
	"os"

	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	echoSwagger "github.com/swaggo/echo-swagger"

	"github.com/trihackathon/api/adapter"
	"github.com/trihackathon/api/controller"
	_ "github.com/trihackathon/api/docs" // Swagger docs
	"github.com/trihackathon/api/driver"
	authmw "github.com/trihackathon/api/middleware"
)

// @title Trihackathon API
// @version 1.0
// @description Trihackathon API Server
// @host localhost:8080
// @BasePath /
//
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

	// CORS設定（PWA対応）
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins:     []string{"*"}, // 本番環境では具体的なドメインを指定
		AllowMethods:     []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete},
		AllowHeaders:     []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept, echo.HeaderAuthorization},
		ExposeHeaders:    []string{echo.HeaderContentLength},
		AllowCredentials: true,
	}))

	// DB接続
	db := driver.NewDB()

	// Swagger UI
	e.GET("/swagger/*", echoSwagger.WrapHandler)

	// Firebase初期化
	fa := adapter.NewFirebaseAdapter()

	// コントローラー初期化
	debugController := controller.NewDebugController(fa)
	userController := controller.NewUserController(db)
	teamController := controller.NewTeamController(db)
	inviteController := controller.NewInviteController(db)
	goalController := controller.NewGoalController(db)
	activityController := controller.NewActivityController(db)
	gymController := controller.NewGymController(db)
	teamStatusController := controller.NewTeamStatusController(db)
	evaluationController := controller.NewEvaluationController(db)
	predictionController := controller.NewPredictionController(db)

	// 認証不要のルート
	e.GET("/debug/health", debugController.Health)
	e.GET("/debug/token", debugController.Token)

	// 認証必須のルートグループ
	api := e.Group("/api")
	api.Use(authmw.FirebaseAuth(fa))

	// ユーザー API
	api.GET("/users/me", userController.GetMe)
	api.POST("/users/me", userController.CreateMe)
	api.PUT("/users/me", userController.UpdateMe)

	// チーム API
	api.POST("/teams", teamController.CreateTeam)
	api.GET("/teams/me", teamController.GetMyTeam)
	api.GET("/teams/:teamId", teamController.GetTeam)

	// 招待コード API
	api.POST("/teams/:teamId/invite", inviteController.CreateInviteCode)
	api.POST("/teams/join", inviteController.JoinTeam)

	// 目標設定 API
	api.POST("/teams/:teamId/goal", goalController.CreateGoal)
	api.GET("/teams/:teamId/goal", goalController.GetGoal)
	api.PUT("/teams/:teamId/goal", goalController.UpdateGoal)

	// アクティビティ API（ランニング）
	api.POST("/activities/running/start", activityController.StartRunning)
	api.POST("/activities/running/:activityId/finish", activityController.FinishRunning)
	api.POST("/activities/running/:activityId/gps", activityController.SendGPSPoints)
	api.GET("/activities/running/:activityId", activityController.GetRunningActivity)

	// アクティビティ API（ジム）
	api.POST("/gym-locations", gymController.CreateGymLocation)
	api.GET("/gym-locations", gymController.GetGymLocations)
	api.DELETE("/gym-locations/:locationId", gymController.DeleteGymLocation)
	api.POST("/activities/gym/checkin", gymController.GymCheckin)
	api.POST("/activities/gym/:activityId/checkout", gymController.GymCheckout)
	api.GET("/activities/gym/:activityId", gymController.GetGymActivity)

	// アクティビティ API（共通）
	api.GET("/activities", activityController.GetMyActivities)
	api.GET("/teams/:teamId/activities", activityController.GetTeamActivities)

	// チーム HP・状態 API
	api.GET("/teams/:teamId/status", teamStatusController.GetTeamStatus)

	// 週次評価 API
	api.GET("/teams/:teamId/evaluations", evaluationController.GetEvaluations)
	api.GET("/teams/:teamId/evaluations/current", evaluationController.GetCurrentWeekEvaluation)

	// 失敗予測 API
	api.GET("/predictions/me", predictionController.GetMyPrediction)

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
