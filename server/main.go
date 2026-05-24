package main

import (
	"log"
	"net/http"
	"os"
)

func main() {
	log.Println("千川投流助手 starting...")
	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "8080"
	}
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
