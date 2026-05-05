package main

import (
	"encoding/json"
	"net/http"

	"emoji-normalizer/emojinorm"
	"emoji-normalizer/protocol"
)

type EmojiHandler struct{}

func NewEmojiHandler() *EmojiHandler {
	return &EmojiHandler{}
}

func (h *EmojiHandler) HandleEmoji(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req protocol.EmojiRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendError(w, "Invalid request body")
		return
	}

	if req.Text == "" {
		h.sendError(w, "Text is required")
		return
	}

	if req.Operation == "" {
		h.sendError(w, "Operation is required")
		return
	}

	resp := h.processRequest(&req)
	json.NewEncoder(w).Encode(resp)
}

func (h *EmojiHandler) processRequest(req *protocol.EmojiRequest) protocol.EmojiResponse {
	switch req.Operation {
	case protocol.OpExtract:
		return h.handleExtract(req)
	case protocol.OpCount:
		return h.handleCount(req)
	case protocol.OpRemove:
		return h.handleRemove(req)
	case protocol.OpReplace:
		return h.handleReplace(req)
	case protocol.OpCheckOnly:
		return h.handleCheckOnly(req)
	case protocol.OpContains:
		return h.handleContains(req)
	default:
		return protocol.EmojiResponse{
			Success: false,
			Error:   "Unknown operation: " + req.Operation,
		}
	}
}

func (h *EmojiHandler) handleExtract(req *protocol.EmojiRequest) protocol.EmojiResponse {
	emojis := emojinorm.ExtractEmojis(req.Text)
	emojiData := make([]protocol.EmojiData, len(emojis))

	for i, e := range emojis {
		emojiData[i] = protocol.EmojiData{
			StartByte: e.StartByte,
			EndByte:   e.EndByte,
			Emoji:     e.Emoji,
		}
	}

	return protocol.EmojiResponse{
		Success:    true,
		EmojiCount: len(emojis),
		Emojis:     emojiData,
	}
}

func (h *EmojiHandler) handleCount(req *protocol.EmojiRequest) protocol.EmojiResponse {
	count := emojinorm.CountEmojis(req.Text)
	return protocol.EmojiResponse{
		Success:    true,
		EmojiCount: count,
	}
}

func (h *EmojiHandler) handleRemove(req *protocol.EmojiRequest) protocol.EmojiResponse {
	result := emojinorm.RemoveEmojis(req.Text)
	return protocol.EmojiResponse{
		Success: true,
		Result:  result,
	}
}

func (h *EmojiHandler) handleReplace(req *protocol.EmojiRequest) protocol.EmojiResponse {
	replacement := req.Replacement
	if replacement == "" {
		replacement = "[表情]"
	}
	result := emojinorm.ReplaceEmojis(req.Text, replacement)
	return protocol.EmojiResponse{
		Success: true,
		Result:  result,
	}
}

func (h *EmojiHandler) handleCheckOnly(req *protocol.EmojiRequest) protocol.EmojiResponse {
	isOnly := emojinorm.IsOnlyEmojis(req.Text)
	return protocol.EmojiResponse{
		Success:      true,
		IsOnlyEmojis: isOnly,
	}
}

func (h *EmojiHandler) handleContains(req *protocol.EmojiRequest) protocol.EmojiResponse {
	contains := emojinorm.ContainsEmoji(req.Text)
	return protocol.EmojiResponse{
		Success: true,
		Result:  func() string { if contains { return "true" } else { return "false" } }(),
	}
}

func (h *EmojiHandler) sendError(w http.ResponseWriter, message string) {
	resp := protocol.EmojiResponse{
		Success: false,
		Error:   message,
	}
	w.WriteHeader(http.StatusBadRequest)
	json.NewEncoder(w).Encode(resp)
}
