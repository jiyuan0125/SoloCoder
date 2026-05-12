package handlers

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"health-archive/internal/models"
	"health-archive/internal/store"
	"health-archive/internal/utils"
)

type Handler struct {
	store *store.Store
}

func New(s *store.Store) *Handler {
	return &Handler{store: s}
}

func jsonResponse(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if data != nil {
		json.NewEncoder(w).Encode(data)
	}
}

func errorResponse(w http.ResponseWriter, status int, message string) {
	jsonResponse(w, status, map[string]string{"error": message})
}

func (h *Handler) EnableCORS(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		next(w, r)
	}
}

func (h *Handler) CreateResident(w http.ResponseWriter, r *http.Request) {
	var resident models.Resident
	if err := json.NewDecoder(r.Body).Decode(&resident); err != nil {
		errorResponse(w, http.StatusBadRequest, "invalid request body")
		return
	}

	idCard := utils.NormalizeIDCard(resident.IDCard)
	if !utils.ValidateIDCard(idCard) {
		errorResponse(w, http.StatusBadRequest, "invalid ID card format")
		return
	}

	created, err := h.store.AddResident(&resident)
	if err != nil {
		if strings.Contains(err.Error(), "already exists") {
			errorResponse(w, http.StatusConflict, err.Error())
			return
		}
		if strings.Contains(err.Error(), "guardian") {
			errorResponse(w, http.StatusBadRequest, err.Error())
			return
		}
		errorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	jsonResponse(w, http.StatusCreated, created)
}

func (h *Handler) GetResident(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Path[len("/api/residents/"):]
	resident, err := h.store.GetResident(id)
	if err != nil {
		errorResponse(w, http.StatusNotFound, "resident not found")
		return
	}
	jsonResponse(w, http.StatusOK, resident)
}

func (h *Handler) SearchResidents(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("name")
	idCard := r.URL.Query().Get("idCard")
	community := r.URL.Query().Get("community")

	residents := h.store.SearchResidents(name, idCard, community)
	if residents == nil {
		residents = []*models.Resident{}
	}
	jsonResponse(w, http.StatusOK, residents)
}

func (h *Handler) CreateCheckup(w http.ResponseWriter, r *http.Request) {
	var checkup models.Checkup
	if err := json.NewDecoder(r.Body).Decode(&checkup); err != nil {
		errorResponse(w, http.StatusBadRequest, "invalid request body")
		return
	}

	created, err := h.store.AddCheckup(&checkup)
	if err != nil {
		if strings.Contains(err.Error(), "resident not found") {
			errorResponse(w, http.StatusNotFound, err.Error())
			return
		}
		errorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	jsonResponse(w, http.StatusCreated, created)
}

func (h *Handler) GetCheckups(w http.ResponseWriter, r *http.Request) {
	residentID := r.URL.Query().Get("residentID")
	checkups := h.store.GetCheckups(residentID)
	if checkups == nil {
		checkups = []*models.Checkup{}
	}
	jsonResponse(w, http.StatusOK, checkups)
}

func (h *Handler) CreateFollowup(w http.ResponseWriter, r *http.Request) {
	var followup models.FollowupRecord
	if err := json.NewDecoder(r.Body).Decode(&followup); err != nil {
		errorResponse(w, http.StatusBadRequest, "invalid request body")
		return
	}

	created, err := h.store.AddFollowup(&followup)
	if err != nil {
		if strings.Contains(err.Error(), "resident not found") {
			errorResponse(w, http.StatusNotFound, err.Error())
			return
		}
		errorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	jsonResponse(w, http.StatusCreated, created)
}

func (h *Handler) GetFollowups(w http.ResponseWriter, r *http.Request) {
	residentID := r.URL.Query().Get("residentID")
	disease := models.ChronicDisease(r.URL.Query().Get("disease"))
	followups := h.store.GetFollowups(residentID, disease)
	if followups == nil {
		followups = []*models.FollowupRecord{}
	}
	jsonResponse(w, http.StatusOK, followups)
}

func (h *Handler) GetChronicPatients(w http.ResponseWriter, r *http.Request) {
	patients := h.store.GetAllChronicPatients()
	if patients == nil {
		patients = []*models.FollowupRecord{}
	}
	jsonResponse(w, http.StatusOK, patients)
}

func (h *Handler) CreateFamily(w http.ResponseWriter, r *http.Request) {
	var family models.Family
	if err := json.NewDecoder(r.Body).Decode(&family); err != nil {
		errorResponse(w, http.StatusBadRequest, "invalid request body")
		return
	}

	created, err := h.store.AddFamily(&family)
	if err != nil {
		if strings.Contains(err.Error(), "resident not found") {
			errorResponse(w, http.StatusNotFound, err.Error())
			return
		}
		errorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	jsonResponse(w, http.StatusCreated, created)
}

func (h *Handler) GetFamilies(w http.ResponseWriter, r *http.Request) {
	families := h.store.GetAllFamilies()
	if families == nil {
		families = []*models.Family{}
	}
	jsonResponse(w, http.StatusOK, families)
}

func (h *Handler) GetFamily(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Path[len("/api/families/"):]
	family, err := h.store.GetFamily(id)
	if err != nil {
		errorResponse(w, http.StatusNotFound, "family not found")
		return
	}
	jsonResponse(w, http.StatusOK, family)
}

func (h *Handler) ExportCommunityStats(w http.ResponseWriter, r *http.Request) {
	community := r.URL.Query().Get("community")
	stats := h.store.GetCommunityStats(community)

	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", "attachment; filename=community_stats.csv")
	writer := csv.NewWriter(w)

	writer.Write([]string{"指标", "数值"})
	writer.Write([]string{"总人数", strconv.Itoa(stats.TotalCount)})
	writer.Write([]string{"男性人数", strconv.Itoa(stats.MaleCount)})
	writer.Write([]string{"女性人数", strconv.Itoa(stats.FemaleCount)})
	writer.Write([]string{"年龄分布", ""})
	for group, count := range stats.AgeGroups {
		writer.Write([]string{group, strconv.Itoa(count)})
	}
	writer.Flush()
}

func (h *Handler) ExportChronicPatients(w http.ResponseWriter, r *http.Request) {
	patients := h.store.GetAllChronicPatients()

	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", "attachment; filename=chronic_patients.csv")
	writer := csv.NewWriter(w)

	writer.Write([]string{"居民ID", "姓名", "慢性病", "最近随访日期", "最近随访数据"})

	for _, p := range patients {
		resident, err := h.store.GetResident(p.ResidentID)
		if err != nil {
			continue
		}

		data := ""
		switch p.Disease {
		case models.ChronicHypertension:
			if p.Hypertension != nil {
				data = fmt.Sprintf("收缩压:%d, 舒张压:%d, 用药:%s",
					p.Hypertension.SystolicBP,
					p.Hypertension.DiastolicBP,
					p.Hypertension.Medication)
			}
		case models.ChronicDiabetes:
			if p.Diabetes != nil {
				data = fmt.Sprintf("空腹血糖:%.1f, 餐后血糖:%.1f, 糖化血红蛋白:%.1f, 用药:%s",
					p.Diabetes.FastingBloodSugar,
					p.Diabetes.PostprandialBloodSugar,
					p.Diabetes.HbA1c,
					p.Diabetes.Medication)
			}
		case models.ChronicCoronary:
			if p.Coronary != nil {
				data = fmt.Sprintf("心功能评估:%s, 用药:%s",
					p.Coronary.CardiacFunction,
					p.Coronary.Medication)
			}
		}

		writer.Write([]string{
			resident.ID,
			resident.Name,
			string(p.Disease),
			p.Date.Format("2006-01-02"),
			data,
		})
	}
	writer.Flush()
}

func (h *Handler) ExportCompleteArchive(w http.ResponseWriter, r *http.Request) {
	residentID := r.URL.Query().Get("residentID")

	resident, err := h.store.GetResident(residentID)
	if err != nil {
		errorResponse(w, http.StatusNotFound, "resident not found")
		return
	}

	checkups := h.store.GetCheckups(residentID)
	followups := h.store.GetFollowups(residentID, "")

	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=archive_%s.csv", resident.Name))
	writer := csv.NewWriter(w)

	writer.Write([]string{"=== 基本信息"})
	writer.Write([]string{"姓名", resident.Name})
	writer.Write([]string{"性别", string(resident.Gender)})
	writer.Write([]string{"身份证号", resident.IDCard})
	writer.Write([]string{"出生日期", resident.BirthDate.Format("2006-01-02")})
	writer.Write([]string{"联系电话", resident.Phone})
	writer.Write([]string{"家庭住址", resident.Address})
	writer.Write([]string{"紧急联系人", resident.EmergencyContact})
	writer.Write([]string{"建档日期", resident.CreateDate.Format("2006-01-02")})
	writer.Write([]string{"责任医生", resident.DoctorInCharge})
	writer.Write([]string{"血型", string(resident.BloodType)})
	if len(resident.Allergies) > 0 {
		for _, a := range resident.Allergies {
			writer.Write([]string{"过敏史", fmt.Sprintf("过敏源:%s, 反应:%s", a.Allergen, a.Reaction)})
		}
	}

	writer.Write([]string{""})
	writer.Write([]string{"=== 体检记录"})
	writer.Write([]string{"日期", "身高", "体重", "BMI", "收缩压", "舒张压", "心率", "左眼视力", "右眼视力", "是否初始体检"})
	for _, c := range checkups {
		writer.Write([]string{
			c.Date.Format("2006-01-02"),
			fmt.Sprintf("%.1f", c.Height),
			fmt.Sprintf("%.1f", c.Weight),
			fmt.Sprintf("%.1f", c.BMI),
			strconv.Itoa(c.SystolicBP),
			strconv.Itoa(c.DiastolicBP),
			strconv.Itoa(c.HeartRate),
			fmt.Sprintf("%.1f", c.VisionLeft),
			fmt.Sprintf("%.1f", c.VisionRight),
			strconv.FormatBool(c.IsInitial),
		})
	}

	writer.Write([]string{""})
	writer.Write([]string{"=== 慢性病随访记录"})
	writer.Write([]string{"日期", "疾病类型", "随访数据", "下次随访日期", "是否超期"})
	for _, f := range followups {
		data := ""
		switch f.Disease {
		case models.ChronicHypertension:
			if f.Hypertension != nil {
				data = fmt.Sprintf("收缩压:%d, 舒张压:%d, 用药:%s",
					f.Hypertension.SystolicBP,
					f.Hypertension.DiastolicBP,
					f.Hypertension.Medication)
			}
		case models.ChronicDiabetes:
			if f.Diabetes != nil {
				data = fmt.Sprintf("空腹血糖:%.1f, 餐后血糖:%.1f, 糖化血红蛋白:%.1f, 用药:%s",
					f.Diabetes.FastingBloodSugar,
					f.Diabetes.PostprandialBloodSugar,
					f.Diabetes.HbA1c,
					f.Diabetes.Medication)
			}
		case models.ChronicCoronary:
			if f.Coronary != nil {
				data = fmt.Sprintf("心功能评估:%s, 用药:%s",
					f.Coronary.CardiacFunction,
					f.Coronary.Medication)
			}
		}
		writer.Write([]string{
			f.Date.Format("2006-01-02"),
			string(f.Disease),
			data,
			f.NextDueDate.Format("2006-01-02"),
			strconv.FormatBool(f.IsOverdue),
		})
	}

	writer.Flush()
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	mux := http.NewServeMux()

	mux.HandleFunc("/api/residents", h.EnableCORS(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			h.SearchResidents(w, r)
		case http.MethodPost:
			h.CreateResident(w, r)
		}
	}))

	mux.HandleFunc("/api/residents/", h.EnableCORS(h.GetResident))

	mux.HandleFunc("/api/checkups", h.EnableCORS(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			h.GetCheckups(w, r)
		case http.MethodPost:
			h.CreateCheckup(w, r)
		}
	}))

	mux.HandleFunc("/api/followups", h.EnableCORS(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			h.GetFollowups(w, r)
		case http.MethodPost:
			h.CreateFollowup(w, r)
		}
	}))

	mux.HandleFunc("/api/chronic-patients", h.EnableCORS(h.GetChronicPatients))

	mux.HandleFunc("/api/families", h.EnableCORS(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			h.GetFamilies(w, r)
		case http.MethodPost:
			h.CreateFamily(w, r)
		}
	}))

	mux.HandleFunc("/api/families/", h.EnableCORS(h.GetFamily))

	mux.HandleFunc("/api/export/community-stats", h.EnableCORS(h.ExportCommunityStats))
	mux.HandleFunc("/api/export/chronic-patients", h.EnableCORS(h.ExportChronicPatients))
	mux.HandleFunc("/api/export/complete-archive", h.EnableCORS(h.ExportCompleteArchive))

	mux.ServeHTTP(w, r)
}
