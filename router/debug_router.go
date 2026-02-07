package router

import (
	"github.com/labstack/echo/v4"
	"github.com/trihackathon/api/controller"
	"go.uber.org/dig"
)

func DebugRouter(e *echo.Echo, container *dig.Container) {
	controller := controller.NewDebugController(container)
	e.GET("/debug/health", controller.Health)
	e.GET("/debug/endpoints", controller.Endpoints)
	e.POST("/debug/echo", controller.Echo)
}
