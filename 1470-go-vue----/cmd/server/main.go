package main

import (
	"encoding/json"
	"flag"
	"io"
	"net/http"
	"os"
	"strings"

	"settlement/pkg/common"
	"settlement/pkg/core"
)

var store = core.NewStore()

func jsonResponse(w http.ResponseWriter, success bool, message string, data interface{}, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	resp := common.APIResponse{
		Success: success,
		Message: message,
		Data:    data,
	}
	json.NewEncoder(w).Encode(resp)
}

func readBody(r *http.Request, v interface{}) error {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return err
	}
	defer r.Body.Close()
	return json.Unmarshal(body, v)
}

func getOperator(r *http.Request) string {
	op := r.Header.Get("X-Operator")
	if op == "" {
		return "system"
	}
	return op
}

func handleCreateProject(w http.ResponseWriter, r *http.Request) {
	var req common.CreateProjectRequest
	if err := readBody(r, &req); err != nil {
		jsonResponse(w, false, "无效请求", nil, http.StatusBadRequest)
		return
	}
	if req.Name == "" {
		jsonResponse(w, false, "项目名称不能为空", nil, http.StatusBadRequest)
		return
	}

	operator := getOperator(r)
	p, err := store.CreateProject(req.Name, req.ContractAmount, operator)
	if err != nil {
		jsonResponse(w, false, err.Error(), nil, http.StatusInternalServerError)
		return
	}
	jsonResponse(w, true, "创建成功", p, http.StatusOK)
}

func handleListProjects(w http.ResponseWriter, r *http.Request) {
	projects := store.ListProjects()
	jsonResponse(w, true, "", projects, http.StatusOK)
}

func handleGetProject(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/api/projects/")
	p, err := store.GetProject(id)
	if err != nil {
		jsonResponse(w, false, err.Error(), nil, http.StatusNotFound)
		return
	}
	jsonResponse(w, true, "", p, http.StatusOK)
}

func handleAddBOQ(w http.ResponseWriter, r *http.Request) {
	var req common.CreateBOQRequest
	if err := readBody(r, &req); err != nil {
		jsonResponse(w, false, "无效请求", nil, http.StatusBadRequest)
		return
	}
	operator := getOperator(r)
	boq, err := store.AddBOQ(req.ProjectID, req.ItemCode, req.ItemName, req.Unit,
		req.ContractQty, req.ActualQty, req.UnitPrice, operator)
	if err != nil {
		jsonResponse(w, false, err.Error(), nil, http.StatusBadRequest)
		return
	}
	jsonResponse(w, true, "添加成功", boq, http.StatusOK)
}

func handleListBOQ(w http.ResponseWriter, r *http.Request) {
	projectID := r.URL.Query().Get("project_id")
	if projectID == "" {
		jsonResponse(w, false, "缺少 project_id 参数", nil, http.StatusBadRequest)
		return
	}
	boqs, err := store.ListBOQ(projectID)
	if err != nil {
		jsonResponse(w, false, err.Error(), nil, http.StatusNotFound)
		return
	}
	jsonResponse(w, true, "", boqs, http.StatusOK)
}

func handleUpdateBOQ(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/boqs/"), "/")
	if len(parts) < 2 {
		jsonResponse(w, false, "无效路径", nil, http.StatusBadRequest)
		return
	}
	projectID := parts[0]
	boqID := parts[1]

	var req common.UpdateBOQRequest
	if err := readBody(r, &req); err != nil {
		jsonResponse(w, false, "无效请求", nil, http.StatusBadRequest)
		return
	}

	operator := getOperator(r)
	boq, err := store.UpdateBOQ(projectID, boqID, &req, operator)
	if err != nil {
		jsonResponse(w, false, err.Error(), nil, http.StatusBadRequest)
		return
	}
	jsonResponse(w, true, "更新成功", boq, http.StatusOK)
}

