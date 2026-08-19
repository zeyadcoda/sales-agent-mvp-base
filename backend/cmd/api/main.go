package main

import (
	"log"
	"net/http"
	"os"

	"salesagent.local/backend/internal/httpapi"
)

func main() {
	addr := "127.0.0.1:8081"
	if value := os.Getenv("API_ADDR"); value != "" {
		addr = value
	}

	log.Printf("sales-agent API bootstrap listening on %s", addr)
	if err := http.ListenAndServe(addr, httpapi.NewRouter()); err != nil {
		log.Fatal(err)
	}
}
