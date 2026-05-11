package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"strings"

	"dist-lock/pkg/api"
	"dist-lock/pkg/lock"
)

const defaultPort = "8080"

func main() {
	var port string
	flag.StringVar(&port, "port", "", "服务端监听端口")
	flag.Parse()

	if port == "" {
		port = os.Getenv("PORT")
	}
	if port == "" {
		port = defaultPort
	}

	lockManager := lock.NewLockManager()

	http.HandleFunc("/locks/acquire", handleAcquire(lockManager))
	http.HandleFunc("/locks/release", handleRelease(lockManager))
	http.HandleFunc("/locks/renew", handleRenew(lockManager))
	http.HandleFunc("/locks/", handleStatus(lockManager))
	http.HandleFunc("/locks", handleList(lockManager))

	fmt.Printf("锁服务启动，监听端口: %s\n", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		fmt.Printf("服务启动失败: %v\n", err)
		os.Exit(1)
	}
}

func writeJSON(w http.ResponseWriter, status int, resp *api.Response) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(resp)
}

func handleAcquire(lm *lock.LockManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			writeJSON(w, http.StatusMethodNotAllowed, api.NewErrorResponse("方法不允许"))
			return
		}

		var req api.AcquireRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSON(w, http.StatusBadRequest, api.NewErrorResponse("请求体解析失败"))
			return
		}
		defer r.Body.Close()

		if req.Name == "" || req.HolderID == "" {
			writeJSON(w, http.StatusBadRequest, api.NewErrorResponse("缺少必要参数"))
			return
		}

		success := lm.Acquire(req.Name, req.HolderID, req.ExpireMs)
		if success {
			writeJSON(w, http.StatusOK, api.NewSuccessResponse(map[string]bool{"acquired": true}))
		} else {
			writeJSON(w, http.StatusConflict, api.NewErrorResponse("获取锁失败，锁已被占用"))
		}
	}
}

func handleRelease(lm *lock.LockManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			writeJSON(w, http.StatusMethodNotAllowed, api.NewErrorResponse("方法不允许"))
			return
		}

		var req api.ReleaseRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSON(w, http.StatusBadRequest, api.NewErrorResponse("请求体解析失败"))
			return
		}
		defer r.Body.Close()

		if req.Name == "" || req.HolderID == "" {
			writeJSON(w, http.StatusBadRequest, api.NewErrorResponse("缺少必要参数"))
			return
		}

		success := lm.Release(req.Name, req.HolderID)
		if success {
			writeJSON(w, http.StatusOK, api.NewSuccessResponse(map[string]bool{"released": true}))
		} else {
			writeJSON(w, http.StatusNotFound, api.NewErrorResponse("释放锁失败，锁不存在或不是当前持有者"))
		}
	}
}

func handleRenew(lm *lock.LockManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			writeJSON(w, http.StatusMethodNotAllowed, api.NewErrorResponse("方法不允许"))
			return
		}

		var req api.RenewRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSON(w, http.StatusBadRequest, api.NewErrorResponse("请求体解析失败"))
			return
		}
		defer r.Body.Close()

		if req.Name == "" || req.HolderID == "" {
			writeJSON(w, http.StatusBadRequest, api.NewErrorResponse("缺少必要参数"))
			return
		}

		success := lm.Renew(req.Name, req.HolderID)
		if success {
			writeJSON(w, http.StatusOK, api.NewSuccessResponse(map[string]bool{"renewed": true}))
		} else {
			writeJSON(w, http.StatusNotFound, api.NewErrorResponse("续期失败，锁不存在或不是当前持有者"))
		}
	}
}

func handleStatus(lm *lock.LockManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			writeJSON(w, http.StatusMethodNotAllowed, api.NewErrorResponse("方法不允许"))
			return
		}

		name := strings.TrimPrefix(r.URL.Path, "/locks/")
		if name == "" {
			writeJSON(w, http.StatusNotFound, api.NewErrorResponse("锁名称为空"))
			return
		}

		lockInfo, remainingMs, exists := lm.Status(name)
		if !exists {
			writeJSON(w, http.StatusNotFound, api.NewErrorResponse("锁不存在或已过期"))
			return
		}

		info := api.LockInfo{
			Name:          lockInfo.Name,
			HolderID:      lockInfo.HolderID,
			ExpireMs:      lockInfo.ExpireMs,
			CreatedAt:     lockInfo.CreatedAt,
			LastRenewedAt: lockInfo.LastRenewedAt,
			RemainingMs:   remainingMs,
		}
		writeJSON(w, http.StatusOK, api.NewSuccessResponse(info))
	}
}

func handleList(lm *lock.LockManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			writeJSON(w, http.StatusMethodNotAllowed, api.NewErrorResponse("方法不允许"))
			return
		}

		locks := lm.List()
		lockInfos := make([]api.LockInfo, 0, len(locks))

		for _, lockInfo := range locks {
			_, remainingMs, _ := lm.Status(lockInfo.Name)
			info := api.LockInfo{
				Name:          lockInfo.Name,
				HolderID:      lockInfo.HolderID,
				ExpireMs:      lockInfo.ExpireMs,
				CreatedAt:     lockInfo.CreatedAt,
				LastRenewedAt: lockInfo.LastRenewedAt,
				RemainingMs:   remainingMs,
			}
			lockInfos = append(lockInfos, info)
		}

		writeJSON(w, http.StatusOK, api.NewSuccessResponse(lockInfos))
	}
}
