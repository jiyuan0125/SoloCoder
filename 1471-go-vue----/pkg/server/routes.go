package server

import (
	"log"
	"net/http"
)

func (s *Server) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/api/schedules/create", http.HandlerFunc(s.createScheduleHandler))
	mux.HandleFunc("/api/schedules/get", http.HandlerFunc(s.getScheduleHandler))
	mux.HandleFunc("/api/schedules/list", http.HandlerFunc(s.listSchedulesHandler))
	mux.HandleFunc("/api/schedules/search", http.HandlerFunc(s.searchSchedulesHandler))
	mux.HandleFunc("/api/schedules/update-status", http.HandlerFunc(s.updateScheduleStatusHandler))
	mux.HandleFunc("/api/schedules/cancel", http.HandlerFunc(s.cancelScheduleHandler))
	mux.HandleFunc("/api/schedules/seats", http.HandlerFunc(s.getSeatsHandler))

	mux.HandleFunc("/api/tickets/purchase", http.HandlerFunc(s.purchaseTicketsHandler))
	mux.HandleFunc("/api/tickets/get", http.HandlerFunc(s.getTicketHandler))

	mux.HandleFunc("/api/checkin", http.HandlerFunc(s.checkInHandler))

	mux.HandleFunc("/api/refunds/request", http.HandlerFunc(s.requestRefundHandler))
	mux.HandleFunc("/api/refunds/process", http.HandlerFunc(s.processRefundHandler))
	mux.HandleFunc("/api/refunds/pending", http.HandlerFunc(s.listPendingRefundsHandler))
}

func (s *Server) Start(addr string) error {
	mux := http.NewServeMux()
	s.RegisterRoutes(mux)

	log.Printf("客运站管理系统服务端启动，监听地址: %s", addr)
	return http.ListenAndServe(addr, mux)
}
