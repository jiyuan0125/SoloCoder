package handlers

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strconv"

	"json-validator/memory"
	"json-validator/storage"
	"json-validator/validator"

	"github.com/gin-gonic/gin"
)

type ValidateRequest struct {
	Schema interface{} `json:"schema" binding:"required"`
	Data   interface{} `json:"data" binding:"required"`
}

type ValidateResponse struct {
	Valid  bool               `json:"valid"`
	Errors []ValidationError  `json:"errors,omitempty"`
	ID     int64              `json:"id,omitempty"`
}

type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

func ValidateHandler(c *gin.Context) {
	monitor := memory.GetMonitor()
	if monitor.IsOverLimit() {
		c.JSON(http.StatusRequestEntityTooLarge, gin.H{
			"error": "内存使用超过限制，暂时无法处理请求",
			"code":  413,
		})
		return
	}

	var rawRequest map[string]interface{}
	if err := c.ShouldBindJSON(&rawRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "请求体格式错误: " + err.Error(),
			"code":  400,
		})
		return
	}

	schemaRaw, schemaOk := rawRequest["schema"]
	dataRaw, dataOk := rawRequest["data"]

	if !schemaOk || !dataOk {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "请求必须包含 schema 和 data 字段",
			"code":  400,
		})
		return
	}

	schemaJSON, err := marshalToJSONBytes(schemaRaw)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "schema 格式错误: " + err.Error(),
			"code":  400,
		})
		return
	}

	dataJSON, err := marshalToJSONBytes(dataRaw)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "data 格式错误: " + err.Error(),
			"code":  400,
		})
		return
	}

	v, err := validator.NewStreamValidator(schemaJSON)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
			"code":  400,
		})
		return
	}

	result, err := v.ValidateStream(io.NopCloser(bytes.NewReader(dataJSON)))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
			"code":  400,
		})
		return
	}

	response := ValidateResponse{
		Valid:  result.Valid,
		Errors: make([]ValidationError, len(result.Errors)),
	}

	for i, e := range result.Errors {
		response.Errors[i] = ValidationError{
			Field:   e.Field,
			Message: e.Message,
		}
	}

	store := storage.GetStore()
	if store != nil {
		id, saveErr := store.SaveValidation(string(schemaJSON), string(dataJSON), result)
		if saveErr == nil {
			response.ID = id
		}
	}

	c.JSON(http.StatusOK, response)
}

func GetValidationHandler(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "无效的 ID",
			"code":  400,
		})
		return
	}

	store := storage.GetStore()
	if store == nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "数据库未初始化",
			"code":  500,
		})
		return
	}

	record, err := store.GetValidation(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "记录不存在",
			"code":  404,
		})
		return
	}

	c.JSON(http.StatusOK, record)
}

func marshalToJSONBytes(v interface{}) ([]byte, error) {
	jsonBytes, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	return jsonBytes, nil
}
