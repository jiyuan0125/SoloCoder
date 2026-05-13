package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"approval-flow/pkg/model"
	"approval-flow/pkg/service"
	"approval-flow/pkg/store"
)

var approvalService = service.NewApprovalService()

func InitTestData() error {
	users := []*model.User{
		{ID: "user1", Name: "申请人A", Token: "token1", ManagerID: "user2"},
		{ID: "user2", Name: "审批人B", Token: "token2", ManagerID: "user3"},
		{ID: "user3", Name: "审批人C", Token: "token3", ManagerID: "user4"},
		{ID: "user4", Name: "审批人D", Token: "token4"},
	}

	for _, u := range users {
		existing, err := store.GetUserByID(u.ID)
		if err != nil {
			if err := store.CreateUser(u); err != nil {
				return err
			}
		} else {
			_ = existing
		}
	}

	return nil
}

func HandleCreateUser(w http.ResponseWriter, r *http.Request) {
	var user model.User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		WriteError(w, errors.New("invalid request body"))
		return
	}

	if err := store.CreateUser(&user); err != nil {
		WriteError(w, err)
		return
	}

	WriteJSON(w, http.StatusCreated, user)
}

func HandleListUsers(w http.ResponseWriter, r *http.Request) {
	users, err := store.ListUsers()
	if err != nil {
		WriteError(w, err)
		return
	}

	WriteJSON(w, http.StatusOK, users)
}

func HandleGetUser(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/api/users/")
	user, err := store.GetUserByID(id)
	if err != nil {
		WriteError(w, err)
		return
	}

	WriteJSON(w, http.StatusOK, user)
}

func HandleCreateChain(w http.ResponseWriter, r *http.Request) {
	var chain model.ApprovalChain
	if err := json.NewDecoder(r.Body).Decode(&chain); err != nil {
		WriteError(w, errors.New("invalid request body"))
		return
	}

	if len(chain.Nodes) == 0 {
		WriteError(w, service.ErrChainEmpty)
		return
	}

	for _, node := range chain.Nodes {
		if node.Condition != "" {
			if err := service.ValidateExpression(node.Condition); err != nil {
				WriteError(w, err)
				return
			}
		}
	}

	if err := store.CreateChain(&chain); err != nil {
		WriteError(w, err)
		return
	}

	WriteJSON(w, http.StatusCreated, chain)
}

func HandleUpdateChain(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/api/chains/")

	var chain model.ApprovalChain
	if err := json.NewDecoder(r.Body).Decode(&chain); err != nil {
		WriteError(w, errors.New("invalid request body"))
		return
	}
	chain.ID = id

	if len(chain.Nodes) == 0 {
		WriteError(w, service.ErrChainEmpty)
		return
	}

	for _, node := range chain.Nodes {
		if node.Condition != "" {
			if err := service.ValidateExpression(node.Condition); err != nil {
				WriteError(w, err)
				return
			}
		}
	}

	if err := store.UpdateChain(&chain); err != nil {
		WriteError(w, err)
		return
	}

	WriteJSON(w, http.StatusOK, chain)
}

func HandleListChains(w http.ResponseWriter, r *http.Request) {
	chains, err := store.ListChains()
	if err != nil {
		WriteError(w, err)
		return
	}

	WriteJSON(w, http.StatusOK, chains)
}

func HandleGetChain(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/api/chains/")
	chain, err := store.GetChainByID(id)
	if err != nil {
		WriteError(w, err)
		return
	}

	WriteJSON(w, http.StatusOK, chain)
}

func HandleDeleteChain(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/api/chains/")
	if err := store.DeleteChain(id); err != nil {
		WriteError(w, err)
		return
	}

	WriteJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

func HandleSubmitApplication(w http.ResponseWriter, r *http.Request) {
	user, ok := GetUserFromContext(r.Context())
	if !ok {
		WriteError(w, store.ErrUserNotFound)
		return
	}

	var app model.Application
	if err := json.NewDecoder(r.Body).Decode(&app); err != nil {
		WriteError(w, errors.New("invalid request body"))
		return
	}

	app.ApplicantID = user.ID

	result, err := approvalService.SubmitApplication(&app)
	if err != nil {
		WriteError(w, err)
		return
	}

	WriteJSON(w, http.StatusCreated, result)
}

func HandleResubmitApplication(w http.ResponseWriter, r *http.Request) {
	user, ok := GetUserFromContext(r.Context())
	if !ok {
		WriteError(w, store.ErrUserNotFound)
		return
	}

	id := strings.TrimPrefix(strings.TrimSuffix(r.URL.Path, "/resubmit"), "/api/applications/")

	var updates model.Application
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		WriteError(w, errors.New("invalid request body"))
		return
	}

	result, err := approvalService.ResubmitApplication(id, user.ID, &updates)
	if err != nil {
		WriteError(w, err)
		return
	}

	WriteJSON(w, http.StatusOK, result)
}

