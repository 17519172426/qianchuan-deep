package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/example/qianchuan-saas/db"
	"github.com/example/qianchuan-saas/models"
)

type CreativeHandler struct{}

func (h *CreativeHandler) List(c *gin.Context) {
	accountID := c.Query("account_id")
	var creatives []models.Creative
	q := db.DB.Order("created_at DESC").Limit(50)
	if accountID != "" {
		q = q.Where("account_id = ?", accountID)
	}
	if err := q.Find(&creatives).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch creatives"})
		return
	}
	c.JSON(http.StatusOK, creatives)
}

func (h *CreativeHandler) Get(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var creative models.Creative
	if err := db.DB.First(&creative, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "creative not found"})
		return
	}
	c.JSON(http.StatusOK, creative)
}

func (h *CreativeHandler) UpdateTags(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var req struct {
		Tags models.JSONMap `json:"tags" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	var creative models.Creative
	if err := db.DB.First(&creative, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "creative not found"})
		return
	}
	if err := db.DB.Model(&creative).Update("tags", req.Tags).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update tags"})
		return
	}
	creative.Tags = req.Tags
	c.JSON(http.StatusOK, creative)
}
