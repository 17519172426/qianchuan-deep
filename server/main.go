package main

import (
	"log"
	"net/http"
)

func main() {
	log.Println("千川投流助手 starting...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
