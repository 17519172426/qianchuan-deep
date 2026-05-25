package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/example/qianchuan-saas/db"
	"github.com/example/qianchuan-saas/models"
)

type RuleHandler struct{}

func (h *RuleHandler) List(c *gin.Context) {
	accountID := c.Query("account_id")
	var rules []models.Rule
	q := db.DB
	if accountID != "" {
		q = q.Where("account_id = ?", accountID)
	}
	if err := q.Order("created_at DESC").Find(&rules).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch rules"})
		return
	}
	c.JSON(http.StatusOK, rules)
}

func (h *RuleHandler) Create(c *gin.Context) {
	var req struct {
		Name          string         `json:"name" binding:"required"`
		Description   string         `json:"description"`
		AccountID     uint           `json:"account_id" binding:"required"`
		ScopeJSON     models.JSONMap `json:"scope_json"`
		ConditionJSON models.JSONMap `json:"condition_json" binding:"required"`
		ActionJSON    models.JSONMap `json:"action_json" binding:"required"`
		Schedule      string         `json:"schedule"`
		Cooldown      string         `json:"cooldown"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	rule := models.Rule{
		Name:          req.Name,
		Description:   req.Description,
		AccountID:     req.AccountID,
		ScopeJSON:     req.ScopeJSON,
		ConditionJSON: req.ConditionJSON,
		ActionJSON:    req.ActionJSON,
		Schedule:      "*/5 * * * *",
		Cooldown:      "1h",
		Enabled:       false,
	}
	if req.Schedule != "" {
		rule.Schedule = req.Schedule
	}
	if req.Cooldown != "" {
		rule.Cooldown = req.Cooldown
	}
	if rule.ScopeJSON == nil {
		rule.ScopeJSON = models.JSONMap{}
	}
	if err := db.DB.Create(&rule).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create rule"})
		return
	}
	c.JSON(http.StatusCreated, rule)
}

func (h *RuleHandler) Get(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var rule models.Rule
	if err := db.DB.First(&rule, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "rule not found"})
		return
	}
	c.JSON(http.StatusOK, rule)
}

func (h *RuleHandler) Update(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var rule models.Rule
	if err := db.DB.First(&rule, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "rule not found"})
		return
	}
	var req struct {
		Name          *string         `json:"name"`
		Description   *string         `json:"description"`
		ScopeJSON     *models.JSONMap `json:"scope_json"`
		ConditionJSON *models.JSONMap `json:"condition_json"`
		ActionJSON    *models.JSONMap `json:"action_json"`
		Schedule      *string         `json:"schedule"`
		Cooldown      *string         `json:"cooldown"`
		Enabled       *bool           `json:"enabled"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	updates := map[string]interface{}{}
	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if req.Description != nil {
		updates["description"] = *req.Description
	}
	if req.ScopeJSON != nil {
		updates["scope_json"] = *req.ScopeJSON
	}
	if req.ConditionJSON != nil {
		updates["condition_json"] = *req.ConditionJSON
	}
	if req.ActionJSON != nil {
		updates["action_json"] = *req.ActionJSON
	}
	if req.Schedule != nil {
		updates["schedule"] = *req.Schedule
	}
	if req.Cooldown != nil {
		updates["cooldown"] = *req.Cooldown
	}
	if req.Enabled != nil {
		updates["enabled"] = *req.Enabled
	}
	if err := db.DB.Model(&rule).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update rule"})
		return
	}
	c.JSON(http.StatusOK, rule)
}

func (h *RuleHandler) Delete(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	if result := db.DB.Delete(&models.Rule{}, id); result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "rule not found"})
		return
	}
	c.JSON(http.StatusNoContent, nil)
}

func (h *RuleHandler) Executions(c *gin.Context) {
	ruleID := c.Query("rule_id")
	var executions []models.RuleExecution
	q := db.DB.Order("triggered_at DESC").Limit(100)
	if ruleID != "" {
		q = q.Where("rule_id = ?", ruleID)
	}
	if err := q.Find(&executions).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch executions"})
		return
	}
	c.JSON(http.StatusOK, executions)
}
