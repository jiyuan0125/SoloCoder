package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"sync"

	"walsh-mining/common"
	"walsh-mining/core"
)

type ServerState struct {
	mu              sync.RWMutex
	transactions    []core.ItemSet
	weights         map[string]float64
	miningResult    *core.MiningResult
	hasMined        bool
}

func NewServerState() *ServerState {
	return &ServerState{
		transactions: []core.ItemSet{},
		weights:      make(map[string]float64),
		hasMined:     false,
	}
}

var state = NewServerState()

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, err error) {
	writeJSON(w, status, common.ErrorResponse{
		Success: false,
		Error:   err.Error(),
	})
}

func importHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, fmt.Errorf("method not allowed"))
		return
	}

	var req common.ImportRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	state.mu.Lock()
	defer state.mu.Unlock()

	state.transactions = make([]core.ItemSet, 0, len(req.Transactions))
	for _, items := range req.Transactions {
		state.transactions = append(state.transactions, core.NewItemSet(items))
	}
	state.hasMined = false
	state.miningResult = nil

	writeJSON(w, http.StatusOK, common.ImportResponse{
		Success:          true,
		Message:          fmt.Sprintf("成功导入 %d 条交易记录", len(req.Transactions)),
		TransactionsCount: len(req.Transactions),
	})
}

func setWeightsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, fmt.Errorf("method not allowed"))
		return
	}

	var req common.SetWeightsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	for item, weightVal := range req.Weights {
		if weightVal < 0 || weightVal > 1 {
			writeError(w, http.StatusBadRequest, fmt.Errorf("商品 %s 的权重 %.2f 超出范围 [0, 1]", item, weightVal))
			return
		}
	}

	state.mu.Lock()
	defer state.mu.Unlock()

	if state.weights == nil {
		state.weights = make(map[string]float64)
	}
	for item, weightVal := range req.Weights {
		state.weights[item] = weightVal
	}
	state.hasMined = false
	state.miningResult = nil

	writeJSON(w, http.StatusOK, common.SetWeightsResponse{
		Success: true,
		Message: fmt.Sprintf("成功设置 %d 个商品的权重", len(req.Weights)),
	})
}

func mineHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, fmt.Errorf("method not allowed"))
		return
	}

	var req common.MineRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	state.mu.Lock()
	defer state.mu.Unlock()

	miner := core.NewWALSMiner(state.transactions, state.weights)
	result, err := miner.Mine(req.MinSup, req.MinConf)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	state.miningResult = result
	state.hasMined = true

	topRules := result.Rules
	if len(topRules) > 20 {
		sort.Slice(topRules, func(i, j int) bool {
			return topRules[i].Confidence > topRules[j].Confidence
		})
		topRules = topRules[:20]
	} else {
		sort.Slice(topRules, func(i, j int) bool {
			return topRules[i].Confidence > topRules[j].Confidence
		})
	}

	ruleDataList := make([]*common.RuleData, 0, len(topRules))
	for _, rule := range topRules {
		ruleDataList = append(ruleDataList, &common.RuleData{
			Antecedent:      rule.Antecedent.Items(),
			Consequent:      rule.Consequent.Items(),
			Confidence:      core.RoundTo(rule.Confidence, 4),
			WeightedSupport: core.RoundTo(rule.WeightedSupport, 4),
		})
	}

	writeJSON(w, http.StatusOK, common.MineResponse{
		Success:       true,
		FrequentCount: len(result.FrequentItemSets),
		RulesCount:    len(result.Rules),
		Stats: &common.MiningStatsData{
			ScansCount:      result.Stats.ScansCount,
			CandidatesCount: result.Stats.CandidatesCount,
			FrequentCount:   result.Stats.FrequentCount,
		},
		TopRules: ruleDataList,
	})
}

func itemsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, fmt.Errorf("method not allowed"))
		return
	}

	state.mu.RLock()
	defer state.mu.RUnlock()

	if !state.hasMined {
		writeError(w, http.StatusBadRequest, fmt.Errorf("尚未执行挖掘，请先调用 /mine"))
		return
	}

	fisDataList := make([]*common.FrequentItemSetData, 0, len(state.miningResult.FrequentItemSets))
	for _, fis := range state.miningResult.FrequentItemSets {
		fisDataList = append(fisDataList, &common.FrequentItemSetData{
			Items:           fis.Items.Items(),
			WeightedSupport: core.RoundTo(fis.WeightedSupport, 4),
			RawCount:        fis.RawCount,
		})
	}

	writeJSON(w, http.StatusOK, common.FrequentItemSetResponse{
		Success:          true,
		FrequentItemSets: fisDataList,
	})
}

func rulesHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, fmt.Errorf("method not allowed"))
		return
	}

	state.mu.RLock()
	defer state.mu.RUnlock()

	if !state.hasMined {
		writeError(w, http.StatusBadRequest, fmt.Errorf("尚未执行挖掘，请先调用 /mine"))
		return
	}

	sortedRules := make([]*core.AssociationRule, len(state.miningResult.Rules))
	copy(sortedRules, state.miningResult.Rules)
	sort.Slice(sortedRules, func(i, j int) bool {
		return sortedRules[i].Confidence > sortedRules[j].Confidence
	})

	ruleDataList := make([]*common.RuleData, 0, len(sortedRules))
	for _, rule := range sortedRules {
		ruleDataList = append(ruleDataList, &common.RuleData{
			Antecedent:      rule.Antecedent.Items(),
			Consequent:      rule.Consequent.Items(),
			Confidence:      core.RoundTo(rule.Confidence, 4),
			WeightedSupport: core.RoundTo(rule.WeightedSupport, 4),
		})
	}

	writeJSON(w, http.StatusOK, common.RulesResponse{
		Success: true,
		Rules:   ruleDataList,
	})
}

func statsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, fmt.Errorf("method not allowed"))
		return
	}

	state.mu.RLock()
	defer state.mu.RUnlock()

	if !state.hasMined {
		writeError(w, http.StatusBadRequest, fmt.Errorf("尚未执行挖掘，请先调用 /mine"))
		return
	}

	writeJSON(w, http.StatusOK, common.StatsResponse{
		Success: true,
		Stats: &common.MiningStatsData{
			ScansCount:      state.miningResult.Stats.ScansCount,
			CandidatesCount: state.miningResult.Stats.CandidatesCount,
			FrequentCount:   state.miningResult.Stats.FrequentCount,
		},
	})
}
