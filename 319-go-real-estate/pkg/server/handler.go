package server

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/google/uuid"

	"realestate/pkg/common"
)

type Handler struct {
	store *Store
}

func NewHandler(store *Store) *Handler {
	return &Handler{store: store}
}

func (h *Handler) writeJSON(w http.ResponseWriter, status int, resp common.Response) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(resp)
}

func (h *Handler) writeError(w http.ResponseWriter, status int, err error) {
	h.writeJSON(w, status, common.Response{
		Success: false,
		Message: err.Error(),
	})
}

func (h *Handler) CreateProperty(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.writeError(w, http.StatusMethodNotAllowed, ErrorInvalidInput)
		return
	}

	var req struct {
		LandlordID  string             `json:"landlord_id"`
		Community   string             `json:"community"`
		HouseType   string             `json:"house_type"`
		Area        float64            `json:"area"`
		Floor       int                `json:"floor"`
		Orientation string             `json:"orientation"`
		PriceType   common.PriceType   `json:"price_type"`
		Price       float64            `json:"price"`
		Contact     string             `json:"contact"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, ErrorInvalidInput)
		return
	}

	if strings.TrimSpace(req.LandlordID) == "" {
		h.writeError(w, http.StatusBadRequest, ErrorInvalidInput)
		return
	}

	if strings.TrimSpace(req.Community) == "" {
		h.writeError(w, http.StatusBadRequest, ErrorInvalidInput)
		return
	}

	if err := common.ValidateHouseType(req.HouseType); err != nil {
		h.writeError(w, http.StatusBadRequest, err)
		return
	}

	if req.Area <= 0 {
		h.writeError(w, http.StatusBadRequest, ErrorInvalidInput)
		return
	}

	if req.Floor <= 0 {
		h.writeError(w, http.StatusBadRequest, ErrorInvalidInput)
		return
	}

	if strings.TrimSpace(req.Contact) == "" {
		h.writeError(w, http.StatusBadRequest, ErrorInvalidInput)
		return
	}

	if req.PriceType != common.PriceTypeNegotiable && req.Price <= 0 {
		h.writeError(w, http.StatusBadRequest, ErrorInvalidInput)
		return
	}

	property := &common.Property{
		ID:          uuid.New().String(),
		LandlordID:  req.LandlordID,
		Community:   req.Community,
		HouseType:   req.HouseType,
		Area:        req.Area,
		Floor:       req.Floor,
		Orientation: req.Orientation,
		PriceType:   req.PriceType,
		Price:       req.Price,
		Contact:     req.Contact,
	}

	if err := h.store.AddProperty(property); err != nil {
		if err == ErrorDuplicateProperty {
			h.writeError(w, http.StatusConflict, err)
		} else {
			h.writeError(w, http.StatusInternalServerError, err)
		}
		return
	}

	h.writeJSON(w, http.StatusCreated, common.Response{
		Success: true,
		Data:    property,
	})
}

func (h *Handler) UpdateProperty(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		h.writeError(w, http.StatusMethodNotAllowed, ErrorInvalidInput)
		return
	}

	id := strings.TrimPrefix(r.URL.Path, "/properties/")
	if id == "" {
		h.writeError(w, http.StatusBadRequest, ErrorInvalidInput)
		return
	}

	var req struct {
		Community   string           `json:"community"`
		HouseType   string           `json:"house_type"`
		Area        float64          `json:"area"`
		Floor       int              `json:"floor"`
		Orientation string           `json:"orientation"`
		PriceType   common.PriceType `json:"price_type"`
		Price       float64          `json:"price"`
		Contact     string           `json:"contact"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, ErrorInvalidInput)
		return
	}

	err := h.store.UpdateProperty(id, func(p *common.Property) error {
		if strings.TrimSpace(req.Community) != "" {
			p.Community = req.Community
		}
		if req.HouseType != "" {
			if err := common.ValidateHouseType(req.HouseType); err != nil {
				return err
			}
			p.HouseType = req.HouseType
		}
		if req.Area > 0 {
			p.Area = req.Area
		}
		if req.Floor > 0 {
			p.Floor = req.Floor
		}
		if req.Orientation != "" {
			p.Orientation = req.Orientation
		}
		if req.PriceType != "" {
			p.PriceType = req.PriceType
		}
		if req.Price > 0 {
			p.Price = req.Price
		}
		if strings.TrimSpace(req.Contact) != "" {
			p.Contact = req.Contact
		}
		return nil
	})

	if err != nil {
		if err == ErrorPropertyNotFound {
			h.writeError(w, http.StatusNotFound, err)
		} else if err == ErrorPropertySold {
			h.writeError(w, http.StatusForbidden, err)
		} else {
			h.writeError(w, http.StatusInternalServerError, err)
		}
		return
	}

	property, _ := h.store.GetProperty(id)
	h.writeJSON(w, http.StatusOK, common.Response{
		Success: true,
		Data:    property,
	})
}

