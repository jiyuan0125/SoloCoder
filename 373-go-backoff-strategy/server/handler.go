package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"backoff/common"
	"backoff/retry"
)

type RetryHandler struct {
}

func NewRetryHandler() *RetryHandler {
	return &RetryHandler{}
}

func (h *RetryHandler) Health(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status": "ok",
	})
}

func (h *RetryHandler) Execute(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.sendError(w, http.StatusMethodNotAllowed, "method_not_allowed", "only POST method allowed")
		return
	}

	var req common.ExecuteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendError(w, http.StatusBadRequest, "invalid_request", "failed to parse request body")
		return
	}

	if req.TaskID == "" {
		h.sendError(w, http.StatusBadRequest, "missing_task_id", "task_id is required")
		return
	}

	if !common.IsValidTaskType(req.TaskType) {
		h.sendError(w, http.StatusBadRequest, "invalid_task_type", "unknown task type")
		return
	}

	strategy, err := h.buildStrategy(req.StrategyConfig)
	if err != nil {
		h.sendError(w, http.StatusBadRequest, "invalid_strategy", err.Error())
		return
	}

	executor, err := retry.NewExecutor(strategy)
	if err != nil {
		h.sendError(w, http.StatusBadRequest, "invalid_executor", err.Error())
		return
	}

	retryEvents := make([]common.RetryEvent, 0)
	executor = executor.WithOnRetry(func(attempt int, delay time.Duration, err error) {
		retryEvents = append(retryEvents, common.RetryEvent{
			Attempt: attempt,
			Delay:   delay,
			Error:   err.Error(),
		})
	})

	executor = executor.WithNonRetryableError(func(err error) bool {
		var nonRetryableErr *NonRetryableError
		return errors.As(err, &nonRetryableErr)
	})

	taskRunner := NewTaskRunner(req.TaskType, req.TaskPayload)
	err = executor.Execute(taskRunner.Run)

	resp := common.ExecuteResponse{
		TaskID:      req.TaskID,
		Attempts:    len(retryEvents) + 1,
		RetryEvents: retryEvents,
	}

	if err != nil {
		resp.Success = false
		resp.Error = err.Error()
		if rErr, ok := err.(*retry.RetryError); ok {
			resp.Attempts = rErr.Attempts
		}
	} else {
		resp.Success = true
		resp.Result = taskRunner.GetResult()
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *RetryHandler) buildStrategy(config common.StrategyConfig) (retry.Strategy, error) {
	if err := config.Type.Validate(); err != nil {
		return nil, err
	}

	switch config.Type {
	case common.StrategyFixedInterval:
		return retry.NewFixedIntervalBuilder().
			WithMaxRetries(config.MaxRetries).
			WithMaxWait(config.MaxWait).
			WithInitialWait(config.InitialWait).
			Build(), nil

	case common.StrategyLinear:
		return retry.NewLinearBuilder().
			WithMaxRetries(config.MaxRetries).
			WithMaxWait(config.MaxWait).
			WithInitialWait(config.InitialWait).
			WithIncrement(config.Increment).
			Build(), nil

	case common.StrategyExponential:
		builder := retry.NewExponentialBuilder().
			WithMaxRetries(config.MaxRetries).
			WithMaxWait(config.MaxWait).
			WithInitialWait(config.InitialWait)
		if config.Multiplier > 0 {
			builder = builder.WithMultiplier(config.Multiplier)
		}
		return builder.Build(), nil

	case common.StrategyExponentialJitter:
		builder := retry.NewExponentialJitterBuilder().
			WithMaxRetries(config.MaxRetries).
			WithMaxWait(config.MaxWait).
			WithInitialWait(config.InitialWait)
		if config.Multiplier > 0 {
			builder = builder.WithMultiplier(config.Multiplier)
		}
		if config.JitterFactor >= 0 && config.JitterFactor <= 1 {
			builder = builder.WithJitterFactor(config.JitterFactor)
		}
		return builder.Build(), nil

	default:
		return nil, common.ErrUnknownStrategyType
	}
}

func (h *RetryHandler) sendError(w http.ResponseWriter, status int, code, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(common.ErrorResponse{
		Code:    code,
		Message: message,
	})
}
