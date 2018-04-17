package main

import (
	//"net/http"
	"github.com/kenmobility/feezbot/router"
	"os"
	"log"
)

func main() {
	e := router.New()
	port := os.Getenv("PORT")

	if port == "" {
		log.Fatal("$PORT must be set")
	}
	e.Logger.Fatal(e.Start(":"+port))
}


/* func check(c echo.Context) error {
	return c.String(http.StatusOK, "Hello, World!")
} */