func handleDeleteBOQ(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/boqs/"), "/")
	if len(parts) < 2 {
		jsonResponse(w, false, "无效路径", nil, http.StatusBadRequest)
		return
	}
	projectID := parts[0]
	boqID := parts[1]

	operator := getOperator(r)
	err := store.DeleteBOQ(projectID, boqID, operator)
	if err != nil {
		jsonResponse(w, false, err.Error(), nil, http.StatusBadRequest)
		return
	}
	jsonResponse(w, true, "删除成功", nil, http.StatusOK)
}

func handleCreateVisa(w http.ResponseWriter, r *http.Request) {
	var req common.CreateVisaRequest
	if err := readBody(r, &req); err != nil {
		jsonResponse(w, false, "无效请求", nil, http.StatusBadRequest)
		return
	}
	operator := getOperator(r)
	if req.Creator == "" {
		req.Creator = operator
	}
	visa, err := store.CreateVisa(req.ProjectID, req.Reason, req.ChangeContent,
		req.IncreaseQty, req.DecreaseQty, req.UnitPrice, req.Creator)
	if err != nil {
		jsonResponse(w, false, err.Error(), nil, http.StatusBadRequest)
		return
	}
	jsonResponse(w, true, "创建成功", visa, http.StatusOK)
}

func handleListVisa(w http.ResponseWriter, r *http.Request) {
	projectID := r.URL.Query().Get("project_id")
	if projectID == "" {
		jsonResponse(w, false, "缺少 project_id 参数", nil, http.StatusBadRequest)
		return
	}
	visas, err := store.ListVisa(projectID)
	if err != nil {
		jsonResponse(w, false, err.Error(), nil, http.StatusNotFound)
		return
	}
	jsonResponse(w, true, "", visas, http.StatusOK)
}

func handleUpdateVisa(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/visas/"), "/")
	if len(parts) < 2 {
		jsonResponse(w, false, "无效路径", nil, http.StatusBadRequest)
		return
	}
	projectID := parts[0]
	visaID := parts[1]

	var req common.UpdateVisaRequest
	if err := readBody(r, &req); err != nil {
		jsonResponse(w, false, "无效请求", nil, http.StatusBadRequest)
		return
	}

	operator := getOperator(r)
	visa, err := store.UpdateVisa(projectID, visaID, &req, operator)
	if err != nil {
		jsonResponse(w, false, err.Error(), nil, http.StatusBadRequest)
		return
	}
	jsonResponse(w, true, "更新成功", visa, http.StatusOK)
}

func handleConfirmVisa(w http.ResponseWriter, r *http.Request) {
	var req common.ConfirmVisaRequest
	if err := readBody(r, &req); err != nil {
		jsonResponse(w, false, "无效请求", nil, http.StatusBadRequest)
		return
	}
	operator := getOperator(r)
	if req.Operator != "" {
		operator = req.Operator
	}

	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/visas/"), "/")
	projectID := parts[0]

	visa, err := store.ConfirmVisa(projectID, req.VisaID, operator, req.Role)
	if err != nil {
		jsonResponse(w, false, err.Error(), nil, http.StatusBadRequest)
		return
	}
	jsonResponse(w, true, "确认成功", visa, http.StatusOK)
}

func handleRejectVisa(w http.ResponseWriter, r *http.Request) {
	var req common.RejectVisaRequest
	if err := readBody(r, &req); err != nil {
		jsonResponse(w, false, "无效请求", nil, http.StatusBadRequest)
		return
	}
	operator := getOperator(r)
	if req.Operator != "" {
		operator = req.Operator
	}

	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/visas/"), "/")
	projectID := parts[0]

	visa, err := store.RejectVisa(projectID, req.VisaID, operator, req.Role, req.Reason)
	if err != nil {
		jsonResponse(w, false, err.Error(), nil, http.StatusBadRequest)
		return
	}
	jsonResponse(w, true, "驳回成功", visa, http.StatusOK)
}

