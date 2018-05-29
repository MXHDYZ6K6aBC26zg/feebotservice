package api

import (
	"github.com/labstack/echo"
	"github.com/kenmobility/feezbot/api/handlers"
)

func MainGroup(e *echo.Echo) {
	e.GET("/", handlers.Test)
	e.GET("/seedRoles", handlers.SeedTable)
	e.GET("/user/transaction/verify", handlers.VerifyTransaction)
}