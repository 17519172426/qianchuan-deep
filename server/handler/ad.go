package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"github.com/gin-gonic/gin"
	"github.com/example/qianchuan-saas/db"
	"github.com/example/qianchuan-saas/models"
	"github.com/example/qianchuan-saas/qianchuan"
)

type AdHandler struct {
	QC *qianchuan.Client
}

func getAdAccount(accountID uint) (*qianchuan.AdAccount, error) {
	var a models.QianchuanAccount
	if err := db.DB.First(&a, accountID).Error; err != nil {
		return nil, err
	}
	return &qianchuan.AdAccount{ID: a.ID, AdvertiserID: a.AdvertiserID}, nil
}

func (h *AdHandler) Create(c *gin.Context) {
	var req qianchuan.CreateAdRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	accountID, err := strconv.Atoi(c.Query("account_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid account_id"})
		return
	}
	acc, err := getAdAccount(uint(accountID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "account not found"})
		return
	}
	adID, err := h.QC.CreateUniAd(acc, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ad := models.UniAd{
		AccountID:       uint(accountID),
		QianchuanAdID:   &adID,
		Name:            req.Name,
		MarketingGoal:   req.MarketingGoal,
		DeliverySetting: models.JSONMap(req.DeliverySetting),
		CreativeSetting: models.JSONMap(req.CreativeSetting),
		Status:          "create",
	}
	if req.AwemeID != 0 {
		ad.AwemeID = &req.AwemeID
	}
	if len(req.ProductIDs) > 0 {
		b, _ := json.Marshal(req.ProductIDs)
		var ids models.JSONMap
		json.Unmarshal(b, &ids)
		ad.ProductIDs = ids
	}
	db.DB.Create(&ad)
	c.JSON(http.StatusCreated, ad)
}

func (h *AdHandler) List(c *gin.Context) {
	accountID := c.Query("account_id")
	var ads []models.UniAd
	q := db.DB.Preload("Account")
	if accountID != "" {
		q = q.Where("account_id = ?", accountID)
	}
	q.Order("created_at DESC").Find(&ads)
	c.JSON(http.StatusOK, ads)
}

var validAdStatuses = map[string]bool{
	"enable":  true,
	"disable": true,
	"delete":  true,
}

func (h *AdHandler) Get(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var ad models.UniAd
	if err := db.DB.Preload("Account").First(&ad, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "ad not found"})
		return
	}
	c.JSON(http.StatusOK, ad)
}

func (h *AdHandler) UpdateStatus(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var ad models.UniAd
	if err := db.DB.First(&ad, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "ad not found"})
		return
	}
	var req struct {
		Status string `json:"status" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if !validAdStatuses[req.Status] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid status, must be enable/disable/delete"})
		return
	}
	acc, err := getAdAccount(ad.AccountID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "account not found"})
		return
	}
	if ad.QianchuanAdID != nil {
		if err := h.QC.UpdateUniAdStatus(acc, []int64{*ad.QianchuanAdID}, req.Status); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}
	db.DB.Model(&ad).Update("status", req.Status)
	c.JSON(http.StatusOK, ad)
}
