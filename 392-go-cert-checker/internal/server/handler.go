package server

import (
	"encoding/json"
	"log"

	"cert-checker/internal/certchecker"
	"cert-checker/internal/protocol"
)

type RequestHandler struct {
	manager *DomainManager
}

func NewRequestHandler(manager *DomainManager) *RequestHandler {
	return &RequestHandler{
		manager: manager,
	}
}

func (h *RequestHandler) HandleRequest(req protocol.Request) protocol.Response {
	switch req.Type {
	case protocol.RequestTypeAddDomain:
		return h.handleAddDomain(req)
	case protocol.RequestTypeRemoveDomain:
		return h.handleRemoveDomain(req)
	case protocol.RequestTypeListDomains:
		return h.handleListDomains()
	case protocol.RequestTypeCheckDomain:
		return h.handleCheckDomain(req)
	case protocol.RequestTypeCheckAll:
		return h.handleCheckAll(req)
	case protocol.RequestTypeGetHistory:
		return h.handleGetHistory(req)
	case protocol.RequestTypeGetStatus:
		return h.handleGetStatus(req)
	default:
		return protocol.Response{
			Type:    protocol.ResponseTypeError,
			Message: "未知的请求类型",
		}
	}
}

func (h *RequestHandler) handleAddDomain(req protocol.Request) protocol.Response {
	var payload protocol.AddDomainPayload
	if err := json.Unmarshal(req.Payload, &payload); err != nil {
		return protocol.Response{
			Type:    protocol.ResponseTypeError,
			Message: "解析请求失败",
		}
	}

	if h.manager.AddDomain(payload.Domain) {
		return protocol.Response{
			Type:    protocol.ResponseTypeSuccess,
			Message: "域名添加成功",
		}
	}

	return protocol.Response{
		Type:    protocol.ResponseTypeError,
		Message: "域名已存在",
	}
}

func (h *RequestHandler) handleRemoveDomain(req protocol.Request) protocol.Response {
	var payload protocol.RemoveDomainPayload
	if err := json.Unmarshal(req.Payload, &payload); err != nil {
		return protocol.Response{
			Type:    protocol.ResponseTypeError,
			Message: "解析请求失败",
		}
	}

	if h.manager.RemoveDomain(payload.Domain) {
		return protocol.Response{
			Type:    protocol.ResponseTypeSuccess,
			Message: "域名删除成功",
		}
	}

	return protocol.Response{
		Type:    protocol.ResponseTypeError,
		Message: "域名不存在",
	}
}

func (h *RequestHandler) handleListDomains() protocol.Response {
	domains := h.manager.ListDomains()
	response := protocol.ListDomainsResponse{
		Domains: domains,
	}

	payload, err := json.Marshal(response)
	if err != nil {
		return protocol.Response{
			Type:    protocol.ResponseTypeError,
			Message: "序列化响应失败",
		}
	}

	return protocol.Response{
		Type:    protocol.ResponseTypeSuccess,
		Payload: payload,
	}
}

func (h *RequestHandler) handleCheckDomain(req protocol.Request) protocol.Response {
	var payload protocol.CheckDomainPayload
	if err := json.Unmarshal(req.Payload, &payload); err != nil {
		return protocol.Response{
			Type:    protocol.ResponseTypeError,
			Message: "解析请求失败",
		}
	}

	log.Printf("正在检查域名: %s", payload.Domain)
	result := certchecker.CheckCertificate(payload.Domain)

	if result.Error != "" {
		log.Printf("检查域名 %s 失败: %s", payload.Domain, result.Error)
	} else {
		log.Printf("检查域名 %s 成功，剩余天数: %d", payload.Domain, result.RemainingDays)
	}

	h.manager.UpdateCheckResult(payload.Domain, result)

	responsePayload, err := json.Marshal(result)
	if err != nil {
		return protocol.Response{
			Type:    protocol.ResponseTypeError,
			Message: "序列化响应失败",
		}
	}

	return protocol.Response{
		Type:    protocol.ResponseTypeSuccess,
		Payload: responsePayload,
	}
}

func (h *RequestHandler) handleCheckAll(req protocol.Request) protocol.Response {
	var payload protocol.CheckAllPayload
	if err := json.Unmarshal(req.Payload, &payload); err != nil {
		payload = protocol.CheckAllPayload{WarnDays: 30}
	}

	domains := h.manager.ListDomains()
	results := make([]protocol.CertInfo, 0, len(domains))

	for _, domain := range domains {
		log.Printf("正在检查域名: %s", domain)
		result := certchecker.CheckCertificate(domain)

		if result.Error != "" {
			log.Printf("检查域名 %s 失败: %s", domain, result.Error)
		} else {
			log.Printf("检查域名 %s 成功，剩余天数: %d", domain, result.RemainingDays)
		}

		h.manager.UpdateCheckResult(domain, result)
		results = append(results, result)
	}

	response := protocol.CheckAllResponse{
		Results: results,
	}

	responsePayload, err := json.Marshal(response)
	if err != nil {
		return protocol.Response{
			Type:    protocol.ResponseTypeError,
			Message: "序列化响应失败",
		}
	}

	return protocol.Response{
		Type:    protocol.ResponseTypeSuccess,
		Payload: responsePayload,
	}
}

func (h *RequestHandler) handleGetHistory(req protocol.Request) protocol.Response {
	var payload protocol.GetHistoryPayload
	if err := json.Unmarshal(req.Payload, &payload); err != nil {
		return protocol.Response{
			Type:    protocol.ResponseTypeError,
			Message: "解析请求失败",
		}
	}

	history, exists := h.manager.GetHistory(payload.Domain, payload.Limit)
	if !exists {
		return protocol.Response{
			Type:    protocol.ResponseTypeError,
			Message: "域名不存在",
		}
	}

	response := protocol.GetHistoryResponse{
		History: history,
	}

	responsePayload, err := json.Marshal(response)
	if err != nil {
		return protocol.Response{
			Type:    protocol.ResponseTypeError,
			Message: "序列化响应失败",
		}
	}

	return protocol.Response{
		Type:    protocol.ResponseTypeSuccess,
		Payload: responsePayload,
	}
}

func (h *RequestHandler) handleGetStatus(req protocol.Request) protocol.Response {
	var payload protocol.AddDomainPayload
	if err := json.Unmarshal(req.Payload, &payload); err != nil {
		return protocol.Response{
			Type:    protocol.ResponseTypeError,
			Message: "解析请求失败",
		}
	}

	status, exists := h.manager.GetStatus(payload.Domain)
	if !exists {
		return protocol.Response{
			Type:    protocol.ResponseTypeError,
			Message: "域名不存在",
		}
	}

	response := protocol.GetStatusResponse{
		Status: *status,
	}

	responsePayload, err := json.Marshal(response)
	if err != nil {
		return protocol.Response{
			Type:    protocol.ResponseTypeError,
			Message: "序列化响应失败",
		}
	}

	return protocol.Response{
		Type:    protocol.ResponseTypeSuccess,
		Payload: responsePayload,
	}
}
