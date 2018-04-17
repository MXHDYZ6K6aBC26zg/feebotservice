package middlewares

import (
	"github.com/labstack/echo"
)

func SetWebMiddlewares(g *echo.Group) {
	g.Use(AuthenticateRequests)
}