func HandleListApplications(w http.ResponseWriter, r *http.Request) {
	apps, err := store.ListApplications()
	if err != nil {
		WriteError(w, err)
		return
	}

	WriteJSON(w, http.StatusOK, apps)
}

func HandleGetApplication(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/api/applications/")
	app, err := store.GetApplicationByID(id)
	if err != nil {
		WriteError(w, err)
		return
	}

	WriteJSON(w, http.StatusOK, app)
}

func HandleApproveApplication(w http.ResponseWriter, r *http.Request) {
	user, ok := GetUserFromContext(r.Context())
	if !ok {
		WriteError(w, store.ErrUserNotFound)
		return
	}

	id := strings.TrimPrefix(strings.TrimSuffix(r.URL.Path, "/approve"), "/api/applications/")

	result, err := approvalService.ApproveApplication(id, user.ID)
	if err != nil {
		WriteError(w, err)
		return
	}

	WriteJSON(w, http.StatusOK, result)
}

func HandleRejectApplication(w http.ResponseWriter, r *http.Request) {
	user, ok := GetUserFromContext(r.Context())
	if !ok {
		WriteError(w, store.ErrUserNotFound)
		return
	}

	id := strings.TrimPrefix(strings.TrimSuffix(r.URL.Path, "/reject"), "/api/applications/")

	var req struct {
		Reason string `json:"reason"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, errors.New("invalid request body"))
		return
	}

	result, err := approvalService.RejectApplication(id, user.ID, req.Reason)
	if err != nil {
		WriteError(w, err)
		return
	}

	WriteJSON(w, http.StatusOK, result)
}

func HandleTransferApplication(w http.ResponseWriter, r *http.Request) {
	user, ok := GetUserFromContext(r.Context())
	if !ok {
		WriteError(w, store.ErrUserNotFound)
		return
	}

	id := strings.TrimPrefix(strings.TrimSuffix(r.URL.Path, "/transfer"), "/api/applications/")

	var req struct {
		TargetUserID string `json:"target_user_id"`
		Reason       string `json:"reason"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, errors.New("invalid request body"))
		return
	}

	result, err := approvalService.TransferApplication(id, user.ID, req.TargetUserID, req.Reason)
	if err != nil {
		WriteError(w, err)
		return
	}

	WriteJSON(w, http.StatusOK, result)
}

func HandleGetApplicationLogs(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(strings.TrimSuffix(r.URL.Path, "/logs"), "/api/applications/")
	ops, err := store.GetOperationsByApplication(id)
	if err != nil {
		WriteError(w, err)
		return
	}

	WriteJSON(w, http.StatusOK, ops)
}

func HandleListNotifications(w http.ResponseWriter, r *http.Request) {
	user, ok := GetUserFromContext(r.Context())
	if !ok {
		WriteError(w, store.ErrUserNotFound)
		return
	}

	notifs, err := store.ListNotificationsByUser(user.ID)
	if err != nil {
		WriteError(w, err)
		return
	}

	WriteJSON(w, http.StatusOK, notifs)
}

func HandleMarkNotificationRead(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(strings.TrimSuffix(r.URL.Path, "/read"), "/api/notifications/")

	if err := store.MarkNotificationAsRead(id); err != nil {
		WriteError(w, err)
		return
	}

	WriteJSON(w, http.StatusOK, map[string]string{"status": "marked"})
}

func HandleListReports(w http.ResponseWriter, r *http.Request) {
	reports, err := store.ListReports()
	if err != nil {
		WriteError(w, err)
		return
	}

	WriteJSON(w, http.StatusOK, reports)
}

func HandleTriggerReportReconcile(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(strings.TrimSuffix(r.URL.Path, "/reconcile"), "/api/reports/")

	if err := approvalService.ReconcileReport(id); err != nil {
		WriteError(w, err)
		return
	}

	WriteJSON(w, http.StatusOK, map[string]string{"status": "reconciled"})
}
