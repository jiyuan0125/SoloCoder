package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"pipeline-system/common"
	"pipeline-system/pipeline"
)

var manager = pipeline.NewManager()

func main() {
	http.HandleFunc("/pipelines", pipelinesHandler)
	http.HandleFunc("/pipelines/", pipelineDetailHandler)
	http.HandleFunc("/handlers", handlersHandler)
	http.HandleFunc("/handlers/", handlerDetailHandler)
	http.HandleFunc("/middlewares", middlewaresHandler)
	http.HandleFunc("/middlewares/", middlewareDetailHandler)
	http.HandleFunc("/execute", executeHandler)
	http.HandleFunc("/traces", tracesHandler)

	fmt.Println("Server starting on :8080")
	http.ListenAndServe(":8080", nil)
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

func pipelinesHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		names := manager.ListPipelines()
		infos := make([]common.PipelineInfo, 0, len(names))
		for _, name := range names {
			if p, ok := manager.GetPipeline(name); ok {
				infos = append(infos, common.PipelineInfo{
					Name:     name,
					Handlers: p.GetHandlerIDs(),
				})
			}
		}
		writeJSON(w, http.StatusOK, infos)
	case http.MethodPost:
		var req common.CreatePipelineRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid request body")
			return
		}
		if req.Name == "" {
			writeError(w, http.StatusBadRequest, "name is required")
			return
		}
		if err := manager.CreatePipeline(req.Name); err != nil {
			writeError(w, http.StatusConflict, err.Error())
			return
		}
		writeJSON(w, http.StatusCreated, map[string]string{"name": req.Name})
	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func pipelineDetailHandler(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Path[len("/pipelines/"):]
	if name == "" {
		writeError(w, http.StatusBadRequest, "pipeline name is required")
		return
	}

	switch r.Method {
	case http.MethodGet:
		p, ok := manager.GetPipeline(name)
		if !ok {
			writeError(w, http.StatusNotFound, "pipeline not found")
			return
		}
		writeJSON(w, http.StatusOK, common.PipelineInfo{
			Name:     name,
			Handlers: p.GetHandlerIDs(),
		})
	case http.MethodDelete:
		if err := manager.DeletePipeline(name); err != nil {
			writeError(w, http.StatusNotFound, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
	case http.MethodPut:
		var action struct {
			Action     string   `json:"action"`
			HandlerID  string   `json:"handler_id"`
			Position   int      `json:"position"`
			HandlerIDs []string `json:"handler_ids"`
		}
		if err := json.NewDecoder(r.Body).Decode(&action); err != nil {
			writeError(w, http.StatusBadRequest, "invalid request body")
			return
		}
		switch action.Action {
		case "add_handler":
			if err := manager.AddHandlerToPipeline(name, action.HandlerID, action.Position); err != nil {
				writeError(w, http.StatusBadRequest, err.Error())
				return
			}
			writeJSON(w, http.StatusOK, map[string]string{"status": "handler added"})
		case "remove_handler":
			if err := manager.RemoveHandlerFromPipeline(name, action.HandlerID); err != nil {
				writeError(w, http.StatusBadRequest, err.Error())
				return
			}
			writeJSON(w, http.StatusOK, map[string]string{"status": "handler removed"})
		case "reorder":
			if err := manager.ReorderPipelineHandlers(name, action.HandlerIDs); err != nil {
				writeError(w, http.StatusBadRequest, err.Error())
				return
			}
			writeJSON(w, http.StatusOK, map[string]string{"status": "handlers reordered"})
		default:
			writeError(w, http.StatusBadRequest, "unknown action")
		}
	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func handlersHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		handlers := manager.ListHandlers()
		infos := make([]common.HandlerInfo, 0, len(handlers))
		for _, h := range handlers {
			infos = append(infos, common.HandlerInfo{
				ID:           h.ID(),
				Name:         h.Name(),
				Type:         "handler",
				PanicHistory: manager.GetHandlerPanicHistory(h.ID()),
			})
		}
		writeJSON(w, http.StatusOK, infos)
	case http.MethodPost:
		var req common.CreateHandlerRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid request body")
			return
		}
		if req.Name == "" {
			writeError(w, http.StatusBadRequest, "name is required")
			return
		}

		id := req.ID
		if id == "" {
			id = manager.GenerateHandlerID()
		}

		var h pipeline.Handler
		switch req.Type {
		case "echo":
			h = pipeline.NewSimpleHandler(id, req.Name, func(ctx *pipeline.Context, next func()) {
				ctx.Set("echo_input", ctx.Body)
				ctx.Response.Body = fmt.Sprintf("Echo: %s", ctx.Body)
				next()
			})
		case "transform":
			h = pipeline.NewSimpleHandler(id, req.Name, func(ctx *pipeline.Context, next func()) {
				transformed := fmt.Sprintf("[TRANSFORMED] %s", ctx.Body)
				ctx.Body = transformed
				ctx.Response.Body = transformed
				next()
			})
		case "validate":
			h = pipeline.NewSimpleHandler(id, req.Name, func(ctx *pipeline.Context, next func()) {
				if ctx.Body == "" {
					ctx.Response.StatusCode = 400
					ctx.Terminate("empty body not allowed")
					return
				}
				ctx.SetMetadata("validated", "true")
				next()
			})
		case "terminate":
			h = pipeline.NewSimpleHandler(id, req.Name, func(ctx *pipeline.Context, next func()) {
				ctx.Response.Body = "Terminated by handler"
				ctx.Terminate("handler requested termination")
			})
		default:
			h = pipeline.NewSimpleHandler(id, req.Name, func(ctx *pipeline.Context, next func()) {
				ctx.SetMetadata("processed_by_"+id, "true")
				next()
			})
		}

		if err := manager.RegisterHandler(h); err != nil {
			writeError(w, http.StatusConflict, err.Error())
			return
		}
		writeJSON(w, http.StatusCreated, common.HandlerInfo{
			ID:           h.ID(),
			Name:         h.Name(),
			Type:         req.Type,
			PanicHistory: []string{},
		})
	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func handlerDetailHandler(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Path[len("/handlers/"):]
	if id == "" {
		writeError(w, http.StatusBadRequest, "handler id is required")
		return
	}

	switch r.Method {
	case http.MethodGet:
		h, ok := manager.GetHandler(id)
		if !ok {
			writeError(w, http.StatusNotFound, "handler not found")
			return
		}
		writeJSON(w, http.StatusOK, common.HandlerInfo{
			ID:           h.ID(),
			Name:         h.Name(),
			Type:         "handler",
			PanicHistory: manager.GetHandlerPanicHistory(id),
		})
	case http.MethodDelete:
		if err := manager.DeleteHandler(id); err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func middlewaresHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		mws := manager.ListMiddlewares()
		infos := make([]common.MiddlewareInfo, 0, len(mws))
		for _, mw := range mws {
			infos = append(infos, common.MiddlewareInfo{
				ID:        mw.ID(),
				Name:      mw.Name(),
				Type:      "middleware",
				Async:     mw.IsAsync(),
				TimeoutMs: mw.Timeout().Milliseconds(),
			})
		}
		writeJSON(w, http.StatusOK, infos)
	case http.MethodPost:
		var req common.CreateMiddlewareRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid request body")
			return
		}
		if req.Name == "" {
			writeError(w, http.StatusBadRequest, "name is required")
			return
		}

		id := req.ID
		if id == "" {
			id = manager.GenerateHandlerID()
		}

		var mw pipeline.Middleware
		if req.Async {
			timeoutMs := req.TimeoutMs
			if timeoutMs <= 0 {
				timeoutMs = 5000
			}
			mw = pipeline.NewAsyncMiddleware(id, req.Name, timeoutMs, func(ctx *pipeline.Context) error {
				ctx.SetMetadata("async_auth_"+id, "passed")
				time.Sleep(100 * time.Millisecond)
				return nil
			})
		} else {
			mw = pipeline.NewSyncMiddleware(id, req.Name,
				func(ctx *pipeline.Context) {
					ctx.SetMetadata("middleware_start_"+id, fmt.Sprintf("%d", time.Now().UnixNano()/1e6))
				},
				func(ctx *pipeline.Context) {
					ctx.SetMetadata("middleware_end_"+id, fmt.Sprintf("%d", time.Now().UnixNano()/1e6))
				},
			)
		}

		if err := manager.RegisterMiddleware(mw); err != nil {
			writeError(w, http.StatusConflict, err.Error())
			return
		}
		writeJSON(w, http.StatusCreated, common.MiddlewareInfo{
			ID:        mw.ID(),
			Name:      mw.Name(),
			Type:      "middleware",
			Async:     mw.IsAsync(),
			TimeoutMs: mw.Timeout().Milliseconds(),
		})
	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func middlewareDetailHandler(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Path[len("/middlewares/"):]
	if id == "" {
		writeError(w, http.StatusBadRequest, "middleware id is required")
		return
	}

	switch r.Method {
	case http.MethodGet:
		mw, ok := manager.GetMiddleware(id)
		if !ok {
			writeError(w, http.StatusNotFound, "middleware not found")
			return
		}
		writeJSON(w, http.StatusOK, common.MiddlewareInfo{
			ID:        mw.ID(),
			Name:      mw.Name(),
			Type:      "middleware",
			Async:     mw.IsAsync(),
			TimeoutMs: mw.Timeout().Milliseconds(),
		})
	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func executeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req common.Request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Pipeline == "" {
		writeError(w, http.StatusBadRequest, "pipeline is required")
		return
	}

	requestID := req.ID
	if requestID == "" {
		requestID = fmt.Sprintf("req_%d", time.Now().UnixNano())
	}

	ctx := pipeline.NewContext()
	ctx.RequestID = requestID
	if req.Header != nil {
		for k, v := range req.Header {
			ctx.Header[k] = v
		}
	}
	ctx.Body = req.Body

	result, err := manager.ExecutePipeline(req.Pipeline, ctx)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	handlerResults := make([]common.HandlerResult, 0, len(result.HandlerResults))
	for _, hr := range result.HandlerResults {
		handlerResults = append(handlerResults, common.HandlerResult{
			HandlerID:   hr.HandlerID,
			HandlerName: hr.HandlerName,
			Status:      hr.Status,
			StartTime:   hr.StartTime,
			EndTime:     hr.EndTime,
			DurationMs:  hr.DurationMs,
			Output:      hr.Output,
			Error:       hr.Error,
			Panic:       hr.Panic,
		})
	}

	execResult := common.ExecutionResult{
		RequestID:      result.RequestID,
		Pipeline:       result.Pipeline,
		Status:         result.Status,
		StartTime:      result.StartTime,
		EndTime:        result.EndTime,
		DurationMs:     result.DurationMs,
		HandlerResults: handlerResults,
		Response: &common.Response{
			StatusCode: result.Response.StatusCode,
			Header:     result.Response.Header,
			Body:       result.Response.Body,
		},
		Error: result.Error,
	}

	writeJSON(w, http.StatusOK, execResult)
}

func tracesHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	requestID := r.URL.Query().Get("request_id")
	startTimeStr := r.URL.Query().Get("start_time")
	endTimeStr := r.URL.Query().Get("end_time")

	var startTime, endTime int64
	if startTimeStr != "" {
		fmt.Sscanf(startTimeStr, "%d", &startTime)
	}
	if endTimeStr != "" {
		fmt.Sscanf(endTimeStr, "%d", &endTime)
	}

	traces := manager.GetTraces(requestID, startTime, endTime)
	writeJSON(w, http.StatusOK, traces)
}
