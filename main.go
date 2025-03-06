package main

import (
	"log"
	"net/http"

	"github.com/FromZeroDev/loki_telegram_alert/server"
)

func main() {
	log.Println("Starting app. Listing on port 9089")
	if err := http.ListenAndServe(":9089", server.New()); err != nil {
		log.Fatal(err.Error())
	}
}
