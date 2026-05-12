package repository

import (
	"context"
	"sync"
	"time"

	"organdonation/internal/models"
)

type Repository struct {
	mu               sync.RWMutex
	donors           map[string]*models.Donor
	recipients       map[string]*models.Recipient
	organs           map[string]*models.Organ
	transplants      map[string]*models.TransplantRecord
	followUps        map[string]*models.FollowUp
	todos            map[string]*models.Todo
	locations        map[string]*models.Location
	personnel        map[string]*models.Personnel
	reports          map[string]*models.StatisticsReport
	notifications    map[string]*models.MatchNotification
	donorNoIndex     map[string]string
	recipientNoIndex map[string]string
}

func New() *Repository {
	return &Repository{
		donors:           make(map[string]*models.Donor),
		recipients:       make(map[string]*models.Recipient),
		organs:           make(map[string]*models.Organ),
		transplants:      make(map[string]*models.TransplantRecord),
		followUps:        make(map[string]*models.FollowUp),
		todos:            make(map[string]*models.Todo),
		locations:        make(map[string]*models.Location),
		personnel:        make(map[string]*models.Personnel),
		reports:          make(map[string]*models.StatisticsReport),
		notifications:    make(map[string]*models.MatchNotification),
		donorNoIndex:     make(map[string]string),
		recipientNoIndex: make(map[string]string),
	}
}

func (r *Repository) GetDonorByNo(ctx context.Context, no string) (*models.Donor, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	id, exists := r.donorNoIndex[no]
	if !exists {
		return nil, false
	}
	donor, ok := r.donors[id]
	return donor, ok
}

func (r *Repository) CreateDonor(ctx context.Context, donor *models.Donor) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.donorNoIndex[donor.DonorNo]; exists {
		return nil
	}
	donor.CreatedAt = time.Now()
	donor.UpdatedAt = time.Now()
	r.donors[donor.ID] = donor
	r.donorNoIndex[donor.DonorNo] = donor.ID
	return nil
}

func (r *Repository) GetDonor(ctx context.Context, id string) (*models.Donor, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	donor, ok := r.donors[id]
	return donor, ok
}

func (r *Repository) ListDonors(ctx context.Context, page, size int) ([]*models.Donor, int) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	all := make([]*models.Donor, 0, len(r.donors))
	for _, d := range r.donors {
		all = append(all, d)
	}
	total := len(all)
	start := (page - 1) * size
	if start >= total {
		return []*models.Donor{}, total
	}
	end := start + size
	if end > total {
		end = total
	}
	return all[start:end], total
}

func (r *Repository) UpdateDonor(ctx context.Context, donor *models.Donor) {
	r.mu.Lock()
	defer r.mu.Unlock()
	donor.UpdatedAt = time.Now()
	r.donors[donor.ID] = donor
}

func (r *Repository) CreateOrgan(ctx context.Context, organ *models.Organ) {
	r.mu.Lock()
	defer r.mu.Unlock()
	organ.CreatedAt = time.Now()
	organ.UpdatedAt = time.Now()
	r.organs[organ.ID] = organ
}

func (r *Repository) GetOrgan(ctx context.Context, id string) (*models.Organ, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	organ, ok := r.organs[id]
	return organ, ok
}

func (r *Repository) UpdateOrgan(ctx context.Context, organ *models.Organ) {
	r.mu.Lock()
	defer r.mu.Unlock()
	organ.UpdatedAt = time.Now()
	r.organs[organ.ID] = organ
}

func (r *Repository) ListOrgansByDonor(ctx context.Context, donorID string) []*models.Organ {
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := make([]*models.Organ, 0)
	for _, o := range r.organs {
		if o.DonorID == donorID {
			result = append(result, o)
		}
	}
	return result
}

func (r *Repository) ListPendingOrgans(ctx context.Context) []*models.Organ {
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := make([]*models.Organ, 0)
	for _, o := range r.organs {
		if o.Status == models.OrganStatusPending && o.Assessment != nil && o.Assessment.FunctionScore >= 60 {
			result = append(result, o)
		}
	}
	return result
}

