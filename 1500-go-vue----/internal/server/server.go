package server

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"oacms/internal/api"
	"oacms/internal/core"
)

type Server struct {
	ownerService      *core.OwnerService
	electionService   *core.ElectionService
	proposalService   *core.ProposalService
	announcementService *core.AnnouncementService
}

func NewServer(
	ownerService *core.OwnerService,
	electionService *core.ElectionService,
	proposalService *core.ProposalService,
	announcementService *core.AnnouncementService,
) *Server {
	return &Server{
		ownerService:      ownerService,
		electionService:   electionService,
		proposalService:   proposalService,
		announcementService: announcementService,
	}
}

func (s *Server) Routes() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/api/owners", s.handleOwners)
	mux.HandleFunc("/api/owners/", s.handleOwnerByID)
	mux.HandleFunc("/api/delegates", s.handleDelegates)
	mux.HandleFunc("/api/delegates/accept", s.handleAcceptDelegate)
	mux.HandleFunc("/api/delegates/reject", s.handleRejectDelegate)

	mux.HandleFunc("/api/elections", s.handleElections)
	mux.HandleFunc("/api/elections/", s.handleElectionByID)
	mux.HandleFunc("/api/elections/candidate", s.handleAddCandidate)
	mux.HandleFunc("/api/elections/vote", s.handleCastVote)
	mux.HandleFunc("/api/elections/finalize", s.handleFinalizeElection)

	mux.HandleFunc("/api/proposals", s.handleProposals)
	mux.HandleFunc("/api/proposals/", s.handleProposalByID)
	mux.HandleFunc("/api/proposals/second", s.handleSecondProposal)
	mux.HandleFunc("/api/proposals/review", s.handleReviewProposal)

	mux.HandleFunc("/api/announcements", s.handleAnnouncements)
	mux.HandleFunc("/api/announcements/", s.handleAnnouncementByID)

	return mux
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func (s *Server) handleOwners(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		owners := s.ownerService.ListOwners()
		dto := make([]api.OwnerDTO, len(owners))
		for i, o := range owners {
			dto[i] = toOwnerDTO(o)
		}
		writeJSON(w, http.StatusOK, api.Response{Success: true, Data: dto})
	case http.MethodPost:
		var req api.CreateOwnerRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSON(w, http.StatusBadRequest, api.Response{Success: false, Error: err.Error()})
			return
		}
		owner, err := s.ownerService.CreateOwner(req.Name, req.Phone, req.RoomNumber, req.Area, core.OwnerStatus(req.Status))
		if err != nil {
			writeJSON(w, http.StatusBadRequest, api.Response{Success: false, Error: err.Error()})
			return
		}
		writeJSON(w, http.StatusCreated, api.Response{Success: true, Data: toOwnerDTO(owner)})
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) handleOwnerByID(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/api/owners/")
	if id == "" {
		writeJSON(w, http.StatusNotFound, api.Response{Success: false, Error: "owner id required"})
		return
	}
	owner, err := s.ownerService.GetOwner(id)
	if err != nil {
		writeJSON(w, http.StatusNotFound, api.Response{Success: false, Error: err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, api.Response{Success: true, Data: toOwnerDTO(owner)})
}

func (s *Server) handleDelegates(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		delegates := s.ownerService.GetDelegates()
		dto := make([]api.DelegateRelationDTO, len(delegates))
		for i, d := range delegates {
			dto[i] = toDelegateDTO(d)
		}
		writeJSON(w, http.StatusOK, api.Response{Success: true, Data: dto})
	case http.MethodPost:
		var req api.CreateDelegateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSON(w, http.StatusBadRequest, api.Response{Success: false, Error: err.Error()})
			return
		}
		d, err := s.ownerService.CreateDelegate(req.PrincipalID, req.AgentID)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, api.Response{Success: false, Error: err.Error()})
			return
		}
		writeJSON(w, http.StatusCreated, api.Response{Success: true, Data: toDelegateDTO(d)})
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) handleAcceptDelegate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req api.AcceptDelegateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, api.Response{Success: false, Error: err.Error()})
		return
	}
	d, err := s.ownerService.AcceptDelegate(req.DelegateID, req.AgentID)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, api.Response{Success: false, Error: err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, api.Response{Success: true, Data: toDelegateDTO(d)})
}

func (s *Server) handleRejectDelegate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req api.AcceptDelegateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, api.Response{Success: false, Error: err.Error()})
		return
	}
	d, err := s.ownerService.RejectDelegate(req.DelegateID, req.AgentID)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, api.Response{Success: false, Error: err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, api.Response{Success: true, Data: toDelegateDTO(d)})
}

