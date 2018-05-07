package api

import (
	"github.com/labstack/echo"
	"github.com/kenmobility/feezbot/api/handlers"
)

func MobileGroup(g *echo.Group) {
	//g.GET("/check", handlers.CheckPassword)

	//*******HANDLERS FOR USER LOGIN AND USER CREATION***********//
	g.POST("/login", handlers.Login)
	g.POST("/user/create", handlers.CreateUser)

	//*******HANDLERS FOR TRANSACTIONS***********//
	g.POST("/user/transaction/initiate", handlers.InitiateTransaction)
	//g.POST("/user/pay/card", handlers.ChargeUserByCard)
	//g.POST("/user/pay/bank", handlers.ChargeUserByBank)

	//*******HANDLERS FOR USER VALIDATION PROCCESSING DURING TRANSACTIONS***********//
	//g.POST("/user/pin/proccess", handlers.ProccessPin)
	//g.POST("/user/otp/proccess", handlers.ProccessOtp)
	//g.POST("/user/phone/proccess", )
	//g.POST("/user/birthday/proccess", )

	//*******GET HANDLERS FOR MERCHANTS***********//
	g.GET("/getMerchants", handlers.ShowMerchantsSummary)
	g.GET("/getMerchantFees", handlers.ShowMerchantFees)

	//*******GET HANDLER FOR CATEGORIES***********//
	g.GET("/getCategories", handlers.ShowCategories)

	//******POST HANDLER FOR DEVICE COORDINATES******//
	g.POST("/user/insert/coordinates", handlers.UserDeviceCoordinate)
}
