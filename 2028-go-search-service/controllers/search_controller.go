package controllers

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"search-service/models"
	"search-service/services"
)

func Search(c *gin.Context) {
	keyword := strings.TrimSpace(c.Query("q"))
	if keyword == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Keyword is required",
			"message": "Search keyword cannot be empty",
		})
		return
	}

	field := strings.TrimSpace(c.Query("field"))

	results, err := services.Search(keyword, field)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Search failed",
			"message": err.Error(),
		})
		return
	}

	if results == nil {
		results = []*models.SearchResult{}
	}

	c.JSON(http.StatusOK, gin.H{
		"keyword": keyword,
		"field":   field,
		"count":   len(results),
		"results": results,
	})
}

func GetHotKeywords(c *gin.Context) {
	keywords, err := services.GetHotKeywords()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get hot keywords",
			"message": err.Error(),
		})
		return
	}

	if keywords == nil {
		keywords = []*models.HotKeyword{}
	}

	c.JSON(http.StatusOK, gin.H{
		"period":  "last_7_days",
		"count":   len(keywords),
		"results": keywords,
	})
}
