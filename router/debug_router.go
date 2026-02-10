package router

import (
	"github.com/labstack/echo/v4"
	"github.com/trihackathon/api/adapter"
	"github.com/trihackathon/api/controller"
)

func DebugRouter(e *echo.Echo, fa *adapter.FirebaseAdapter) {
	ctrl := controller.NewDebugController(fa)
	e.GET("/debug/health", ctrl.Health)
	e.GET("/debug/endpoints", ctrl.Endpoints)
	e.POST("/debug/echo", ctrl.Echo)
}
