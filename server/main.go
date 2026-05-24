package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/example/qianchuan-saas/config"
	"github.com/example/qianchuan-saas/db"
	"github.com/example/qianchuan-saas/models"
)

func main() {
	cfg := config.Load()

	db.Connect(cfg.DatabaseURL)
	db.AutoMigrate(
		&models.User{},
		&models.QianchuanAccount{},
		&models.UniAd{},
		&models.Creative{},
		&models.UniAdCreative{},
		&models.Rule{},
		&models.RuleExecution{},
		&models.AIRecommendation{},
	)
	log.Println("database connected and migrated")

	log.Printf("千川投流助手 starting on :%s", cfg.ServerPort)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", cfg.ServerPort), nil))
}
