package worker

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/example/qianchuan-saas/db"
	rpc "github.com/example/qianchuan-saas/grpc"
	"github.com/example/qianchuan-saas/models"
	"github.com/example/qianchuan-saas/qianchuan"
	pb "github.com/example/qianchuan-saas/grpc/strategy"
)

type RuleWorker struct {
	QC       *qianchuan.Client
	Grpc     *rpc.Client
	Interval time.Duration
}

func NewRuleWorker(qc *qianchuan.Client, grpcClient *rpc.Client) *RuleWorker {
	return &RuleWorker{QC: qc, Grpc: grpcClient, Interval: 5 * time.Minute}
}

func (w *RuleWorker) Start() {
	log.Printf("rule worker started, interval=%s", w.Interval)
	ticker := time.NewTicker(w.Interval)
	go func() {
		for range ticker.C {
			if w.Grpc == nil {
				continue
			}
			w.evaluateAndExecute()
		}
	}()
}

func (w *RuleWorker) evaluateAndExecute() {
	var rules []models.Rule
	db.DB.Where("enabled = ?", true).Find(&rules)
	if len(rules) == 0 {
		return
	}

	var ads []models.UniAd
	db.DB.Where("qianchuan_ad_id IS NOT NULL").Find(&ads)

	adMap := make(map[uint]*models.UniAd)
	var adContexts []*pb.AdContext
	var ruleDefs []*pb.RuleDef
	now := time.Now()

	for i := range ads {
		adMap[ads[i].ID] = &ads[i]
		cost := float64(0)
		roi := float64(0)
		ctr := float64(0)
		conversions := int64(0)
		impressions := int64(0)
		if ads[i].MetricsJSON != nil {
			metrics := map[string]interface{}(ads[i].MetricsJSON)
			if v, ok := metrics["cost"].(float64); ok {
				cost = v
			}
			if v, ok := metrics["roi"].(float64); ok {
				roi = v
			}
			if v, ok := metrics["ctr"].(float64); ok {
				ctr = v
			}
			if v, ok := metrics["conversions"].(float64); ok {
				conversions = int64(v)
			}
			if v, ok := metrics["impressions"].(float64); ok {
				impressions = int64(v)
			}
		}
		qianchuanAdID := int64(0)
		if ads[i].QianchuanAdID != nil {
			qianchuanAdID = *ads[i].QianchuanAdID
		}
		adContexts = append(adContexts, &pb.AdContext{
			AdId:          int64(ads[i].ID),
			QianchuanAdId: qianchuanAdID,
			Cost:          cost,
			Roi:           roi,
			Ctr:           ctr,
			Conversions:   conversions,
			Impressions:   impressions,
		})
	}

	for i := range rules {
		rule := &rules[i]
		condBytes, err := json.Marshal(rule.ConditionJSON)
		if err != nil {
			log.Printf("marshal condition_json failed for rule %d: %v", rule.ID, err)
			continue
		}
		actionBytes, err := json.Marshal(rule.ActionJSON)
		if err != nil {
			log.Printf("marshal action_json failed for rule %d: %v", rule.ID, err)
			continue
		}
		ruleDefs = append(ruleDefs, &pb.RuleDef{
			Id:            int64(rule.ID),
			Name:          rule.Name,
			AccountId:     int64(rule.AccountID),
			ConditionJson: string(condBytes),
			ActionJson:    string(actionBytes),
			Cooldown:      rule.Cooldown,
		})
	}

	if len(ruleDefs) == 0 {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	actions, err := w.Grpc.EvaluateRules(ctx, ruleDefs, adContexts)
	if err != nil {
		log.Printf("rule evaluation failed: %v", err)
		return
	}

	for _, action := range actions {
		ad, ok := adMap[uint(action.AdId)]
		if !ok {
			continue
		}

		safe := w.applySafetyLimits(action)
		if !safe {
			log.Printf("safety limit blocked rule %d action on ad %d: %s", action.RuleId, action.AdId, action.ActionType)
			continue
		}

		var acc models.QianchuanAccount
		if err := db.DB.First(&acc, ad.AccountID).Error; err != nil {
			log.Printf("account not found for ad %d", ad.ID)
			continue
		}
		accRef := qianchuan.AdAccount{ID: acc.ID, AdvertiserID: acc.AdvertiserID}

		execStatus := "success"
		var execErr error
		switch action.ActionType {
		case "pause_ad":
			execErr = w.QC.UpdateUniAdStatus(&accRef, []int64{*ad.QianchuanAdID}, "disable")
		case "resume_ad":
			execErr = w.QC.UpdateUniAdStatus(&accRef, []int64{*ad.QianchuanAdID}, "enable")
		case "update_budget":
			newBudget := action.Value
			if action.ValueType == "percentage" {
				currentBudget := float64(0)
				if ad.DeliverySetting != nil {
					if b, ok := ad.DeliverySetting["budget"].(float64); ok {
						currentBudget = b
					}
				}
				newBudget = currentBudget * (1 + action.Value)
			}
			req := &qianchuan.UpdateAdRequest{
				AdvertiserID:    accRef.AdvertiserID,
				AdID:            *ad.QianchuanAdID,
				DeliverySetting: map[string]interface{}{"budget": newBudget},
			}
			execErr = w.QC.UpdateUniAd(&accRef, req)
			if execErr == nil {
				if ad.DeliverySetting == nil {
					ad.DeliverySetting = models.JSONMap{}
				}
				ad.DeliverySetting["budget"] = newBudget
				db.DB.Model(ad).Update("delivery_setting", ad.DeliverySetting)
				execStatus = "success"
			} else {
				execStatus = "failed"
			}
		case "update_roi_goal":
			newROI := action.Value
			if action.ValueType == "percentage" {
				currentROI := float64(0)
				if ad.DeliverySetting != nil {
					if r, ok := ad.DeliverySetting["roi_goal"].(float64); ok {
						currentROI = r
					}
				}
				newROI = currentROI * (1 + action.Value)
			}
			req := &qianchuan.UpdateAdRequest{
				AdvertiserID:    accRef.AdvertiserID,
				AdID:            *ad.QianchuanAdID,
				DeliverySetting: map[string]interface{}{"roi_goal": newROI},
			}
			execErr = w.QC.UpdateUniAd(&accRef, req)
			if execErr == nil {
				if ad.DeliverySetting == nil {
					ad.DeliverySetting = models.JSONMap{}
				}
				ad.DeliverySetting["roi_goal"] = newROI
				db.DB.Model(ad).Update("delivery_setting", ad.DeliverySetting)
				execStatus = "success"
			} else {
				execStatus = "failed"
			}
		case "raise_ad":
			newBid := action.Value
			if action.ValueType == "percentage" {
				currentBid := float64(0)
				if ad.DeliverySetting != nil {
					if b, ok := ad.DeliverySetting["bid"].(float64); ok {
						currentBid = b
					}
				}
				newBid = currentBid * (1 + action.Value)
			}
			req := &qianchuan.UpdateAdRequest{
				AdvertiserID:    accRef.AdvertiserID,
				AdID:            *ad.QianchuanAdID,
				DeliverySetting: map[string]interface{}{"bid": newBid},
			}
			execErr = w.QC.UpdateUniAd(&accRef, req)
			if execErr == nil {
				if ad.DeliverySetting == nil {
					ad.DeliverySetting = models.JSONMap{}
				}
				ad.DeliverySetting["bid"] = newBid
				db.DB.Model(ad).Update("delivery_setting", ad.DeliverySetting)
				execStatus = "success"
			} else {
				execStatus = "failed"
			}
		case "notify":
			log.Printf("notify action for rule %d ad %d", action.RuleId, action.AdId)
		default:
			execStatus = "skipped"
			log.Printf("unknown action %s for rule %d ad %d", action.ActionType, action.RuleId, action.AdId)
		}

		if execErr != nil {
			execStatus = "failed"
		}

		execution := models.RuleExecution{
			RuleID:        uint(action.RuleId),
			UniAdID:       uint(action.AdId),
			TriggeredAt:   &now,
			ConditionJSON: models.JSONMap{},
			ActionJSON:    models.JSONMap{"type": action.ActionType, "value": action.Value},
			Status:        execStatus,
		}
		if execErr != nil {
			execution.ResultJSON = models.JSONMap{"error": "external API call failed"}
			log.Printf("action execution failed: rule=%d ad=%d action=%s err=%v",
				action.RuleId, action.AdId, action.ActionType, execErr)
		}
		if execStatus == "skipped" {
			execution.ResultJSON = models.JSONMap{"info": "action not yet implemented"}
		}
		if err := db.DB.Create(&execution).Error; err != nil {
			log.Printf("failed to create rule execution record: %v", err)
		}
	}
}

func (w *RuleWorker) applySafetyLimits(action *pb.RuleAction) bool {
	switch action.ActionType {
	case "update_budget":
		if action.ValueType == "percentage" {
			return action.Value >= -0.5 && action.Value <= 1.0
		}
		if action.Value < 300 {
			return false
		}
	case "update_roi_goal":
		if action.ValueType == "percentage" {
			return action.Value >= -0.5 && action.Value <= 0.5
		}
		return action.Value > 0.01 && action.Value < 100
	case "raise_ad":
		if action.ValueType == "percentage" {
			return action.Value >= -0.3 && action.Value <= 1.0
		}
		return action.Value > 0
	case "pause_ad", "resume_ad", "notify":
		return true
	}
	return true
}
