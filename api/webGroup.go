package api

import (
	"github.com/labstack/echo"
	"github.com/kenmobility/feezbot/api/handlers"
)

func WebGroup(g *echo.Group) {
	//*******HANDLER FOR CREATING SUBACCOUNT***********//
	g.POST("/create/subaccount", handlers.CreateSubAccount)

	//*******HANDLER FOR SENDING MAIL FROM WEB***********//
	g.POST("/send/mail", handlers.SendMailFromWeb)
}