package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"search-service/models"
	"search-service/services"
)

func IndexDocument(c *gin.Context) {
	var req models.DocumentIndexRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request",
			"message": err.Error(),
		})
		return
	}

	doc, err := services.IndexDocument(req)
	if err != nil {
		if err.Error() == "title is required" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "Title is required",
				"message": "Title field cannot be empty",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to index document",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, doc)
}

func ListDocuments(c *gin.Context) {
	docs, err := services.ListDocuments()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to list documents",
			"message": err.Error(),
		})
		return
	}

	if docs == nil {
		docs = []*models.Document{}
	}

	c.JSON(http.StatusOK, docs)
}
