package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/example/qianchuan-saas/db"
	"github.com/example/qianchuan-saas/models"
)

type DashboardHandler struct{}

type DashboardStats struct {
	TodayCost        float64 `json:"today_cost"`
	AvgROI           float64 `json:"avg_roi"`
	TotalConversions int64   `json:"total_conversions"`
	ActiveAds        int64   `json:"active_ads"`
	TotalAccounts    int64   `json:"total_accounts"`
}

func (h *DashboardHandler) Stats(c *gin.Context) {
	today := time.Now().Format("2006-01-02")
	var stats DashboardStats

	db.DB.Model(&models.UniAdReport{}).
		Where("report_date = ?", today).
		Select("COALESCE(SUM(cost), 0) as today_cost").
		Scan(&stats.TodayCost)

	db.DB.Model(&models.UniAdReport{}).
		Where("report_date = ?", today).
		Select("COALESCE(AVG(roi), 0) as avg_roi").
		Scan(&stats.AvgROI)

	db.DB.Model(&models.UniAdReport{}).
		Where("report_date = ?", today).
		Select("COALESCE(SUM(conversions), 0) as total_conversions").
		Scan(&stats.TotalConversions)

	db.DB.Model(&models.UniAd{}).
		Where("status = ?", "enable").
		Count(&stats.ActiveAds)

	db.DB.Model(&models.QianchuanAccount{}).Count(&stats.TotalAccounts)

	c.JSON(http.StatusOK, stats)
}

type TrendPoint struct {
	Date string  `json:"date"`
	Cost float64 `json:"cost"`
	ROI  float64 `json:"roi"`
}

func (h *DashboardHandler) Trend(c *gin.Context) {
	var points []TrendPoint
	db.DB.Model(&models.UniAdReport{}).
		Select("report_date::text as date, SUM(cost) as cost, AVG(roi) as roi").
		Where("report_date >= CURRENT_DATE - INTERVAL '7 days'").
		Group("report_date").
		Order("report_date ASC").
		Scan(&points)
	c.JSON(http.StatusOK, points)
}