func (h *Handler) OfflineProperty(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.writeError(w, http.StatusMethodNotAllowed, ErrorInvalidInput)
		return
	}

	id := strings.TrimPrefix(r.URL.Path, "/properties/")
	id = strings.TrimSuffix(id, "/offline")
	if id == "" {
		h.writeError(w, http.StatusBadRequest, ErrorInvalidInput)
		return
	}

	err := h.store.UpdateProperty(id, func(p *common.Property) error {
		p.Status = common.StatusOffline
		return nil
	})

	if err != nil {
		if err == ErrorPropertyNotFound {
			h.writeError(w, http.StatusNotFound, err)
		} else if err == ErrorPropertySold {
			h.writeError(w, http.StatusForbidden, err)
		} else {
			h.writeError(w, http.StatusInternalServerError, err)
		}
		return
	}

	h.writeJSON(w, http.StatusOK, common.Response{
		Success: true,
		Message: "房源已下架",
	})
}

func (h *Handler) SellProperty(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.writeError(w, http.StatusMethodNotAllowed, ErrorInvalidInput)
		return
	}

	id := strings.TrimPrefix(r.URL.Path, "/properties/")
	id = strings.TrimSuffix(id, "/sold")
	if id == "" {
		h.writeError(w, http.StatusBadRequest, ErrorInvalidInput)
		return
	}

	err := h.store.UpdateProperty(id, func(p *common.Property) error {
		p.Status = common.StatusSold
		return nil
	})

	if err != nil {
		if err == ErrorPropertyNotFound {
			h.writeError(w, http.StatusNotFound, err)
		} else {
			h.writeError(w, http.StatusInternalServerError, err)
		}
		return
	}

	h.writeJSON(w, http.StatusOK, common.Response{
		Success: true,
		Message: "房源已标记为成交",
	})
}

func (h *Handler) GetLandlordProperties(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.writeError(w, http.StatusMethodNotAllowed, ErrorInvalidInput)
		return
	}

	landlordID := r.URL.Query().Get("landlord_id")
	if landlordID == "" {
		h.writeError(w, http.StatusBadRequest, ErrorInvalidInput)
		return
	}

	properties := h.store.GetPropertiesByLandlord(landlordID)
	h.writeJSON(w, http.StatusOK, common.Response{
		Success: true,
		Data:    properties,
	})
}

func (h *Handler) FilterProperties(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.writeError(w, http.StatusMethodNotAllowed, ErrorInvalidInput)
		return
	}

	var req common.FilterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, ErrorInvalidInput)
		return
	}

	properties := h.store.FilterProperties(&req)
	h.writeJSON(w, http.StatusOK, common.Response{
		Success: true,
		Data:    properties,
	})
}