func (r *Repository) GetRecipientByNo(ctx context.Context, no string) (*models.Recipient, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	id, exists := r.recipientNoIndex[no]
	if !exists {
		return nil, false
	}
	recipient, ok := r.recipients[id]
	return recipient, ok
}

func (r *Repository) CreateRecipient(ctx context.Context, recipient *models.Recipient) {
	r.mu.Lock()
	defer r.mu.Unlock()
	recipient.CreatedAt = time.Now()
	recipient.UpdatedAt = time.Now()
	r.recipients[recipient.ID] = recipient
	r.recipientNoIndex[recipient.RecipientNo] = recipient.ID
}

func (r *Repository) GetRecipient(ctx context.Context, id string) (*models.Recipient, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	recipient, ok := r.recipients[id]
	return recipient, ok
}

func (r *Repository) ListRecipients(ctx context.Context, page, size int) ([]*models.Recipient, int) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	all := make([]*models.Recipient, 0, len(r.recipients))
	for _, r := range r.recipients {
		all = append(all, r)
	}
	total := len(all)
	start := (page - 1) * size
	if start >= total {
		return []*models.Recipient{}, total
	}
	end := start + size
	if end > total {
		end = total
	}
	return all[start:end], total
}

func (r *Repository) ListWaitingRecipients(ctx context.Context) []*models.Recipient {
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := make([]*models.Recipient, 0)
	for _, r := range r.recipients {
		if r.MatchedOrganID == nil {
			result = append(result, r)
		}
	}
	return result
}

func (r *Repository) UpdateRecipient(ctx context.Context, recipient *models.Recipient) {
	r.mu.Lock()
	defer r.mu.Unlock()
	recipient.UpdatedAt = time.Now()
	r.recipients[recipient.ID] = recipient
}

func (r *Repository) CreateTransplant(ctx context.Context, transplant *models.TransplantRecord) {
	r.mu.Lock()
	defer r.mu.Unlock()
	transplant.CreatedAt = time.Now()
	transplant.UpdatedAt = time.Now()
	r.transplants[transplant.ID] = transplant
}

func (r *Repository) GetTransplant(ctx context.Context, id string) (*models.TransplantRecord, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	transplant, ok := r.transplants[id]
	return transplant, ok
}

func (r *Repository) ListTransplants(ctx context.Context, page, size int) ([]*models.TransplantRecord, int) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	all := make([]*models.TransplantRecord, 0, len(r.transplants))
	for _, t := range r.transplants {
		all = append(all, t)
	}
	total := len(all)
	start := (page - 1) * size
	if start >= total {
		return []*models.TransplantRecord{}, total
	}
	end := start + size
	if end > total {
		end = total
	}
	return all[start:end], total
}

func (r *Repository) UpdateTransplant(ctx context.Context, transplant *models.TransplantRecord) {
	r.mu.Lock()
	defer r.mu.Unlock()
	transplant.UpdatedAt = time.Now()
	r.transplants[transplant.ID] = transplant
}

func (r *Repository) CreateFollowUp(ctx context.Context, followUp *models.FollowUp) {
	r.mu.Lock()
	defer r.mu.Unlock()
	followUp.CreatedAt = time.Now()
	r.followUps[followUp.ID] = followUp
}

func (r *Repository) ListFollowUpsByTransplant(ctx context.Context, transplantID string) []*models.FollowUp {
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := make([]*models.FollowUp, 0)
	for _, f := range r.followUps {
		if f.TransplantID == transplantID {
			result = append(result, f)
		}
	}
	return result
}

func (r *Repository) CreateTodo(ctx context.Context, todo *models.Todo) {
	r.mu.Lock()
	defer r.mu.Unlock()
	todo.CreatedAt = time.Now()
	todo.UpdatedAt = time.Now()
	r.todos[todo.ID] = todo
}

func (r *Repository) GetTodo(ctx context.Context, id string) (*models.Todo, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	todo, ok := r.todos[id]
	return todo, ok
}

