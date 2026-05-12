package router

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/gin-gonic/gin"

	"dataconverter/config"
	"dataconverter/engine"
)

type ConvertRequest struct {
	Rule string                 `json:"rule" binding:"required"`
	Data map[string]interface{} `json:"data" binding:"required"`
}

type RuleRequest struct {
	Rules []config.Rule `json:"rules" binding:"required"`
}

func SetupRouter(loader *engine.RuleLoader) *gin.Engine {
	r := gin.Default()
	
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "ok",
			"message": "service is running",
		})
	})
	
	r.GET("/rules", func(c *gin.Context) {
		ruleNames := loader.GetEngine().GetRuleNames()
		c.JSON(http.StatusOK, gin.H{
			"rules": ruleNames,
			"count": len(ruleNames),
		})
	})
	
	r.POST("/rules/reload", func(c *gin.Context) {
		body, err := ioutil.ReadAll(c.Request.Body)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		
		if err := loader.LoadFromJSON(body); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		
		c.JSON(http.StatusOK, gin.H{"message": "rules reloaded successfully"})
	})
	
	r.POST("/convert", func(c *gin.Context) {
		body, err := ioutil.ReadAll(c.Request.Body)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   "failed to read request body",
			})
			return
		}
		
		var req struct {
			Rule string                 `json:"rule"`
			Data map[string]interface{} `json:"data"`
		}
		
		if err := json.Unmarshal(body, &req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   "invalid JSON format: " + err.Error(),
			})
			return
		}
		
		if req.Rule == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   "rule is required",
			})
			return
		}
		
		if req.Data == nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   "data is required",
			})
			return
		}
		
		if !loader.GetEngine().HasRule(req.Rule) {
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"error":   "rule not found: " + req.Rule,
			})
			return
		}
		
		dataBytes, _ := json.Marshal(req.Data)
		result, err := loader.GetEngine().Convert(dataBytes, req.Rule)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"error":   err.Error(),
			})
			return
		}
		
		c.JSON(http.StatusOK, result)
	})
	
	return r
}
