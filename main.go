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
		log.Println("PORT must be set")
		port = "8000"
	}
	e.Logger.Fatal(e.Start(":"+port))
}


/* func check(c echo.Context) error {
	return c.String(http.StatusOK, "Hello, World!")
} */