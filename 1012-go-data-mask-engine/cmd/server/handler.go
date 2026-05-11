package main

import (
"encoding/json"
"net/http"
"strings"

"mask-engine/internal/config"
"mask-engine/internal/mask"
"mask-engine/pkg/api"
)

type Handler struct {
configManager *config.Manager
batchMasker   *mask.BatchMasker
masker        *mask.Masker
}

func NewHandler(cm *config.Manager) *Handler {
rules := cm.GetRules()
m := mask.NewMasker(rules)
return &Handler{
configManager: cm,
batchMasker:   mask.NewBatchMasker(m),
masker:        m,
}
}

func (h *Handler) refreshMasker() {
rules := h.configManager.GetRules()
h.masker = mask.NewMasker(rules)
h.batchMasker = mask.NewBatchMasker(h.masker)
}

func (h *Handler) HandleMask(w http.ResponseWriter, r *http.Request) {
if r.Method != http.MethodPost {
http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
return
}
var req api.MaskRequest
if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
http.Error(w, "Invalid request body", http.StatusBadRequest)
return
}
var overrideRules *mask.DefaultRules
if req.OverrideRules != nil {
overrideRules = convertAPIRulesToMaskRules(req.OverrideRules)
}
masked := h.batchMasker.MaskText(req.Text, overrideRules)
resp := api.MaskResponse{
Original: req.Text,
Masked:   masked,
}
w.Header().Set("Content-Type", "application/json")
json.NewEncoder(w).Encode(resp)
}

func (h *Handler) HandleMaskByType(w http.ResponseWriter, r *http.Request) {
if r.Method != http.MethodPost {
http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
return
}
var req api.MaskByTypeRequest
if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
http.Error(w, "Invalid request body", http.StatusBadRequest)
return
}

dataType := parseDataType(req.Type)
if dataType == "" {
http.Error(w, "Invalid type. Valid types: phone, idcard, bankcard, email, name", http.StatusBadRequest)
return
}

var rule *mask.MaskRule
if req.OverrideRule != nil {
rule = &mask.MaskRule{
PrefixKeep: req.OverrideRule.PrefixKeep,
SuffixKeep: req.OverrideRule.SuffixKeep,
}
}

masked := h.masker.MaskByType(dataType, req.Value, rule)
resp := api.MaskByTypeResponse{
Original: req.Value,
Type:     req.Type,
Masked:   masked,
}
w.Header().Set("Content-Type", "application/json")
json.NewEncoder(w).Encode(resp)
}

func parseDataType(t string) mask.DataType {
switch strings.ToLower(t) {
case "phone":
return mask.Phone
case "idcard":
return mask.IDCard
case "bankcard":
return mask.BankCard
case "email":
return mask.Email
case "name":
return mask.Name
default:
return ""
}
}

func (h *Handler) HandleRules(w http.ResponseWriter, r *http.Request) {
switch r.Method {
case http.MethodGet:
h.handleGetRules(w, r)
case http.MethodPut:
h.handleUpdateRules(w, r)
default:
http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
}
}

func (h *Handler) handleGetRules(w http.ResponseWriter, r *http.Request) {
rules := h.configManager.GetRules()
resp := api.RulesResponse{
Rules: convertMaskRulesToAPIRules(rules),
}
w.Header().Set("Content-Type", "application/json")
json.NewEncoder(w).Encode(resp)
}

func (h *Handler) handleUpdateRules(w http.ResponseWriter, r *http.Request) {
var req api.UpdateRulesRequest
if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
http.Error(w, "Invalid request body", http.StatusBadRequest)
return
}
maskRules := convertAPIRulesToMaskRules(&req)
if err := h.configManager.UpdateRules(maskRules); err != nil {
http.Error(w, "Failed to update rules: "+err.Error(), http.StatusInternalServerError)
return
}
h.refreshMasker()
resp := api.UpdateRulesResponse{Success: true}
w.Header().Set("Content-Type", "application/json")
json.NewEncoder(w).Encode(resp)
}

func convertAPIRulesToMaskRules(apiRules *api.Rules) *mask.DefaultRules {
if apiRules == nil {
return nil
}
r := &mask.DefaultRules{}
if apiRules.Phone != nil {
r.Phone = &mask.MaskRule{PrefixKeep: apiRules.Phone.PrefixKeep, SuffixKeep: apiRules.Phone.SuffixKeep}
}
if apiRules.IDCard != nil {
r.IDCard = &mask.MaskRule{PrefixKeep: apiRules.IDCard.PrefixKeep, SuffixKeep: apiRules.IDCard.SuffixKeep}
}
if apiRules.BankCard != nil {
r.BankCard = &mask.MaskRule{PrefixKeep: apiRules.BankCard.PrefixKeep, SuffixKeep: apiRules.BankCard.SuffixKeep}
}
if apiRules.Email != nil {
r.Email = &mask.MaskRule{PrefixKeep: apiRules.Email.PrefixKeep, SuffixKeep: apiRules.Email.SuffixKeep}
}
if apiRules.Name != nil {
r.Name = &mask.MaskRule{PrefixKeep: apiRules.Name.PrefixKeep, SuffixKeep: apiRules.Name.SuffixKeep}
}
return r
}

func convertMaskRulesToAPIRules(maskRules *mask.DefaultRules) *api.Rules {
if maskRules == nil {
return nil
}
r := &api.Rules{}
if maskRules.Phone != nil {
r.Phone = &api.MaskRule{PrefixKeep: maskRules.Phone.PrefixKeep, SuffixKeep: maskRules.Phone.SuffixKeep}
}
if maskRules.IDCard != nil {
r.IDCard = &api.MaskRule{PrefixKeep: maskRules.IDCard.PrefixKeep, SuffixKeep: maskRules.IDCard.SuffixKeep}
}
if maskRules.BankCard != nil {
r.BankCard = &api.MaskRule{PrefixKeep: maskRules.BankCard.PrefixKeep, SuffixKeep: maskRules.BankCard.SuffixKeep}
}
if maskRules.Email != nil {
r.Email = &api.MaskRule{PrefixKeep: maskRules.Email.PrefixKeep, SuffixKeep: maskRules.Email.SuffixKeep}
}
if maskRules.Name != nil {
r.Name = &api.MaskRule{PrefixKeep: maskRules.Name.PrefixKeep, SuffixKeep: maskRules.Name.SuffixKeep}
}
return r
}
