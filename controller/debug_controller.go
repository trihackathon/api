package controller

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/trihackathon/api/requests"
	"github.com/trihackathon/api/response"
	"go.uber.org/dig"
)

type DebugController struct {
	container *dig.Container
}

func NewDebugController(container *dig.Container) *DebugController {
	return &DebugController{
		container: container,
	}
}

// @Summary Health check
// @Tags debug
// @Success 200 {object} map[string]string "OK! API is healthy"
// @Router /debug/health [get]
func (c *DebugController) Health(ctx echo.Context) error {
	return ctx.JSON(http.StatusOK, map[string]string{
		"status": "healthy",
	})
}

// @Summary List endpoints
// @Tags debug
// @Success 200 {object} map[string]interface{} "List of available endpoints"
// @Router /debug/endpoints [get]
func (c *DebugController) Endpoints(ctx echo.Context) error {
	return ctx.JSON(http.StatusOK, map[string]interface{}{
		"endpoints": []string{
			"GET /debug/health",
			"GET /debug/endpoints",
			"POST /debug/echo",
		},
	})
}

// @Summary Echo
// @Tags debug
// @Param message body requests.DebugEchoRequest true "メッセージ"
// @Success 200 {object} response.DebugEchoResponse "OK! Echo is healthy"
// @Router /debug/echo [post]
func (c *DebugController) Echo(ctx echo.Context) error {
	req := new(requests.DebugEchoRequest)
	if err := ctx.Bind(req); err != nil {
		return ctx.JSON(http.StatusBadRequest, map[string]interface{}{
			"message": "Failed to bind request",
		})
	}
	return ctx.JSON(http.StatusOK, &response.DebugEchoResponse{Message: req.Message})
}
