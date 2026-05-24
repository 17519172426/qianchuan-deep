package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/example/qianchuan-saas/config"
)

func main() {
	cfg := config.Load()
	log.Printf("千川投流助手 starting on :%s", cfg.ServerPort)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", cfg.ServerPort), nil))
}
