package main

import (
	"net/http"

	"github.com/labstack/echo/v4"
	echoSwagger "github.com/swaggo/echo-swagger"
	
	_ "github.com/trihackathon/api/docs" // Swagger docs
)

func main() {
	e := echo.New()

	// Swagger UI
	e.GET("/swagger/*", echoSwagger.WrapHandler)

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

	e.Start(":8080")
}