func (s *Server) handleElections(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		elections := s.electionService.ListElections()
		dto := make([]api.ElectionDTO, len(elections))
		for i, e := range elections {
			dto[i] = toElectionDTO(e)
		}
		writeJSON(w, http.StatusOK, api.Response{Success: true, Data: dto})
	case http.MethodPost:
		var req api.CreateElectionRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSON(w, http.StatusBadRequest, api.Response{Success: false, Error: err.Error()})
			return
		}
		election, err := s.electionService.CreateElection(req.Title, req.Description, req.Rules, req.CommitteeSize)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, api.Response{Success: false, Error: err.Error()})
			return
		}
		writeJSON(w, http.StatusCreated, api.Response{Success: true, Data: toElectionDTO(election)})
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) handleElectionByID(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/api/elections/")
	if id == "" {
		writeJSON(w, http.StatusNotFound, api.Response{Success: false, Error: "election id required"})
		return
	}
	election, err := s.electionService.GetElection(id)
	if err != nil {
		writeJSON(w, http.StatusNotFound, api.Response{Success: false, Error: err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, api.Response{Success: true, Data: toElectionDTO(election)})
}

func (s *Server) handleAddCandidate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req api.AddCandidateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, api.Response{Success: false, Error: err.Error()})
		return
	}
	c, err := s.electionService.AddCandidate(req.ElectionID, req.OwnerID, req.Reason)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, api.Response{Success: false, Error: err.Error()})
		return
	}
	writeJSON(w, http.StatusCreated, api.Response{Success: true, Data: toCandidateDTO(c)})
}

func (s *Server) handleCastVote(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req api.CastVoteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, api.Response{Success: false, Error: err.Error()})
		return
	}
	if err := s.electionService.CastVote(req.ElectionID, req.VoterID, req.CandidateID, req.Abstain); err != nil {
		writeJSON(w, http.StatusBadRequest, api.Response{Success: false, Error: err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, api.Response{Success: true, Message: "vote cast successfully"})
}

func (s *Server) handleFinalizeElection(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	body, _ := io.ReadAll(r.Body)
	var data struct {
		ElectionID string `json:"election_id"`
	}
	json.Unmarshal(body, &data)
	if data.ElectionID == "" {
		writeJSON(w, http.StatusBadRequest, api.Response{Success: false, Error: "election_id required"})
		return
	}
	e, err := s.electionService.FinalizeElection(data.ElectionID)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, api.Response{Success: false, Error: err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, api.Response{Success: true, Data: toElectionDTO(e)})
}

func (s *Server) handleProposals(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		proposals := s.proposalService.ListProposals()
		dto := make([]api.ProposalDTO, len(proposals))
		for i, p := range proposals {
			dto[i] = toProposalDTO(p)
		}
		writeJSON(w, http.StatusOK, api.Response{Success: true, Data: dto})
	case http.MethodPost:
		var req api.CreateProposalRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSON(w, http.StatusBadRequest, api.Response{Success: false, Error: err.Error()})
			return
		}
		proposal, err := s.proposalService.CreateProposal(req.Title, req.Content, req.ProposerID)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, api.Response{Success: false, Error: err.Error()})
			return
		}
		writeJSON(w, http.StatusCreated, api.Response{Success: true, Data: toProposalDTO(proposal)})
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) handleProposalByID(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/api/proposals/")
	if id == "" {
		writeJSON(w, http.StatusNotFound, api.Response{Success: false, Error: "proposal id required"})
		return
	}
	proposal, err := s.proposalService.GetProposal(id)
	if err != nil {
		writeJSON(w, http.StatusNotFound, api.Response{Success: false, Error: err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, api.Response{Success: true, Data: toProposalDTO(proposal)})
}

func (s *Server) handleSecondProposal(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req api.SecondProposalRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, api.Response{Success: false, Error: err.Error()})
		return
	}
	p, err := s.proposalService.SecondProposal(req.ProposalID, req.SeconderID)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, api.Response{Success: false, Error: err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, api.Response{Success: true, Data: toProposalDTO(p)})
}

func (s *Server) handleReviewProposal(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req api.ReviewProposalRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, api.Response{Success: false, Error: err.Error()})
		return
	}
	var deadline *time.Time
	if req.DeadlineDays > 0 {
		d := time.Now().Add(time.Duration(req.DeadlineDays) * 24 * time.Hour)
		deadline = &d
	}
	p, err := s.proposalService.ReviewProposal(req.ProposalID, core.ProposalStatus(req.Decision), req.Opinion, req.ResponsibleDept, deadline)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, api.Response{Success: false, Error: err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, api.Response{Success: true, Data: toProposalDTO(p)})
}

