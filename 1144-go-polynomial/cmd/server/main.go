package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"polynomial-service/api"
	"polynomial-service/polynomial"
)

func main() {
	http.HandleFunc("/add", handleBinaryOp(polynomial.Add))
	http.HandleFunc("/sub", handleBinaryOp(polynomial.Sub))
	http.HandleFunc("/mul", handleBinaryOp(polynomial.Mul))
	http.HandleFunc("/div", handleDiv)
	http.HandleFunc("/eval", handleEval)
	http.HandleFunc("/factor", handleFactor)
	http.HandleFunc("/deriv", handleDeriv)

	fmt.Println("服务端启动，监听端口 8080...")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		fmt.Printf("启动失败: %v\n", err)
	}
}

func handleBinaryOp(op func(a, b polynomial.Polynomial) polynomial.Polynomial) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "仅支持POST请求", http.StatusMethodNotAllowed)
			return
		}

		var req api.BinaryOpRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			sendError(w, "请求体解析失败")
			return
		}

		a, err := polynomial.New(req.A)
		if err != nil {
			sendError(w, err.Error())
			return
		}

		b, err := polynomial.New(req.B)
		if err != nil {
			sendError(w, err.Error())
			return
		}

		result := op(a, b)
		sendPolynomialResponse(w, result)
	}
}

func handleDiv(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "仅支持POST请求", http.StatusMethodNotAllowed)
		return
	}

	var req api.BinaryOpRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, "请求体解析失败")
		return
	}

	dividend, err := polynomial.New(req.A)
	if err != nil {
		sendError(w, err.Error())
		return
	}

	divisor, err := polynomial.New(req.B)
	if err != nil {
		sendError(w, err.Error())
		return
	}

	quotient, remainder, err := polynomial.Div(dividend, divisor)
	if err != nil {
		sendError(w, err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(api.DivResponse{
		Success:   true,
		Quotient:  quotient,
		Remainder: remainder,
	})
}

func handleEval(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "仅支持POST请求", http.StatusMethodNotAllowed)
		return
	}

	var req api.EvalRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, "请求体解析失败")
		return
	}

	p, err := polynomial.New(req.Polynomial)
	if err != nil {
		sendError(w, err.Error())
		return
	}

	result := polynomial.Eval(p, req.X)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(api.EvalResponse{
		Success: true,
		Result:  result,
	})
}

func handleFactor(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "仅支持POST请求", http.StatusMethodNotAllowed)
		return
	}

	var req api.SinglePolynomialRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, "请求体解析失败")
		return
	}

	p, err := polynomial.New(req.Polynomial)
	if err != nil {
		sendError(w, err.Error())
		return
	}

	result, err := polynomial.Factor(p)
	if err != nil {
		sendError(w, err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(api.FactorResponse{
		Success: true,
		Result:  result,
	})
}

func handleDeriv(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "仅支持POST请求", http.StatusMethodNotAllowed)
		return
	}

	var req api.SinglePolynomialRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, "请求体解析失败")
		return
	}

	p, err := polynomial.New(req.Polynomial)
	if err != nil {
		sendError(w, err.Error())
		return
	}

	result := polynomial.Derivative(p)
	sendPolynomialResponse(w, result)
}

func sendError(w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)
	json.NewEncoder(w).Encode(api.PolynomialResponse{
		Success: false,
		Error:   message,
	})
}

func sendPolynomialResponse(w http.ResponseWriter, result polynomial.Polynomial) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(api.PolynomialResponse{
		Success: true,
		Result:  result,
	})
}
