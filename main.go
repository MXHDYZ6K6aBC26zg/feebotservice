package main

import (
	//"net/http"
	"github.com/kenmobility/feezbot/router"
)

func main() {
	e := router.New()

	e.Logger.Fatal(e.Start(":8000"))
}


/* func check(c echo.Context) error {
	return c.String(http.StatusOK, "Hello, World!")
} */