package main

import (
	"encoding/json"
	"net/http"
	"pest-control/api"
	"pest-control/core"

	"github.com/gorilla/mux"
)

type Server struct {
	service *core.Service
}

func NewServer(service *core.Service) *Server {
	return &Server{service: service}
}

func (s *Server) Respond(w http.ResponseWriter, status int, data interface{}, err error) {
	resp := api.Response{Success: err == nil}
	if err != nil {
		resp.Error = err.Error()
	} else {
		resp.Data = data
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(resp)
}

func (s *Server) CreateCustomer(w http.ResponseWriter, r *http.Request) {
	var req api.CreateCustomerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.Respond(w, http.StatusBadRequest, nil, err)
		return
	}
	cust, err := s.service.CreateCustomer(req.Name, req.Address, req.PlaceType, req.Area, req.PestTypes)
	if err != nil {
		s.Respond(w, http.StatusInternalServerError, nil, err)
		return
	}
	s.Respond(w, http.StatusCreated, cust, nil)
}

func (s *Server) ListCustomers(w http.ResponseWriter, r *http.Request) {
	customers, err := s.service.ListCustomers()
	if err != nil {
		s.Respond(w, http.StatusInternalServerError, nil, err)
		return
	}
	s.Respond(w, http.StatusOK, api.CustomerListResponse{Customers: customers}, nil)
}

func (s *Server) CreateContract(w http.ResponseWriter, r *http.Request) {
	var req api.CreateContractRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.Respond(w, http.StatusBadRequest, nil, err)
		return
	}
	contract, err := s.service.CreateContract(req.CustomerID, req.StartDate, req.EndDate, req.ServicePerVisit)
	if err != nil {
		s.Respond(w, http.StatusInternalServerError, nil, err)
		return
	}
	s.Respond(w, http.StatusCreated, contract, nil)
}

func (s *Server) ListContracts(w http.ResponseWriter, r *http.Request) {
	contracts, err := s.service.ListContracts()
	if err != nil {
		s.Respond(w, http.StatusInternalServerError, nil, err)
		return
	}
	s.Respond(w, http.StatusOK, api.ContractListResponse{Contracts: contracts}, nil)
}

func (s *Server) CreateControlPoint(w http.ResponseWriter, r *http.Request) {
	var req api.CreateControlPointRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.Respond(w, http.StatusBadRequest, nil, err)
		return
	}
	cp, err := s.service.CreateControlPoint(req.ContractID, req.Code, req.Location, req.Description)
	if err != nil {
		s.Respond(w, http.StatusInternalServerError, nil, err)
		return
	}
	s.Respond(w, http.StatusCreated, cp, nil)
}

func (s *Server) GetControlPoints(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	contractID := vars["id"]
	points, err := s.service.GetControlPoints(contractID)
	if err != nil {
		s.Respond(w, http.StatusInternalServerError, nil, err)
		return
	}
	s.Respond(w, http.StatusOK, api.ControlPointListResponse{ControlPoints: points}, nil)
}

func (s *Server) CreateStaff(w http.ResponseWriter, r *http.Request) {
	var req api.CreateStaffRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.Respond(w, http.StatusBadRequest, nil, err)
		return
	}
	st, err := s.service.CreateStaff(req.Name, req.Role, req.Phone)
	if err != nil {
		s.Respond(w, http.StatusInternalServerError, nil, err)
		return
	}
	s.Respond(w, http.StatusCreated, st, nil)
}

func (s *Server) ListStaff(w http.ResponseWriter, r *http.Request) {
	staff, err := s.service.ListActiveStaff()
	if err != nil {
		s.Respond(w, http.StatusInternalServerError, nil, err)
		return
	}
	s.Respond(w, http.StatusOK, api.StaffListResponse{Staff: staff}, nil)
}

func (s *Server) CreateChemical(w http.ResponseWriter, r *http.Request) {
	var req api.CreateChemicalRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.Respond(w, http.StatusBadRequest, nil, err)
		return
	}
	chem, err := s.service.CreateChemical(req.Name, req.Unit, req.UnitPrice, req.Stock, req.SafetyStock)
	if err != nil {
		s.Respond(w, http.StatusInternalServerError, nil, err)
		return
	}
	s.Respond(w, http.StatusCreated, chem, nil)
}

func (s *Server) ListChemicals(w http.ResponseWriter, r *http.Request) {
	chemicals, err := s.service.ListChemicals()
	if err != nil {
		s.Respond(w, http.StatusInternalServerError, nil, err)
		return
	}
	s.Respond(w, http.StatusOK, api.ChemicalListResponse{Chemicals: chemicals}, nil)
}

func (s *Server) CreateInspection(w http.ResponseWriter, r *http.Request) {
	var req api.CreateInspectionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.Respond(w, http.StatusBadRequest, nil, err)
		return
	}
	points := make([]core.InspectionPoint, 0, len(req.ControlPoints))
	for _, p := range req.ControlPoints {
		points = append(points, core.InspectionPoint{
			ControlPointID:   p.ControlPointID,
			ControlPointCode: p.ControlPointCode,
			FacilityStatus:   p.FacilityStatus,
			PestRecords:      p.PestRecords,
		})
	}
	insp, err := s.service.CreateInspection(req.ContractID, req.InspectorID, req.InspectionDate, points, req.Notes)
	if err != nil {
		s.Respond(w, http.StatusInternalServerError, nil, err)
		return
	}
	s.Respond(w, http.StatusCreated, insp, nil)
}

func (s *Server) ListInspections(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	contractID := vars["id"]
	inspections, err := s.service.ListInspections(contractID)
	if err != nil {
		s.Respond(w, http.StatusInternalServerError, nil, err)
		return
	}
	s.Respond(w, http.StatusOK, api.InspectionListResponse{Inspections: inspections}, nil)
}

func (s *Server) CreateOperation(w http.ResponseWriter, r *http.Request) {
	var req api.CreateOperationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.Respond(w, http.StatusBadRequest, nil, err)
		return
	}
	usages := make([]core.UsageRequest, 0, len(req.Usages))
	for _, u := range req.Usages {
		usages = append(usages, core.UsageRequest{
			ChemicalID: u.ChemicalID,
			Amount:     u.Amount,
		})
	}
	op, err := s.service.CreateOperation(req.ContractID, req.OperatorID, req.OperationDate, req.Scope, usages, req.Notes)
	if err != nil {
		if err == core.ErrInsufficientStock || err == core.ErrStaffOverloaded {
			s.Respond(w, http.StatusConflict, nil, err)
			return
		}
		s.Respond(w, http.StatusInternalServerError, nil, err)
		return
	}
	s.Respond(w, http.StatusCreated, op, nil)
}

func (s *Server) ListOperations(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	contractID := vars["id"]
	ops, err := s.service.ListOperations(contractID)
	if err != nil {
		s.Respond(w, http.StatusInternalServerError, nil, err)
		return
	}
	s.Respond(w, http.StatusOK, api.OperationListResponse{Operations: ops}, nil)
}

func (s *Server) EvaluateEffect(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	inspID := vars["id"]
	eval, err := s.service.EvaluateEffect(inspID)
	if err != nil {
		s.Respond(w, http.StatusInternalServerError, nil, err)
		return
	}
	s.Respond(w, http.StatusOK, eval, nil)
}

func (s *Server) GetPendingServices(w http.ResponseWriter, r *http.Request) {
	services, err := s.service.GetPendingServices()
	if err != nil {
		s.Respond(w, http.StatusInternalServerError, nil, err)
		return
	}
	s.Respond(w, http.StatusOK, api.PendingServiceListResponse{Services: services}, nil)
}
