package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	_ "github.com/mattn/go-sqlite3"
	"mini-orm/common"
	"mini-orm/orm"
)

var (
	ormDB *orm.DB
)

func initDB() error {
	sqlDB, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		return err
	}

	_, err = sqlDB.Exec(`
		CREATE TABLE users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT,
			email TEXT,
			age INTEGER,
			active INTEGER,
			created_at TEXT
		)
	`)
	if err != nil {
		return err
	}

	ormDB = orm.NewDB(sqlDB)
	return nil
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func insertHandler(w http.ResponseWriter, r *http.Request) {
	var req common.InsertRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, common.InsertResponse{
			Success: false,
			Error:   "invalid request body",
		})
		return
	}

	var err error
	switch req.Model {
	case "user":
		user := &common.User{}
		if name, ok := req.Data["name"].(string); ok {
			user.Name = name
		}
		if email, ok := req.Data["email"].(string); ok {
			user.Email = email
		}
		if age, ok := req.Data["age"].(float64); ok {
			user.Age = int(age)
		}
		if active, ok := req.Data["active"].(bool); ok {
			user.Active = active
		}
		err = ormDB.Insert(user)
	default:
		err = fmt.Errorf("unknown model: %s", req.Model)
	}

	if err != nil {
		writeJSON(w, http.StatusInternalServerError, common.InsertResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	writeJSON(w, http.StatusOK, common.InsertResponse{
		Success: true,
	})
}

func getByIDHandler(w http.ResponseWriter, r *http.Request) {
	var req common.GetByIDRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, common.GetByIDResponse{
			Success: false,
			Error:   "invalid request body",
		})
		return
	}

	var data map[string]interface{}
	var err error

	switch req.Model {
	case "user":
		user := &common.User{}
		err = ormDB.GetByID(user, req.ID)
		if err == nil {
			data = map[string]interface{}{
				"id":         user.ID,
				"name":       user.Name,
				"email":      user.Email,
				"age":        user.Age,
				"active":     user.Active,
				"created_at": user.CreatedAt,
			}
		}
	default:
		err = fmt.Errorf("unknown model: %s", req.Model)
	}

	if err != nil {
		writeJSON(w, http.StatusInternalServerError, common.GetByIDResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	writeJSON(w, http.StatusOK, common.GetByIDResponse{
		Success: true,
		Data:    data,
	})
}

func updateHandler(w http.ResponseWriter, r *http.Request) {
	var req common.UpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, common.UpdateResponse{
			Success: false,
			Error:   "invalid request body",
		})
		return
	}

	var err error
	switch req.Model {
	case "user":
		user := &common.User{}
		if id, ok := req.ID.(float64); ok {
			user.ID = int64(id)
		}
		err = ormDB.Update(user, req.Data)
	default:
		err = fmt.Errorf("unknown model: %s", req.Model)
	}

	if err != nil {
		writeJSON(w, http.StatusInternalServerError, common.UpdateResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	writeJSON(w, http.StatusOK, common.UpdateResponse{
		Success: true,
	})
}

func deleteHandler(w http.ResponseWriter, r *http.Request) {
	var req common.DeleteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, common.DeleteResponse{
			Success: false,
			Error:   "invalid request body",
		})
		return
	}

	var err error
	switch req.Model {
	case "user":
		err = ormDB.Delete(&common.User{}, req.ID)
	default:
		err = fmt.Errorf("unknown model: %s", req.Model)
	}

	if err != nil {
		writeJSON(w, http.StatusInternalServerError, common.DeleteResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	writeJSON(w, http.StatusOK, common.DeleteResponse{
		Success: true,
	})
}

func listHandler(w http.ResponseWriter, r *http.Request) {
	var req common.ListRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, common.ListResponse{
			Success: false,
			Error:   "invalid request body",
		})
		return
	}

	var data []map[string]interface{}
	var err error

	options := &orm.ListOptions{
		Where:   req.Where,
		OrderBy: req.OrderBy,
		Limit:   req.Limit,
		Offset:  req.Offset,
	}

	switch req.Model {
	case "user":
		var users []common.User
		err = ormDB.List(&users, options)
		if err == nil {
			data = make([]map[string]interface{}, len(users))
			for i, user := range users {
				data[i] = map[string]interface{}{
					"id":         user.ID,
					"name":       user.Name,
					"email":      user.Email,
					"age":        user.Age,
					"active":     user.Active,
					"created_at": user.CreatedAt,
				}
			}
		}
	default:
		err = fmt.Errorf("unknown model: %s", req.Model)
	}

	if err != nil {
		writeJSON(w, http.StatusInternalServerError, common.ListResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	writeJSON(w, http.StatusOK, common.ListResponse{
		Success: true,
		Data:    data,
	})
}

func main() {
	if err := initDB(); err != nil {
		log.Fatalf("failed to init db: %v", err)
	}

	http.HandleFunc("/insert", insertHandler)
	http.HandleFunc("/get", getByIDHandler)
	http.HandleFunc("/update", updateHandler)
	http.HandleFunc("/delete", deleteHandler)
	http.HandleFunc("/list", listHandler)

	log.Println("server starting on :8204")
	if err := http.ListenAndServe(":8204", nil); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
