package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"points-mall/models"
	"points-mall/repository"
)

func ListProducts(w http.ResponseWriter, r *http.Request) {
	products, err := repository.GetOnlineProducts()
	if err != nil {
		RespondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	RespondJSON(w, http.StatusOK, products)
}

func GetProduct(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimPrefix(r.URL.Path, "/api/products/")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		RespondError(w, http.StatusBadRequest, "无效的商品ID")
		return
	}

	product, err := repository.GetProductByID(id)
	if err != nil {
		RespondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if product == nil {
		RespondError(w, http.StatusNotFound, "商品不存在")
		return
	}

	RespondJSON(w, http.StatusOK, product)
}

func GetProductExchanges(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/products/")
	parts := strings.Split(path, "/")
	if len(parts) < 2 {
		RespondError(w, http.StatusNotFound, "路由不存在")
		return
	}

	productID, err := strconv.Atoi(parts[0])
	if err != nil {
		RespondError(w, http.StatusBadRequest, "无效的商品ID")
		return
	}

	product, err := repository.GetProductByID(productID)
	if err != nil {
		RespondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if product == nil {
		RespondError(w, http.StatusNotFound, "商品不存在")
		return
	}

	userIDStr := r.URL.Query().Get("user_id")
	if userIDStr == "" {
		RespondError(w, http.StatusBadRequest, "缺少user_id参数")
		return
	}

	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		RespondError(w, http.StatusBadRequest, "无效的用户ID")
		return
	}

	exchanges, err := repository.GetUserExchanges(userID)
	if err != nil {
		RespondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	var productExchanges []*models.Exchange
	for _, e := range exchanges {
		if e.ProductID == productID {
			productExchanges = append(productExchanges, e)
		}
	}

	RespondJSON(w, http.StatusOK, productExchanges)
}

func UpdateProductStatus(w http.ResponseWriter, r *http.Request) {
	var req struct {
		IsOnline bool `json:"is_online"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondError(w, http.StatusBadRequest, "无效的请求体")
		return
	}

	path := strings.TrimPrefix(r.URL.Path, "/api/admin/products/")
	parts := strings.Split(path, "/")
	if len(parts) < 2 || parts[1] != "status" {
		RespondError(w, http.StatusNotFound, "路由不存在")
		return
	}

	id, err := strconv.Atoi(parts[0])
	if err != nil {
		RespondError(w, http.StatusBadRequest, "无效的商品ID")
		return
	}

	product, err := repository.GetProductByID(id)
	if err != nil {
		RespondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if product == nil {
		RespondError(w, http.StatusNotFound, "商品不存在")
		return
	}

	if err := repository.UpdateProductStatus(id, req.IsOnline); err != nil {
		RespondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	product, _ = repository.GetProductByID(id)
	RespondJSON(w, http.StatusOK, product)
}

func UpdateProductPoints(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Points int `json:"points"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondError(w, http.StatusBadRequest, "无效的请求体")
		return
	}

	path := strings.TrimPrefix(r.URL.Path, "/api/admin/products/")
	parts := strings.Split(path, "/")
	if len(parts) < 2 || parts[1] != "points" {
		RespondError(w, http.StatusNotFound, "路由不存在")
		return
	}

	id, err := strconv.Atoi(parts[0])
	if err != nil {
		RespondError(w, http.StatusBadRequest, "无效的商品ID")
		return
	}

	product, err := repository.GetProductByID(id)
	if err != nil {
		RespondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if product == nil {
		RespondError(w, http.StatusNotFound, "商品不存在")
		return
	}

	if req.Points < 0 {
		RespondError(w, http.StatusBadRequest, "积分不能为负数")
		return
	}

	if err := repository.UpdateProductPoints(id, req.Points); err != nil {
		RespondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	product, _ = repository.GetProductByID(id)
	RespondJSON(w, http.StatusOK, product)
}

func CreateProduct(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name       string `json:"name"`
		Points     int    `json:"points"`
		Stock      int    `json:"stock"`
		DailyLimit int    `json:"daily_limit"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondError(w, http.StatusBadRequest, "无效的请求体")
		return
	}

	if req.Name == "" {
		RespondError(w, http.StatusBadRequest, "商品名称不能为空")
		return
	}
	if req.Points < 0 {
		RespondError(w, http.StatusBadRequest, "积分不能为负数")
		return
	}
	if req.Stock < 0 {
		RespondError(w, http.StatusBadRequest, "库存不能为负数")
		return
	}
	if req.DailyLimit < 1 {
		req.DailyLimit = 1
	}

	product, err := repository.CreateProduct(req.Name, req.Points, req.Stock, req.DailyLimit)
	if err != nil {
		RespondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	RespondJSON(w, http.StatusCreated, product)
}

func ListAdminProducts(w http.ResponseWriter, r *http.Request) {
	products, err := repository.GetAllProducts()
	if err != nil {
		RespondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	RespondJSON(w, http.StatusOK, products)
}

func CreateUser(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name   string `json:"name"`
		Points int    `json:"points"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondError(w, http.StatusBadRequest, "无效的请求体")
		return
	}

	if req.Name == "" {
		RespondError(w, http.StatusBadRequest, "用户名不能为空")
		return
	}

	user, err := repository.CreateUser(req.Name, req.Points)
	if err != nil {
		RespondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	RespondJSON(w, http.StatusCreated, user)
}

func GetUser(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimPrefix(r.URL.Path, "/api/users/")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		RespondError(w, http.StatusBadRequest, "无效的用户ID")
		return
	}

	user, err := repository.GetUserByID(id)
	if err != nil {
		RespondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if user == nil {
		RespondError(w, http.StatusNotFound, "用户不存在")
		return
	}

	RespondJSON(w, http.StatusOK, user)
}
