package main

import (
	"fmt"
	"log"

	"github.com/example/qianchuan-saas/auth"
	"github.com/example/qianchuan-saas/config"
	"github.com/example/qianchuan-saas/db"
	rpc "github.com/example/qianchuan-saas/grpc"
	"github.com/example/qianchuan-saas/models"
	"github.com/example/qianchuan-saas/qianchuan"
	"github.com/example/qianchuan-saas/router"
	"github.com/example/qianchuan-saas/worker"
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
		&models.UniAdReport{},
	)
	db.DB.Exec("CREATE UNIQUE INDEX IF NOT EXISTS idx_uni_ad_reports_uniq ON uni_ad_reports (uni_ad_id, report_date, report_hour)")
	log.Println("database connected and migrated")

	qc := qianchuan.NewClient(cfg.QianchuanAppID, cfg.QianchuanSecret)

	syncWorker := worker.NewSyncWorker(qc)
	syncWorker.Start()

	grpcClient, grpcErr := rpc.NewClient("localhost:50051")
	if grpcErr != nil {
		log.Printf("WARNING: gRPC strategy service not available: %v", grpcErr)
	} else {
		ruleWorker := worker.NewRuleWorker(qc, grpcClient)
		ruleWorker.Start()
		aiWorker := worker.NewAIWorker(grpcClient)
		aiWorker.Start()
	}

	r := router.Setup(qc)

	addr := fmt.Sprintf(":%s", cfg.ServerPort)
	log.Printf("千川投流助手 starting on %s", addr)
	log.Fatal(r.Run(addr))
}
