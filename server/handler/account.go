package handler

import (
	"net/http"
	"strconv"
	"github.com/gin-gonic/gin"
	"github.com/example/qianchuan-saas/db"
	"github.com/example/qianchuan-saas/models"
	"github.com/example/qianchuan-saas/qianchuan"
)

type AccountHandler struct {
	QC *qianchuan.Client
}

func (h *AccountHandler) List(c *gin.Context) {
	var accounts []models.QianchuanAccount
	db.DB.Find(&accounts)
	c.JSON(http.StatusOK, accounts)
}

func (h *AccountHandler) Create(c *gin.Context) {
	var req struct {
		AccountName  string `json:"account_name" binding:"required"`
		AdvertiserID int64  `json:"advertiser_id" binding:"required"`
		AuthCode     string `json:"auth_code" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	tr, err := h.QC.OAuth.GetToken(req.AuthCode)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "get token failed: " + err.Error()})
		return
	}
	account := models.QianchuanAccount{
		AccountName:  req.AccountName,
		AdvertiserID: req.AdvertiserID,
		AccessToken:  tr.AccessToken,
		RefreshToken: tr.RefreshToken,
	}
	db.DB.Create(&account)
	c.JSON(http.StatusCreated, account)
}

func (h *AccountHandler) Get(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var account models.QianchuanAccount
	if err := db.DB.First(&account, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "account not found"})
		return
	}
	c.JSON(http.StatusOK, account)
}

func (h *AccountHandler) Delete(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	if result := db.DB.Delete(&models.QianchuanAccount{}, id); result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "account not found"})
		return
	}
	c.JSON(http.StatusNoContent, nil)
}
