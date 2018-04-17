package router

import (
	"github.com/labstack/echo"
	"github.com/kenmobility/feezbot/api/middlewares"
	"github.com/kenmobility/feezbot/api"
)

func New() *echo.Echo {
	e := echo.New()

	// create groups
	//adminGroup := e.Group("/admin")
	mobileGroup := e.Group("/mobile")
	webGroup := e.Group("/web")

	// set all middlewares
	middlewares.SetMobileMiddlewares(mobileGroup)
	middlewares.SetWebMiddlewares(webGroup)
	middlewares.SetMainMiddlewares(e)

	// set main routes
	api.MainGroup(e)

	// set mobile group routes
	api.MobileGroup(mobileGroup)

	// set web group routes
	api.WebGroup(webGroup)

	return e
}