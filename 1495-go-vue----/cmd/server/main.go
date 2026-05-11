package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"pest-control/core"

	"github.com/gorilla/mux"
)

func main() {
	var port string
	flag.StringVar(&port, "port", "", "Server port (e.g., :8080)")
	flag.Parse()

	if port == "" {
		port = os.Getenv("PEST_CONTROL_PORT")
		if port == "" {
			port = ":9015"
		}
	}
	if port[0] != ':' {
		port = ":" + port
	}

	store := core.NewInMemoryStore()
	service := core.NewService(store)
	server := NewServer(service)

	seedData(store, service)

	r := mux.NewRouter()

	r.HandleFunc("/customers", server.ListCustomers).Methods("GET")
	r.HandleFunc("/customers", server.CreateCustomer).Methods("POST")

	r.HandleFunc("/contracts", server.ListContracts).Methods("GET")
	r.HandleFunc("/contracts", server.CreateContract).Methods("POST")

	r.HandleFunc("/contracts/{id}/control-points", server.GetControlPoints).Methods("GET")
	r.HandleFunc("/control-points", server.CreateControlPoint).Methods("POST")

	r.HandleFunc("/staff", server.ListStaff).Methods("GET")
	r.HandleFunc("/staff", server.CreateStaff).Methods("POST")

	r.HandleFunc("/chemicals", server.ListChemicals).Methods("GET")
	r.HandleFunc("/chemicals", server.CreateChemical).Methods("POST")

	r.HandleFunc("/inspections", server.CreateInspection).Methods("POST")
	r.HandleFunc("/contracts/{id}/inspections", server.ListInspections).Methods("GET")

	r.HandleFunc("/operations", server.CreateOperation).Methods("POST")
	r.HandleFunc("/contracts/{id}/operations", server.ListOperations).Methods("GET")

	r.HandleFunc("/inspections/{id}/evaluate", server.EvaluateEffect).Methods("POST")

	r.HandleFunc("/pending-services", server.GetPendingServices).Methods("GET")

	fmt.Printf("Pest Control Server starting on %s...\n", port)
	log.Fatal(http.ListenAndServe(port, r))
}

func seedData(store *core.InMemoryStore, service *core.Service) {
	customer, _ := service.CreateCustomer(
		"美味餐厅",
		"北京市朝阳区建国路88号",
		core.PlaceTypeRestaurant,
		200.0,
		[]core.PestType{core.PestTypeCockroach, core.PestTypeRat},
	)

	_ = customer
}