func handleGetSettlement(w http.ResponseWriter, r *http.Request) {
	projectID := r.URL.Query().Get("project_id")
	if projectID == "" {
		jsonResponse(w, false, "缺少 project_id 参数", nil, http.StatusBadRequest)
		return
	}
	settlement, err := store.GetSettlement(projectID)
	if err != nil {
		jsonResponse(w, false, err.Error(), nil, http.StatusNotFound)
		return
	}
	jsonResponse(w, true, "", settlement, http.StatusOK)
}

func handleAuditPass(w http.ResponseWriter, r *http.Request) {
	var req common.AuditRequest
	if err := readBody(r, &req); err != nil {
		jsonResponse(w, false, "无效请求", nil, http.StatusBadRequest)
		return
	}
	operator := getOperator(r)
	if req.Auditor == "" {
		req.Auditor = operator
	}
	settlement, err := store.AuditPass(req.ProjectID, req.Auditor, req.Opinion)
	if err != nil {
		jsonResponse(w, false, err.Error(), nil, http.StatusBadRequest)
		return
	}
	jsonResponse(w, true, "审计通过", settlement, http.StatusOK)
}

func handleAuditLogs(w http.ResponseWriter, r *http.Request) {
	projectID := r.URL.Query().Get("project_id")
	if projectID == "" {
		jsonResponse(w, false, "缺少 project_id 参数", nil, http.StatusBadRequest)
		return
	}
	logs, err := store.GetAuditLogs(projectID)
	if err != nil {
		jsonResponse(w, false, err.Error(), nil, http.StatusNotFound)
		return
	}
	jsonResponse(w, true, "", logs, http.StatusOK)
}

func main() {
	var port string
	flag.StringVar(&port, "port", "", "服务监听端口")
	flag.Parse()

	if port == "" {
		port = os.Getenv("SETTLEMENT_PORT")
	}
	if port == "" {
		port = "8080"
	}

	mux := http.NewServeMux()

	mux.HandleFunc("/api/projects", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			handleCreateProject(w, r)
		} else if r.Method == http.MethodGet {
			handleListProjects(w, r)
		} else {
			jsonResponse(w, false, "不支持的方法", nil, http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/api/projects/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			handleGetProject(w, r)
		} else {
			jsonResponse(w, false, "不支持的方法", nil, http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/api/boqs", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			handleAddBOQ(w, r)
		} else if r.Method == http.MethodGet {
			handleListBOQ(w, r)
		} else {
			jsonResponse(w, false, "不支持的方法", nil, http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/api/boqs/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPut {
			handleUpdateBOQ(w, r)
		} else if r.Method == http.MethodDelete {
			handleDeleteBOQ(w, r)
		} else {
			jsonResponse(w, false, "不支持的方法", nil, http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/api/visas", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			handleCreateVisa(w, r)
		} else if r.Method == http.MethodGet {
			handleListVisa(w, r)
		} else {
			jsonResponse(w, false, "不支持的方法", nil, http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/api/visas/", func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/confirm") && r.Method == http.MethodPost {
			handleConfirmVisa(w, r)
		} else if strings.HasSuffix(r.URL.Path, "/reject") && r.Method == http.MethodPost {
			handleRejectVisa(w, r)
		} else if r.Method == http.MethodPut {
			handleUpdateVisa(w, r)
		} else {
			jsonResponse(w, false, "不支持的方法", nil, http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/api/settlement", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			handleGetSettlement(w, r)
		} else {
			jsonResponse(w, false, "不支持的方法", nil, http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/api/audit/pass", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			handleAuditPass(w, r)
		} else {
			jsonResponse(w, false, "不支持的方法", nil, http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/api/audit/logs", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			handleAuditLogs(w, r)
		} else {
			jsonResponse(w, false, "不支持的方法", nil, http.StatusMethodNotAllowed)
		}
	})

	addr := ":" + port
	println("竣工结算服务已启动，监听端口:", port)
	http.ListenAndServe(addr, mux)
}
