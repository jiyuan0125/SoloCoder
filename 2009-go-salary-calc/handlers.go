package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

func WriteJSONResponse(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(APIResponse{Success: true, Data: data})
}

func WriteError(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(APIResponse{Success: false, Error: message})
}

func GetIDFromPath(path string, prefix string) (int64, error) {
	parts := strings.Split(strings.TrimPrefix(path, prefix), "/")
	if len(parts) == 0 {
		return 0, fmt.Errorf("无效路径")
	}
	return strconv.ParseInt(parts[0], 10, 64)
}

func ParseIntQuery(values url.Values, key string, defaultValue int64) int64 {
	str := values.Get(key)
	if str == "" {
		return defaultValue
	}
	v, err := strconv.ParseInt(str, 10, 64)
	if err != nil {
		return defaultValue
	}
	return v
}

func CreateEmployeeHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	var req Employee
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, http.StatusBadRequest, "无效的请求体: "+err.Error())
		return
	}

	if req.Name == "" {
		WriteError(w, http.StatusBadRequest, "员工姓名不能为空")
		return
	}
	if req.BaseSalary <= 0 {
		WriteError(w, http.StatusBadRequest, "基本工资必须大于0")
		return
	}
	if !ValidatePerformance(req.Performance) {
		WriteError(w, http.StatusBadRequest, "绩效系数必须在 0.5 到 2.0 之间")
		return
	}

	if req.JoinDate.IsZero() {
		req.JoinDate = time.Now().UTC()
	}

	if err := CreateEmployee(db, &req); err != nil {
		WriteError(w, http.StatusInternalServerError, "创建员工失败: "+err.Error())
		return
	}

	WriteJSONResponse(w, http.StatusCreated, req)
}

func GetEmployeeHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	id, err := GetIDFromPath(r.URL.Path, "/employees/")
	if err != nil {
		WriteError(w, http.StatusBadRequest, "无效的员工ID")
		return
	}

	emp, err := GetEmployee(db, id)
	if err != nil {
		if err.Error() == "员工不存在" {
			WriteError(w, http.StatusNotFound, err.Error())
			return
		}
		WriteError(w, http.StatusInternalServerError, "获取员工失败: "+err.Error())
		return
	}

	WriteJSONResponse(w, http.StatusOK, emp)
}

func ListEmployeesHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	employees, err := ListEmployees(db)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "获取员工列表失败: "+err.Error())
		return
	}

	WriteJSONResponse(w, http.StatusOK, employees)
}

func UpdateEmployeeHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	id, err := GetIDFromPath(r.URL.Path, "/employees/")
	if err != nil {
		WriteError(w, http.StatusBadRequest, "无效的员工ID")
		return
	}

	var req Employee
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, http.StatusBadRequest, "无效的请求体: "+err.Error())
		return
	}

	req.ID = id

	if req.BaseSalary <= 0 {
		WriteError(w, http.StatusBadRequest, "基本工资必须大于0")
		return
	}
	if !ValidatePerformance(req.Performance) {
		WriteError(w, http.StatusBadRequest, "绩效系数必须在 0.5 到 2.0 之间")
		return
	}

	if err := UpdateEmployee(db, &req); err != nil {
		if err.Error() == "员工不存在" {
			WriteError(w, http.StatusNotFound, err.Error())
			return
		}
		WriteError(w, http.StatusInternalServerError, "更新员工失败: "+err.Error())
		return
	}

	emp, err := GetEmployee(db, id)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "获取更新后员工失败: "+err.Error())
		return
	}

	WriteJSONResponse(w, http.StatusOK, emp)
}

func DeleteEmployeeHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	id, err := GetIDFromPath(r.URL.Path, "/employees/")
	if err != nil {
		WriteError(w, http.StatusBadRequest, "无效的员工ID")
		return
	}

	if err := DeleteEmployee(db, id); err != nil {
		if err.Error() == "员工不存在" {
			WriteError(w, http.StatusNotFound, err.Error())
			return
		}
		WriteError(w, http.StatusInternalServerError, "删除员工失败: "+err.Error())
		return
	}

	WriteJSONResponse(w, http.StatusOK, map[string]string{"message": "员工已删除"})
}

func CreateResourceHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	var req Resource
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, http.StatusBadRequest, "无效的请求体: "+err.Error())
		return
	}

	if req.Name == "" {
		WriteError(w, http.StatusBadRequest, "资源名称不能为空")
		return
	}
	if req.Type == "" {
		WriteError(w, http.StatusBadRequest, "资源类型不能为空")
		return
	}

	if err := CreateResource(db, &req); err != nil {
		WriteError(w, http.StatusInternalServerError, "创建资源失败: "+err.Error())
		return
	}

	WriteJSONResponse(w, http.StatusCreated, req)
}

func GetResourceHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	id, err := GetIDFromPath(r.URL.Path, "/resources/")
	if err != nil {
		WriteError(w, http.StatusBadRequest, "无效的资源ID")
		return
	}

	res, err := GetResource(db, id)
	if err != nil {
		if err.Error() == "资源不存在" {
			WriteError(w, http.StatusNotFound, err.Error())
			return
		}
		WriteError(w, http.StatusInternalServerError, "获取资源失败: "+err.Error())
		return
	}

	WriteJSONResponse(w, http.StatusOK, res)
}

func ListResourcesHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	resourceType := r.URL.Query().Get("type")
	resources, err := ListResources(db, resourceType)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "获取资源列表失败: "+err.Error())
		return
	}

	WriteJSONResponse(w, http.StatusOK, resources)
}

func UpdateResourceHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	id, err := GetIDFromPath(r.URL.Path, "/resources/")
	if err != nil {
		WriteError(w, http.StatusBadRequest, "无效的资源ID")
		return
	}

	var req Resource
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, http.StatusBadRequest, "无效的请求体: "+err.Error())
		return
	}

	req.ID = id

	if req.Name == "" {
		WriteError(w, http.StatusBadRequest, "资源名称不能为空")
		return
	}
	if req.Type == "" {
		WriteError(w, http.StatusBadRequest, "资源类型不能为空")
		return
	}

	if err := UpdateResource(db, &req); err != nil {
		if err.Error() == "资源不存在" {
			WriteError(w, http.StatusNotFound, err.Error())
			return
		}
		WriteError(w, http.StatusInternalServerError, "更新资源失败: "+err.Error())
		return
	}

	res, err := GetResource(db, id)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "获取更新后资源失败: "+err.Error())
		return
	}

	WriteJSONResponse(w, http.StatusOK, res)
}

func DeleteResourceHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	id, err := GetIDFromPath(r.URL.Path, "/resources/")
	if err != nil {
		WriteError(w, http.StatusBadRequest, "无效的资源ID")
		return
	}

	if err := DeleteResource(db, id); err != nil {
		if err.Error() == "资源不存在" {
			WriteError(w, http.StatusNotFound, err.Error())
			return
		}
		WriteError(w, http.StatusInternalServerError, "删除资源失败: "+err.Error())
		return
	}

	WriteJSONResponse(w, http.StatusOK, map[string]string{"message": "资源已删除"})
}

func CreateResourceRelationHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	var req ResourceRelation
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, http.StatusBadRequest, "无效的请求体: "+err.Error())
		return
	}

	if req.SourceID <= 0 || req.SourceType == "" {
		WriteError(w, http.StatusBadRequest, "源资源信息不完整")
		return
	}
	if req.TargetID <= 0 || req.TargetType == "" {
		WriteError(w, http.StatusBadRequest, "目标资源信息不完整")
		return
	}
	if req.Relation == "" {
		WriteError(w, http.StatusBadRequest, "关联关系不能为空")
		return
	}

	if err := CreateResourceRelation(db, &req); err != nil {
		WriteError(w, http.StatusInternalServerError, "创建资源关联失败: "+err.Error())
		return
	}

	WriteJSONResponse(w, http.StatusCreated, req)
}

func ListResourceRelationsHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	resourceID := ParseIntQuery(r.URL.Query(), "resource_id", 0)
	resourceType := r.URL.Query().Get("resource_type")

	if resourceID <= 0 || resourceType == "" {
		WriteError(w, http.StatusBadRequest, "必须提供 resource_id 和 resource_type")
		return
	}

	relations, err := ListResourceRelations(db, resourceID, resourceType)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "获取资源关联列表失败: "+err.Error())
		return
	}

	WriteJSONResponse(w, http.StatusOK, relations)
}

func DeleteResourceRelationHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	id, err := GetIDFromPath(r.URL.Path, "/resource-relations/")
	if err != nil {
		WriteError(w, http.StatusBadRequest, "无效的关联ID")
		return
	}

	if err := DeleteResourceRelation(db, id); err != nil {
		if err.Error() == "资源关联不存在" {
			WriteError(w, http.StatusNotFound, err.Error())
			return
		}
		WriteError(w, http.StatusInternalServerError, "删除资源关联失败: "+err.Error())
		return
	}

	WriteJSONResponse(w, http.StatusOK, map[string]string{"message": "资源关联已删除"})
}

func CalculateSalaryHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	var req SalaryCalculationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, http.StatusBadRequest, "无效的请求体: "+err.Error())
		return
	}

	if req.EmployeeID <= 0 {
		WriteError(w, http.StatusBadRequest, "员工ID不能为空")
		return
	}
	if req.Year <= 0 || req.Month <= 0 || req.Month > 12 {
		WriteError(w, http.StatusBadRequest, "无效的年月")
		return
	}

	emp, err := GetEmployee(db, req.EmployeeID)
	if err != nil {
		if err.Error() == "员工不存在" {
			WriteError(w, http.StatusNotFound, err.Error())
			return
		}
		WriteError(w, http.StatusInternalServerError, "获取员工失败: "+err.Error())
		return
	}

	if !ValidatePerformance(emp.Performance) {
		WriteError(w, http.StatusBadRequest, "绩效系数必须在 0.5 到 2.0 之间")
		return
	}

	record, err := CalculateFullSalary(emp, req.Year, req.Month, req.Overtime)
	if err != nil {
		if ve, ok := err.(*ValidationError); ok {
			WriteError(w, http.StatusBadRequest, ve.Message)
			return
		}
		WriteError(w, http.StatusInternalServerError, "薪资计算失败: "+err.Error())
		return
	}

	if req.ResourceID > 0 && req.ResourceType != "" {
		record.ResourceID = &req.ResourceID
		record.ResourceType = &req.ResourceType
	}

	if err := CreateSalaryRecord(db, record); err != nil {
		if strings.Contains(err.Error(), "已存在") {
			WriteError(w, http.StatusConflict, err.Error())
			return
		}
		WriteError(w, http.StatusInternalServerError, "保存薪资记录失败: "+err.Error())
		return
	}

	WriteJSONResponse(w, http.StatusCreated, record)
}

func GetSalaryHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	id, err := GetIDFromPath(r.URL.Path, "/salaries/")
	if err != nil {
		WriteError(w, http.StatusBadRequest, "无效的薪资记录ID")
		return
	}

	record, err := GetSalaryRecord(db, id)
	if err != nil {
		if err.Error() == "薪资记录不存在" {
			WriteError(w, http.StatusNotFound, err.Error())
			return
		}
		WriteError(w, http.StatusInternalServerError, "获取薪资记录失败: "+err.Error())
		return
	}

	WriteJSONResponse(w, http.StatusOK, record)
}

func ListSalariesHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	employeeID := ParseIntQuery(r.URL.Query(), "employee_id", 0)
	year := ParseIntQuery(r.URL.Query(), "year", 0)
	month := ParseIntQuery(r.URL.Query(), "month", 0)

	records, err := ListSalaryRecords(db, employeeID, int(year), int(month))
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "获取薪资记录列表失败: "+err.Error())
		return
	}

	WriteJSONResponse(w, http.StatusOK, records)
}

func UpdateSalaryStatusHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	path := r.URL.Path
	prefix := "/salaries/"
	suffix := "/next-status"

	if !strings.HasPrefix(path, prefix) || !strings.HasSuffix(path, suffix) {
		WriteError(w, http.StatusBadRequest, "无效的请求路径")
		return
	}

	idStr := strings.TrimPrefix(path, prefix)
	idStr = strings.TrimSuffix(idStr, suffix)

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		WriteError(w, http.StatusBadRequest, "无效的薪资记录ID")
		return
	}

	record, err := UpdateSalaryStatus(db, id)
	if err != nil {
		if err.Error() == "薪资记录不存在" {
			WriteError(w, http.StatusNotFound, err.Error())
			return
		}
		if strings.Contains(err.Error(), "无法继续推进") {
			WriteError(w, http.StatusBadRequest, err.Error())
			return
		}
		WriteError(w, http.StatusInternalServerError, "更新状态失败: "+err.Error())
		return
	}

	WriteJSONResponse(w, http.StatusOK, record)
}

func ResourceSummaryHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	resourceID := ParseIntQuery(r.URL.Query(), "resource_id", 0)
	resourceType := r.URL.Query().Get("resource_type")

	if resourceID > 0 && resourceType != "" {
		summary, err := GetResourceSummary(db, resourceID, resourceType)
		if err != nil {
			if err.Error() == "资源不存在" {
				WriteError(w, http.StatusNotFound, err.Error())
				return
			}
			WriteError(w, http.StatusInternalServerError, "获取资源汇总失败: "+err.Error())
			return
		}
		WriteJSONResponse(w, http.StatusOK, summary)
		return
	}

	summaries, err := ListAllResourceSummaries(db, resourceType)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "获取资源汇总列表失败: "+err.Error())
		return
	}

	WriteJSONResponse(w, http.StatusOK, summaries)
}