func (s *Server) handleAnnouncements(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		includeArchived := false
		if q := r.URL.Query().Get("include_archived"); q == "true" {
			includeArchived = true
		}
		announcements := s.announcementService.ListAnnouncements(includeArchived)
		dto := make([]api.AnnouncementDTO, len(announcements))
		for i, a := range announcements {
			dto[i] = toAnnouncementDTO(a)
		}
		writeJSON(w, http.StatusOK, api.Response{Success: true, Data: dto})
	case http.MethodPost:
		var req api.CreateAnnouncementRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSON(w, http.StatusBadRequest, api.Response{Success: false, Error: err.Error()})
			return
		}
		announcement, err := s.announcementService.CreateAnnouncement(req.Title, req.Content, req.ValidityDays)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, api.Response{Success: false, Error: err.Error()})
			return
		}
		writeJSON(w, http.StatusCreated, api.Response{Success: true, Data: toAnnouncementDTO(announcement)})
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) handleAnnouncementByID(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		id := strings.TrimPrefix(r.URL.Path, "/api/announcements/")
		if id == "" {
			writeJSON(w, http.StatusNotFound, api.Response{Success: false, Error: "announcement id required"})
			return
		}
		announcement, err := s.announcementService.GetAnnouncement(id)
		if err != nil {
			writeJSON(w, http.StatusNotFound, api.Response{Success: false, Error: err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, api.Response{Success: true, Data: toAnnouncementDTO(announcement)})
	case http.MethodPost:
		parts := strings.Split(r.URL.Path, "/")
		if len(parts) < 5 || parts[4] != "archive" {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		id := parts[3]
		a, err := s.announcementService.ArchiveAnnouncement(id)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, api.Response{Success: false, Error: err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, api.Response{Success: true, Data: toAnnouncementDTO(a)})
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func toOwnerDTO(o *core.Owner) api.OwnerDTO {
	return api.OwnerDTO{
		ID:         o.ID,
		Name:       o.Name,
		Phone:      o.Phone,
		RoomNumber: o.RoomNumber,
		Area:       o.Area,
		Status:     string(o.Status),
		VoteWeight: o.VoteWeight,
	}
}

func toDelegateDTO(d *core.DelegateRelation) api.DelegateRelationDTO {
	return api.DelegateRelationDTO{
		ID:          d.ID,
		PrincipalID: d.PrincipalID,
		AgentID:     d.AgentID,
		Status:      string(d.Status),
		CreatedAt:   d.CreatedAt,
		AcceptedAt:  d.AcceptedAt,
	}
}

func toCandidateDTO(c *core.Candidate) api.CandidateDTO {
	return api.CandidateDTO{
		ID:        c.ID,
		OwnerID:   c.OwnerID,
		Name:      c.Name,
		Reason:    c.Reason,
		VoteCount: c.VoteCount,
	}
}

func toElectionDTO(e *core.Election) api.ElectionDTO {
	candidates := make([]api.CandidateDTO, len(e.Candidates))
	for i, c := range e.Candidates {
		candidates[i] = toCandidateDTO(&c)
	}
	return api.ElectionDTO{
		ID:            e.ID,
		Title:         e.Title,
		Description:   e.Description,
		Rules:         e.Rules,
		CommitteeSize: e.CommitteeSize,
		StartTime:     e.StartTime,
		EndTime:       e.EndTime,
		Status:        string(e.Status),
		Candidates:    candidates,
		Winners:       e.Winners,
	}
}

func toProposalDTO(p *core.Proposal) api.ProposalDTO {
	return api.ProposalDTO{
		ID:              p.ID,
		Title:           p.Title,
		Content:         p.Content,
		ProposerID:      p.ProposerID,
		Seconders:       p.Seconders,
		Status:          string(p.Status),
		CommitteeOpinion: p.CommitteeOpinion,
		ResponsibleDept: p.ResponsibleDept,
		Deadline:        p.Deadline,
		CreatedAt:       p.CreatedAt,
		ReviewedAt:      p.ReviewedAt,
	}
}

func toAnnouncementDTO(a *core.Announcement) api.AnnouncementDTO {
	return api.AnnouncementDTO{
		ID:          a.ID,
		Title:       a.Title,
		Content:     a.Content,
		PublishedAt: a.PublishedAt,
		ExpireAt:    a.ExpireAt,
		Status:      string(a.Status),
	}
}

func GetPort() string {
	return "8080"
}

func parsePort(arg string, defaultPort string) string {
	if arg != "" {
		_, err := strconv.Atoi(arg)
		if err == nil {
			return ":" + arg
		}
	}
	return ":" + defaultPort
}
