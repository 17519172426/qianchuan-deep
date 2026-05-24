package main

import (
	"fmt"
	"log"

	"github.com/example/qianchuan-saas/auth"
	"github.com/example/qianchuan-saas/config"
	"github.com/example/qianchuan-saas/db"
	"github.com/example/qianchuan-saas/models"
	"github.com/example/qianchuan-saas/qianchuan"
	"github.com/example/qianchuan-saas/router"
)

func main() {
	cfg := config.Load()

	auth.InitJWT(cfg.JWTSecret)
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

	qc := qianchuan.NewClient(cfg.QianchuanAppID, cfg.QianchuanSecret)

	r := router.Setup(qc)

	addr := fmt.Sprintf(":%s", cfg.ServerPort)
	log.Printf("千川投流助手 starting on %s", addr)
	log.Fatal(r.Run(addr))
}
