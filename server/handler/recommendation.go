package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/example/qianchuan-saas/db"
	"github.com/example/qianchuan-saas/models"
)

type RecommendationHandler struct{}

func (h *RecommendationHandler) List(c *gin.Context) {
	status := c.DefaultQuery("status", "pending")
	var recs []models.AIRecommendation
	q := db.DB.Order("created_at DESC").Limit(50)
	if status != "" {
		q = q.Where("status = ?", status)
	}
	if err := q.Find(&recs).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch recommendations"})
		return
	}
	c.JSON(http.StatusOK, recs)
}

func (h *RecommendationHandler) UpdateStatus(c *gin.Context) {
	id := c.Param("id")
	var req struct {
		Status string `json:"status" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.Status != "accepted" && req.Status != "ignored" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "status must be accepted or ignored"})
		return
	}
	var rec models.AIRecommendation
	if err := db.DB.First(&rec, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "recommendation not found"})
		return
	}
	if err := db.DB.Model(&rec).Update("status", req.Status).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update"})
		return
	}
	rec.Status = req.Status
	c.JSON(http.StatusOK, rec)
}
