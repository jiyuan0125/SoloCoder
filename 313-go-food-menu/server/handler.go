package server

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/example/food-menu/common"
)

type Handler struct {
	store       *Store
	persistence *Persistence
}

func NewHandler(store *Store, persistence *Persistence) *Handler {
	return &Handler{
		store:       store,
		persistence: persistence,
	}
}

func (h *Handler) jsonResponse(w http.ResponseWriter, success bool, message string, data interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	resp := common.APIResponse{
		Success: success,
		Message: message,
		Data:    data,
	}
	json.NewEncoder(w).Encode(resp)
}

func (h *Handler) decodeJSON(r *http.Request, v interface{}) error {
	return json.NewDecoder(r.Body).Decode(v)
}

func (h *Handler) CreateDish(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.jsonResponse(w, false, "方法不允许", nil)
		return
	}

	var req common.CreateDishRequest
	if err := h.decodeJSON(r, &req); err != nil {
		h.jsonResponse(w, false, "请求格式错误", nil)
		return
	}

	dish, err := h.store.CreateDish(&req)
	if err != nil {
		h.jsonResponse(w, false, err.Error(), nil)
		return
	}

	if err := h.persistence.SaveAll(); err != nil {
		h.jsonResponse(w, false, "保存数据失败", nil)
		return
	}

	h.jsonResponse(w, true, "创建成功", dish)
}

func (h *Handler) UpdateDish(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		h.jsonResponse(w, false, "方法不允许", nil)
		return
	}

	id := strings.TrimPrefix(r.URL.Path, "/api/dishes/")
	if id == "" {
		h.jsonResponse(w, false, "缺少菜品ID", nil)
		return
	}

	var req common.UpdateDishRequest
	if err := h.decodeJSON(r, &req); err != nil {
		h.jsonResponse(w, false, "请求格式错误", nil)
		return
	}

	dish, err := h.store.UpdateDish(id, &req)
	if err != nil {
		h.jsonResponse(w, false, err.Error(), nil)
		return
	}

	if err := h.persistence.SaveAll(); err != nil {
		h.jsonResponse(w, false, "保存数据失败", nil)
		return
	}

	h.jsonResponse(w, true, "更新成功", dish)
}

func (h *Handler) SetDishOnSale(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		h.jsonResponse(w, false, "方法不允许", nil)
		return
	}

	id := strings.TrimPrefix(strings.TrimSuffix(r.URL.Path, "/on_sale"), "/api/dishes/")
	if id == "" {
		h.jsonResponse(w, false, "缺少菜品ID", nil)
		return
	}

	err := h.store.SetDishStatus(id, common.StatusOnSale)
	if err != nil {
		h.jsonResponse(w, false, err.Error(), nil)
		return
	}

	if err := h.persistence.SaveAll(); err != nil {
		h.jsonResponse(w, false, "保存数据失败", nil)
		return
	}

	h.jsonResponse(w, true, "上架成功", nil)
}

func (h *Handler) SetDishOffSale(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		h.jsonResponse(w, false, "方法不允许", nil)
		return
	}

	id := strings.TrimPrefix(strings.TrimSuffix(r.URL.Path, "/off_sale"), "/api/dishes/")
	if id == "" {
		h.jsonResponse(w, false, "缺少菜品ID", nil)
		return
	}

	err := h.store.SetDishStatus(id, common.StatusOffSale)
	if err != nil {
		h.jsonResponse(w, false, err.Error(), nil)
		return
	}

	if err := h.persistence.SaveAll(); err != nil {
		h.jsonResponse(w, false, "保存数据失败", nil)
		return
	}

	h.jsonResponse(w, true, "下架成功", nil)
}

func (h *Handler) GetDish(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.jsonResponse(w, false, "方法不允许", nil)
		return
	}

	id := strings.TrimPrefix(r.URL.Path, "/api/dishes/")
	if id == "" {
		h.jsonResponse(w, false, "缺少菜品ID", nil)
		return
	}

	dish, err := h.store.GetDish(id)
	if err != nil {
		h.jsonResponse(w, false, err.Error(), nil)
		return
	}

	h.jsonResponse(w, true, "", dish)
}

func (h *Handler) ListDishes(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.jsonResponse(w, false, "方法不允许", nil)
		return
	}

	category := r.URL.Query().Get("category")
	onlyOnSale := r.URL.Query().Get("status") == "on_sale"

	dishes, err := h.store.ListDishesByCategory(category, onlyOnSale)
	if err != nil {
		h.jsonResponse(w, false, err.Error(), nil)
		return
	}

	resp := common.DishListResponse{
		Dishes: make([]common.Dish, 0, len(dishes)),
		Total:  len(dishes),
	}
	for _, d := range dishes {
		resp.Dishes = append(resp.Dishes, *d)
	}

	h.jsonResponse(w, true, "", resp)
}