func (h *Handler) AddFavorite(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.writeError(w, http.StatusMethodNotAllowed, ErrorInvalidInput)
		return
	}

	var req struct {
		UserID     string `json:"user_id"`
		PropertyID string `json:"property_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, ErrorInvalidInput)
		return
	}

	if req.UserID == "" || req.PropertyID == "" {
		h.writeError(w, http.StatusBadRequest, ErrorInvalidInput)
		return
	}

	favorite := &common.Favorite{
		ID:         uuid.New().String(),
		UserID:     req.UserID,
		PropertyID: req.PropertyID,
	}

	if err := h.store.AddFavorite(favorite); err != nil {
		if err == ErrorAlreadyFavorited {
			h.writeError(w, http.StatusConflict, err)
		} else {
			h.writeError(w, http.StatusInternalServerError, err)
		}
		return
	}

	h.writeJSON(w, http.StatusCreated, common.Response{
		Success: true,
		Data:    favorite,
	})
}

func (h *Handler) RemoveFavorite(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		h.writeError(w, http.StatusMethodNotAllowed, ErrorInvalidInput)
		return
	}

	userID := r.URL.Query().Get("user_id")
	propertyID := r.URL.Query().Get("property_id")

	if userID == "" || propertyID == "" {
		h.writeError(w, http.StatusBadRequest, ErrorInvalidInput)
		return
	}

	if err := h.store.RemoveFavorite(userID, propertyID); err != nil {
		if err == ErrorFavoriteNotFound {
			h.writeError(w, http.StatusNotFound, err)
		} else {
			h.writeError(w, http.StatusInternalServerError, err)
		}
		return
	}

	h.writeJSON(w, http.StatusOK, common.Response{
		Success: true,
		Message: "已取消收藏",
	})
}

func (h *Handler) GetFavorites(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.writeError(w, http.StatusMethodNotAllowed, ErrorInvalidInput)
		return
	}

	userID := r.URL.Query().Get("user_id")
	if userID == "" {
		h.writeError(w, http.StatusBadRequest, ErrorInvalidInput)
		return
	}

	favorites := h.store.GetFavoritesByUser(userID)

	var result []map[string]interface{}
	for _, f := range favorites {
		property, exists := h.store.GetProperty(f.PropertyID)
		item := map[string]interface{}{
			"favorite_id": f.ID,
			"property_id": f.PropertyID,
			"created_at":  f.CreatedAt,
		}

		if !exists {
			item["property_removed"] = true
			item["status"] = "房源已移除"
		} else {
			item["property"] = property
			item["status"] = property.Status
			if property.Status == common.StatusOffline {
				item["status_text"] = "已下架"
			}
		}
		result = append(result, item)
	}

	h.writeJSON(w, http.StatusOK, common.Response{
		Success: true,
		Data:    result,
	})
}

func (h *Handler) CheckFavorite(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.writeError(w, http.StatusMethodNotAllowed, ErrorInvalidInput)
		return
	}

	userID := r.URL.Query().Get("user_id")
	propertyID := r.URL.Query().Get("property_id")

	if userID == "" || propertyID == "" {
		h.writeError(w, http.StatusBadRequest, ErrorInvalidInput)
		return
	}

	isFavorited := h.store.IsFavorited(userID, propertyID)
	h.writeJSON(w, http.StatusOK, common.Response{
		Success: true,
		Data: map[string]bool{
			"is_favorited": isFavorited,
		},
	})
}

func (h *Handler) GetProperty(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.writeError(w, http.StatusMethodNotAllowed, ErrorInvalidInput)
		return
	}

	id := strings.TrimPrefix(r.URL.Path, "/properties/")
	if id == "" {
		h.writeError(w, http.StatusBadRequest, ErrorInvalidInput)
		return
	}

	property, exists := h.store.GetProperty(id)
	if !exists {
		h.writeError(w, http.StatusNotFound, ErrorPropertyNotFound)
		return
	}

	h.writeJSON(w, http.StatusOK, common.Response{
		Success: true,
		Data:    property,
	})
}
