package main

import (
	"archivesystem/core"
	"net/http"
)

func setupRouter(service *core.ArchiveService) http.Handler {
	mux := http.NewServeMux()

	h := &Handler{service: service}

	mux.HandleFunc("/api/archives/create", h.CreateArchive)
	mux.HandleFunc("/api/archives/list", h.ListArchives)
	mux.HandleFunc("/api/archives/get", h.GetArchive)
	mux.HandleFunc("/api/archives/export", h.ExportArchives)

	mux.HandleFunc("/api/borrow/apply", h.ApplyBorrow)
	mux.HandleFunc("/api/borrow/approve", h.ApproveBorrow)
	mux.HandleFunc("/api/borrow/confirm", h.ConfirmBorrow)
	mux.HandleFunc("/api/borrow/return", h.ReturnArchive)
	mux.HandleFunc("/api/borrow/list", h.ListBorrowRecords)
	mux.HandleFunc("/api/borrow/overdue", h.GetOverdueReminders)
	mux.HandleFunc("/api/borrow/export", h.ExportBorrowRecords)

	mux.HandleFunc("/api/destroy/pending", h.ListPendingDestroy)
	mux.HandleFunc("/api/destroy/execute", h.DestroyArchive)
	mux.HandleFunc("/api/destroy/records", h.ListDestroyRecords)

	return mux
}
