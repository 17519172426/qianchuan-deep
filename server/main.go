package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/example/qianchuan-saas/config"
	"github.com/example/qianchuan-saas/db"
)

func main() {
	cfg := config.Load()

	db.Connect(cfg.DatabaseURL)
	log.Println("database connected")

	log.Printf("千川投流助手 starting on :%s", cfg.ServerPort)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", cfg.ServerPort), nil))
}
