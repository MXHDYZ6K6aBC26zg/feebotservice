package middlewares

import (
	"github.com/labstack/echo"
	//"log"
	"net/http"
	h "github.com/kenmobility/feezbot/helper"
	"github.com/kenmobility/feezbot/api/handlers"
)

func SetMobileMiddlewares(g *echo.Group) {
	g.Use(AuthenticateRequests)
}

func AuthenticateRequests(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		headers := c.Request().Header
		nonce := headers["Nonce"]
		apiKey := headers["Api-Key"]
		apiSecret := headers["Api-Secret"]
		signature := headers["Signature"]
		if nonce == nil || apiKey == nil || apiSecret == nil || signature == nil {
			res := h.Response{
				Status: "error",
				Message:"None or Incomplete authentication details",
			}
			return c.JSON(http.StatusUnauthorized, res)
		}
		//log.Println("signature is ", signature[0], "api key is ", apiKey[0], "apiSecret is ", apiSecret[0])
		//Hash credentials and check if it matches the one sent
		if hCheck := handlers.CheckHash(nonce[0],apiKey[0], apiSecret[0], signature[0]); !hCheck {
			res := h.Response{
				Status: "error",
				Message:"Invalid signature hash",
			}
			return c.JSON(http.StatusBadRequest, res)
		}
		//check if signature matches the one on database
		if signCheck,msg := handlers.ValidateSignature(apiKey[0], apiSecret[0], signature[0]); !signCheck {
			res := h.Response {
				Status: "error",
				Message:msg,
			}
			return c.JSON(http.StatusBadRequest, res)
		}
		return next(c)
	}
}