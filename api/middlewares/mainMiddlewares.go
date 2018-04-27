package middlewares

import (
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
)

func SetMainMiddlewares(e *echo.Echo) {
	e.Use(middleware.Logger())
	e.Use(middleware.CORS())
}