func (r *Repository) ListTodos(ctx context.Context, page, size int) ([]*models.Todo, int) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	all := make([]*models.Todo, 0, len(r.todos))
	for _, t := range r.todos {
		all = append(all, t)
	}
	total := len(all)
	start := (page - 1) * size
	if start >= total {
		return []*models.Todo{}, total
	}
	end := start + size
	if end > total {
		end = total
	}
	return all[start:end], total
}

func (r *Repository) UpdateTodo(ctx context.Context, todo *models.Todo) {
	r.mu.Lock()
	defer r.mu.Unlock()
	todo.UpdatedAt = time.Now()
	r.todos[todo.ID] = todo
}

func (r *Repository) CreateLocation(ctx context.Context, location *models.Location) {
	r.mu.Lock()
	defer r.mu.Unlock()
	location.CreatedAt = time.Now()
	location.UpdatedAt = time.Now()
	r.locations[location.ID] = location
}

func (r *Repository) GetLocation(ctx context.Context, id string) (*models.Location, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	location, ok := r.locations[id]
	return location, ok
}

func (r *Repository) ListLocations(ctx context.Context) []*models.Location {
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := make([]*models.Location, 0, len(r.locations))
	for _, l := range r.locations {
		result = append(result, l)
	}
	return result
}

func (r *Repository) CreatePersonnel(ctx context.Context, personnel *models.Personnel) {
	r.mu.Lock()
	defer r.mu.Unlock()
	personnel.CreatedAt = time.Now()
	personnel.UpdatedAt = time.Now()
	r.personnel[personnel.ID] = personnel
}

func (r *Repository) ListPersonnel(ctx context.Context) []*models.Personnel {
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := make([]*models.Personnel, 0, len(r.personnel))
	for _, p := range r.personnel {
		result = append(result, p)
	}
	return result
}

func (r *Repository) CreateReport(ctx context.Context, report *models.StatisticsReport) {
	r.mu.Lock()
	defer r.mu.Unlock()
	report.CreatedAt = time.Now()
	r.reports[report.ID] = report
}

func (r *Repository) ListReports(ctx context.Context) []*models.StatisticsReport {
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := make([]*models.StatisticsReport, 0, len(r.reports))
	for _, r := range r.reports {
		result = append(result, r)
	}
	return result
}

func (r *Repository) CreateNotification(ctx context.Context, notification *models.MatchNotification) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.notifications[notification.ID] = notification
}

func (r *Repository) GetPendingNotifications(ctx context.Context) []*models.MatchNotification {
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := make([]*models.MatchNotification, 0)
	now := time.Now()
	for _, n := range r.notifications {
		if !n.Confirmed && now.Before(n.ExpiresAt) {
			result = append(result, n)
		}
	}
	return result
}

func (r *Repository) UpdateNotification(ctx context.Context, notification *models.MatchNotification) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.notifications[notification.ID] = notification
}

func (r *Repository) CountNewDonorsSince(ctx context.Context, since time.Time) int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	count := 0
	for _, d := range r.donors {
		if d.CreatedAt.After(since) {
			count++
		}
	}
	return count
}

func (r *Repository) CountSuccessfulMatchesSince(ctx context.Context, since time.Time) int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	count := 0
	for _, t := range r.transplants {
		if t.CreatedAt.After(since) {
			count++
		}
	}
	return count
}

func (r *Repository) CountOrgansByType(ctx context.Context) map[models.OrganType]int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := make(map[models.OrganType]int)
	for _, o := range r.organs {
		result[o.OrganType]++
	}
	return result
}

func (r *Repository) CountRecipientsByOrganNeeded(ctx context.Context) map[models.OrganType]int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := make(map[models.OrganType]int)
	for _, r := range r.recipients {
		if r.MatchedOrganID == nil {
			result[r.OrganNeeded]++
		}
	}
	return result
}

func (r *Repository) GetAllRecipients(ctx context.Context) []*models.Recipient {
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := make([]*models.Recipient, 0, len(r.recipients))
	for _, r := range r.recipients {
		result = append(result, r)
	}
	return result
}
