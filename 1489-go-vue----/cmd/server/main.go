package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	"lockservice/pkg/lockservice"
)

func main() {
	var port int
	flag.IntVar(&port, "port", 0, "listen port")
	flag.Parse()

	if port == 0 {
		if envPort := os.Getenv("PORT"); envPort != "" {
			fmt.Sscanf(envPort, "%d", &port)
		}
	}

	if port == 0 {
		port = 9009
	}

	service := lockservice.NewService()

	service.AddMaster(&lockservice.Master{
		ID:          "m1",
		Name:        "张师傅",
		Phone:       "13800138001",
		ServiceArea: "北京市朝阳区",
		SupportedLocks: []lockservice.LockType{
			lockservice.LockTypeSecurity,
			lockservice.LockTypeInterior,
		},
		Status:      lockservice.MasterStatusIdle,
		Location:    "望京",
		BookedSlots: make(map[string]map[lockservice.TimeSlot]bool),
	})

	service.AddMaster(&lockservice.Master{
		ID:          "m2",
		Name:        "李师傅",
		Phone:       "13800138002",
		ServiceArea: "北京市朝阳区",
		SupportedLocks: []lockservice.LockType{
			lockservice.LockTypePassword,
			lockservice.LockTypeFingerprint,
			lockservice.LockTypeSecurity,
		},
		Status:      lockservice.MasterStatusIdle,
		Location:    "国贸",
		BookedSlots: make(map[string]map[lockservice.TimeSlot]bool),
	})

	service.AddMaster(&lockservice.Master{
		ID:          "m3",
		Name:        "王师傅",
		Phone:       "13800138003",
		ServiceArea: "北京市海淀区",
		SupportedLocks: []lockservice.LockType{
			lockservice.LockTypeCar,
			lockservice.LockTypeSecurity,
			lockservice.LockTypeInterior,
		},
		Status:      lockservice.MasterStatusIdle,
		Location:    "中关村",
		BookedSlots: make(map[string]map[lockservice.TimeSlot]bool),
	})

	handler := NewHandler(service)

	mux := http.NewServeMux()
	mux.HandleFunc("/order/create", handler.CreateOrder)
	mux.HandleFunc("/order/accept", handler.AcceptOrder)
	mux.HandleFunc("/order/start", handler.StartService)
	mux.HandleFunc("/order/complete", handler.CompleteService)
	mux.HandleFunc("/order/pay", handler.PayOrder)
	mux.HandleFunc("/order/rate", handler.RateOrder)
	mux.HandleFunc("/order/get", handler.GetOrder)
	mux.HandleFunc("/order/list", handler.ListOrders)
	mux.HandleFunc("/master/list", handler.ListMasters)
	mux.HandleFunc("/master/add", handler.AddMaster)

	addr := fmt.Sprintf(":%d", port)
	log.Printf("server starting on %s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
