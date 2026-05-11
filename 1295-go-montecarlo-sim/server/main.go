package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"strconv"

	"montecarlo-sim/common"
	"montecarlo-sim/montecarlo"
)

func main() {
	port := flag.String("port", "", "服务器监听端口（默认8080）")
	flag.Parse()

	if *port == "" {
		if envPort := os.Getenv("PORT"); envPort != "" {
			*port = envPort
		} else {
			*port = "8104"
		}
	}

	if _, err := strconv.Atoi(*port); err != nil {
		fmt.Printf("无效的端口号: %s\n", *port)
		os.Exit(1)
	}

	http.HandleFunc("/simulate", handleSimulate)
	http.HandleFunc("/health", handleHealth)

	addr := ":" + *port
	fmt.Printf("蒙特卡洛模拟服务端已启动，监听 %s\n", addr)
	fmt.Printf("端点: POST /simulate, GET /health\n")
	if err := http.ListenAndServe(addr, nil); err != nil {
		fmt.Printf("服务端启动失败: %v\n", err)
		os.Exit(1)
	}
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func handleSimulate(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(common.SimulationResponse{
			Success: false,
			Message: "仅支持POST方法",
		})
		return
	}

	var req common.SimulationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(common.SimulationResponse{
			Success: false,
			Message: "请求体解析失败: " + err.Error(),
		})
		return
	}

	if err := req.Validate(); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(common.SimulationResponse{
			Success: false,
			Message: "参数验证失败: " + err.Error(),
		})
		return
	}

	var result *montecarlo.SimulationResult
	var err error

	switch req.Type {
	case common.TypePi:
		result, err = montecarlo.EstimatePi(req.Pi.Samples, req.Pi.Confidence)
	case common.TypeOption:
		params := &montecarlo.OptionParams{
			S0:    req.Option.S0,
			K:     req.Option.K,
			T:     req.Option.T,
			R:     req.Option.R,
			Sigma: req.Option.Sigma,
			Mu:    req.Option.Mu,
			Type:  montecarlo.OptionType(req.Option.Type),
			Steps: req.Option.Steps,
		}
		result, err = montecarlo.PriceEuropeanOption(params, req.Option.Samples, req.Option.Confidence)
	case common.TypeProbability:
		var check montecarlo.EventCheck
		switch req.Probability.Event {
		case common.EventCoinHeads:
			check = func() bool {
				return montecarlo.RandomUniform() < 0.5
			}
		case common.EventDiceSumGreater:
			threshold := 7.0
			if v, ok := req.Probability.EventParams["threshold"]; ok {
				switch val := v.(type) {
				case float64:
					threshold = val
				}
			}
			check = func() bool {
				d1 := int(montecarlo.RandomUniform()*6) + 1
				d2 := int(montecarlo.RandomUniform()*6) + 1
				return float64(d1+d2) > threshold
			}
		default:
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(common.SimulationResponse{
				Success: false,
				Message: fmt.Sprintf("不支持的概率事件: %s", req.Probability.Event),
			})
			return
		}
		result, err = montecarlo.EstimateProbability(check, req.Probability.Samples, req.Probability.Confidence)
	}

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(common.SimulationResponse{
			Success: false,
			Message: "模拟执行失败: " + err.Error(),
		})
		return
	}

	json.NewEncoder(w).Encode(common.SimulationResponse{
		Success:  true,
		Estimate: result.Estimate,
		StdDev:   result.StdDev,
		StdErr:   result.StdErr,
		CIUpper:  result.CIUpper,
		CILower:  result.CILower,
		Samples:  result.SampleCount,
		Warnings: result.Warnings,
	})
}
