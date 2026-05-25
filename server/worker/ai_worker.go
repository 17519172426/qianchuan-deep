package worker

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/example/qianchuan-saas/db"
	rpc "github.com/example/qianchuan-saas/grpc"
	"github.com/example/qianchuan-saas/models"
	pb "github.com/example/qianchuan-saas/grpc/strategy"
)

type AIWorker struct {
	Grpc     *rpc.Client
	Interval time.Duration
}

func NewAIWorker(grpcClient *rpc.Client) *AIWorker {
	return &AIWorker{Grpc: grpcClient, Interval: 30 * time.Minute}
}

func (w *AIWorker) Start() {
	log.Printf("AI worker started, interval=%s", w.Interval)
	ticker := time.NewTicker(w.Interval)
	go func() {
		for range ticker.C {
			if w.Grpc == nil {
				continue
			}
			w.generateRecommendations()
		}
	}()
}

func (w *AIWorker) generateRecommendations() {
	var ads []models.UniAd
	db.DB.Where("qianchuan_ad_id IS NOT NULL").Find(&ads)
	if len(ads) == 0 {
		return
	}

	adIDs := make([]int64, 0, len(ads))
	currentMetrics := make([]*pb.AdMetrics, 0, len(ads))
	for i := range ads {
		adIDs = append(adIDs, int64(ads[i].ID))
		cost, roi, ctr, conversions, impressions, cpa := float64(0), float64(0), float64(0), int64(0), int64(0), float64(0)
		if ads[i].MetricsJSON != nil {
			metrics := map[string]interface{}(ads[i].MetricsJSON)
			if v, ok := metrics["cost"].(float64); ok { cost = v }
			if v, ok := metrics["roi"].(float64); ok { roi = v }
			if v, ok := metrics["ctr"].(float64); ok { ctr = v }
			if v, ok := metrics["conversions"].(float64); ok { conversions = int64(v) }
			if v, ok := metrics["impressions"].(float64); ok { impressions = int64(v) }
			if v, ok := metrics["cpa"].(float64); ok { cpa = v }
		}
		currentMetrics = append(currentMetrics, &pb.AdMetrics{
			AdId: int64(ads[i].ID), Cost: cost, Roi: roi,
			Ctr: ctr, Conversions: conversions, Impressions: impressions, Cpa: cpa,
		})
	}

	cutoff := time.Now().AddDate(0, 0, -7)
	var historyWindows []*pb.MetricsWindow
	for d := 0; d < 7; d++ {
		dayStart := cutoff.AddDate(0, 0, d)
		dayEnd := dayStart.AddDate(0, 0, 1)
		var reports []models.UniAdReport
		db.DB.Where("report_date >= ? AND report_date < ?", dayStart, dayEnd).Find(&reports)

		var hourly []*pb.AdMetrics
		for _, r := range reports {
			hourly = append(hourly, &pb.AdMetrics{
				AdId: int64(r.UniAdID), Cost: r.Cost, Roi: r.ROI,
				Ctr: r.CTR, Conversions: int64(r.Conversions),
				Impressions: r.Impressions, Cpa: float64(0),
			})
		}
		historyWindows = append(historyWindows, &pb.MetricsWindow{Hourly: hourly})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	recs, err := w.Grpc.GenerateRecommendations(ctx, adIDs, currentMetrics, historyWindows)
	if err != nil {
		log.Printf("AI recommendation generation failed: %v", err)
		return
	}

	for _, rec := range recs {
		suggestedAction := models.JSONMap{}
		if err := json.Unmarshal([]byte(rec.SuggestedActionJson), &suggestedAction); err != nil {
			suggestedAction = models.JSONMap{"type": "notify"}
		}

		db.DB.Create(&models.AIRecommendation{
			UniAdID:         uintPtr(uint(rec.AdId)),
			Type:            rec.Type,
			Title:           rec.Title,
			Description:     rec.Description,
			MetricsJSON:     models.JSONMap{},
			Confidence:      rec.Confidence,
			SuggestedAction: suggestedAction,
			Status:          "pending",
		})
	}
	log.Printf("AI worker generated %d recommendations for %d ads", len(recs), len(ads))
}

func uintPtr(u uint) *uint { return &u }
