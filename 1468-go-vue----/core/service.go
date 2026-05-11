package core

import (
	"time"

	"tender-management/common"
)

type TenderService struct {
	store          *Store
	projectService *ProjectService
	bidService     *BidService
	openingService *OpeningService
}

func NewTenderService() *TenderService {
	store := NewStore()
	projectService := NewProjectService(store)
	bidService := NewBidService(store, projectService)
	openingService := NewOpeningService(store, projectService, bidService)
	return &TenderService{
		store:          store,
		projectService: projectService,
		bidService:     bidService,
		openingService: openingService,
	}
}

func (s *TenderService) CreateProject(req common.CreateProjectRequest) (*common.Project, error) {
	return s.projectService.CreateProject(req)
}

func (s *TenderService) UpdateProject(projectID string, req common.UpdateProjectRequest) error {
	return s.projectService.UpdateProject(projectID, req)
}

func (s *TenderService) PublishProject(projectID string) error {
	return s.projectService.PublishProject(projectID)
}

func (s *TenderService) GetProject(projectID string) (*common.Project, error) {
	return s.projectService.GetProject(projectID)
}

func (s *TenderService) ListProjectsForSupplier(supplierID string) []common.Project {
	return s.projectService.ListProjectsForSupplier(supplierID)
}

func (s *TenderService) ListAllProjects() []common.Project {
	return s.projectService.ListAllProjects()
}

func (s *TenderService) SubmitBid(req common.SubmitBidRequest) (*common.Bid, error) {
	return s.bidService.SubmitBid(req, time.Now())
}

func (s *TenderService) UpdateBid(bidID string, req common.UpdateBidRequest) (*common.Bid, error) {
	return s.bidService.UpdateBid(bidID, req, time.Now())
}

func (s *TenderService) GetBid(bidID string) (*common.Bid, error) {
	return s.bidService.GetBid(bidID)
}

func (s *TenderService) GetBidBySupplier(projectID, supplierID string) (*common.Bid, error) {
	return s.bidService.GetBidBySupplier(projectID, supplierID)
}

func (s *TenderService) ListBidsByProject(projectID string) []common.Bid {
	return s.bidService.ListBidsByProject(projectID)
}

func (s *TenderService) OpenProject(projectID string) (*common.OpeningResult, error) {
	return s.openingService.OpenProject(projectID, time.Now())
}

func (s *TenderService) GetOpeningResult(projectID string) (*common.OpeningResult, error) {
	return s.openingService.GetOpeningResult(projectID)
}
