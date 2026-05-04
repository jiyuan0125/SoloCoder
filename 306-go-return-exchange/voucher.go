package main

import (
	"net/http"
)

type UploadVoucherRequest struct {
	ApplicationID string `json:"application_id"`
	ImageBase64   string `json:"image_base64"`
}

type VoucherHandler struct {
	storage *FileStorage
}

func (h *VoucherHandler) UploadVoucher(w http.ResponseWriter, r *http.Request) {
	var req UploadVoucherRequest
	if err := ParseJSONBody(r, &req); err != nil {
		JSONError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.ApplicationID == "" || req.ImageBase64 == "" {
		JSONError(w, http.StatusBadRequest, "application_id and image_base64 are required")
		return
	}

	_, exists := h.storage.GetApplicationByID(req.ApplicationID)
	if !exists {
		JSONError(w, http.StatusNotFound, "application not found")
		return
	}

	voucher := &Voucher{
		VoucherID:     h.storage.GenerateVoucherID(),
		ApplicationID: req.ApplicationID,
		ImageBase64:   req.ImageBase64,
		CreatedAt:     GetNowTime(),
	}

	if err := h.storage.CreateVoucher(voucher); err != nil {
		JSONError(w, http.StatusInternalServerError, "failed to upload voucher")
		return
	}

	JSONResponse(w, http.StatusCreated, voucher)
}

func (h *VoucherHandler) GetVoucher(w http.ResponseWriter, r *http.Request) {
	voucherID, ok := GetIDFromPath(r, "/api/vouchers/")
	if !ok || voucherID == "" {
		JSONError(w, http.StatusBadRequest, "invalid voucher ID")
		return
	}

	voucher, exists := h.storage.GetVoucherByID(voucherID)
	if !exists {
		JSONError(w, http.StatusNotFound, "voucher not found")
		return
	}

	JSONResponse(w, http.StatusOK, voucher)
}
