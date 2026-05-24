package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/example/qianchuan-saas/db"
	"github.com/example/qianchuan-saas/models"
)

type ReportHandler struct{}

func (h *ReportHandler) ByAd(c *gin.Context) {
	adID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid ad id"})
		return
	}

	days := c.DefaultQuery("days", "7")

	var reports []models.UniAdReport
	db.DB.Where("uni_ad_id = ? AND report_date >= CURRENT_DATE - INTERVAL '1 day' * ?::int", adID, days).
		Order("report_date DESC, report_hour DESC").
		Find(&reports)
	c.JSON(http.StatusOK, reports)
}

func (h *ReportHandler) SummaryByDate(c *gin.Context) {
	accountID := c.Query("account_id")
	startDate := c.DefaultQuery("start_date", time.Now().AddDate(0, 0, -7).Format("2006-01-02"))
	endDate := c.DefaultQuery("end_date", time.Now().Format("2006-01-02"))

	type DailySummary struct {
		Date        string  `json:"date"`
		Cost        float64 `json:"cost"`
		Impressions int64   `json:"impressions"`
		Clicks      int64   `json:"clicks"`
		Conversions int     `json:"conversions"`
		ROI         float64 `json:"roi"`
	}

	var summaries []DailySummary
	q := db.DB.Model(&models.UniAdReport{}).
		Select("report_date::text as date, SUM(cost) as cost, SUM(impressions) as impressions, SUM(clicks) as clicks, SUM(conversions) as conversions, AVG(roi) as roi").
		Where("report_date BETWEEN ? AND ?", startDate, endDate)

	if accountID != "" {
		q = q.Where("uni_ad_id IN (SELECT id FROM uni_ads WHERE account_id = ?)", accountID)
	}

	q.Group("report_date").Order("report_date ASC").Scan(&summaries)
	c.JSON(http.StatusOK, summaries)
}