func (h *Handler) SearchDishes(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.jsonResponse(w, false, "方法不允许", nil)
		return
	}

	category := r.URL.Query().Get("category")
	keyword := r.URL.Query().Get("keyword")

	dishes, err := h.store.SearchDishes(category, keyword)
	if err != nil {
		h.jsonResponse(w, false, err.Error(), nil)
		return
	}

	resp := common.DishListResponse{
		Dishes: make([]common.Dish, 0, len(dishes)),
		Total:  len(dishes),
	}
	for _, d := range dishes {
		resp.Dishes = append(resp.Dishes, *d)
	}

	h.jsonResponse(w, true, "", resp)
}

func (h *Handler) GetCategorySummary(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.jsonResponse(w, false, "方法不允许", nil)
		return
	}

	summaries := h.store.GetCategorySummary()
	resp := common.CategorySummaryResponse{
		Summaries: make([]common.CategorySummary, 0, len(summaries)),
	}
	for _, s := range summaries {
		resp.Summaries = append(resp.Summaries, *s)
	}

	h.jsonResponse(w, true, "", resp)
}

func (h *Handler) SetTodayRecommend(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.jsonResponse(w, false, "方法不允许", nil)
		return
	}

	var req common.SetRecommendRequest
	if err := h.decodeJSON(r, &req); err != nil {
		h.jsonResponse(w, false, "请求格式错误", nil)
		return
	}

	err := h.store.SetTodayRecommend(req.DishIDs)
	if err != nil {
		h.jsonResponse(w, false, err.Error(), nil)
		return
	}

	if err := h.persistence.SaveAll(); err != nil {
		h.jsonResponse(w, false, "保存数据失败", nil)
		return
	}

	info, dishes := h.store.GetTodayRecommend()
	resp := common.RecommendResponse{
		RecommendInfo: *info,
		Dishes:        make([]common.Dish, 0, len(dishes)),
	}
	for _, d := range dishes {
		resp.Dishes = append(resp.Dishes, *d)
	}

	h.jsonResponse(w, true, "设置推荐成功", resp)
}

func (h *Handler) GetTodayRecommend(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.jsonResponse(w, false, "方法不允许", nil)
		return
	}

	info, dishes := h.store.GetTodayRecommend()
	resp := common.RecommendResponse{
		RecommendInfo: *info,
		Dishes:        make([]common.Dish, 0, len(dishes)),
	}
	for _, d := range dishes {
		resp.Dishes = append(resp.Dishes, *d)
	}

	h.jsonResponse(w, true, "", resp)
}

func (h *Handler) GetCategories(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.jsonResponse(w, false, "方法不允许", nil)
		return
	}

	h.jsonResponse(w, true, "", common.ValidCategories)
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path

	if path == "/api/categories" {
		h.GetCategories(w, r)
		return
	}

	if path == "/api/dishes" {
		if r.Method == http.MethodGet {
			h.ListDishes(w, r)
		} else if r.Method == http.MethodPost {
			h.CreateDish(w, r)
		} else {
			h.jsonResponse(w, false, "方法不允许", nil)
		}
		return
	}

	if path == "/api/dishes/search" {
		h.SearchDishes(w, r)
		return
	}

	if path == "/api/categories/summary" {
		h.GetCategorySummary(w, r)
		return
	}

	if path == "/api/recommend" {
		if r.Method == http.MethodGet {
			h.GetTodayRecommend(w, r)
		} else if r.Method == http.MethodPost {
			h.SetTodayRecommend(w, r)
		} else {
			h.jsonResponse(w, false, "方法不允许", nil)
		}
		return
	}

	if strings.HasPrefix(path, "/api/dishes/") {
		remaining := strings.TrimPrefix(path, "/api/dishes/")
		if strings.HasSuffix(remaining, "/on_sale") {
			h.SetDishOnSale(w, r)
		} else if strings.HasSuffix(remaining, "/off_sale") {
			h.SetDishOffSale(w, r)
		} else {
			if r.Method == http.MethodGet {
				h.GetDish(w, r)
			} else if r.Method == http.MethodPut {
				h.UpdateDish(w, r)
			} else {
				h.jsonResponse(w, false, "方法不允许", nil)
			}
		}
		return
	}

	h.jsonResponse(w, false, "接口不存在", nil)
}
