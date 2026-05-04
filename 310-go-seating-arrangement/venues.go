package main

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"time"
)

var (
	ErrSectionNameDuplicate = errors.New("section name already exists")
	ErrVenueNotFound        = errors.New("venue not found")
	ErrSectionNotFound      = errors.New("section not found")
	ErrRowNotFound          = errors.New("row not found")
	ErrSeatNotFound         = errors.New("seat not found")
	ErrIsAisle              = errors.New("position is an aisle")
)

type CreateVenueRequest struct {
	Name     string                 `json:"name"`
	Sections []CreateSectionRequest `json:"sections"`
}

type CreateSectionRequest struct {
	Name      string            `json:"name"`
	RowCount  int               `json:"row_count"`
	SeatsPerRow int              `json:"seats_per_row"`
	Aisles    map[int][]int     `json:"aisles,omitempty"`
}

func generateID() string {
	bytes := make([]byte, 8)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

func CreateVenue(req CreateVenueRequest) (*Venue, error) {
	venueMutex.Lock()
	defer venueMutex.Unlock()

	sectionNames := make(map[string]bool)
	for _, sec := range req.Sections {
		if sec.Name == "" {
			return nil, errors.New("section name cannot be empty")
		}
		if sectionNames[sec.Name] {
			return nil, ErrSectionNameDuplicate
		}
		sectionNames[sec.Name] = true
	}

	venue := &Venue{
		ID:        generateID(),
		Name:      req.Name,
		Sections:  make(map[string]*Section),
		CreatedAt: time.Now(),
	}

	totalSeats := 0
	for _, secReq := range req.Sections {
		section, err := createSection(secReq)
		if err != nil {
			return nil, err
		}
		venue.Sections[secReq.Name] = section
		totalSeats += section.SeatCount
	}

	venue.TotalSeats = totalSeats
	venues[venue.ID] = venue

	return venue, nil
}

func createSection(req CreateSectionRequest) (*Section, error) {
	if req.RowCount <= 0 {
		return nil, errors.New("row count must be greater than 0")
	}
	if req.SeatsPerRow <= 0 {
		return nil, errors.New("seats per row must be greater than 0")
	}

	section := &Section{
		Name:     req.Name,
		Rows:     make([]*Row, req.RowCount),
	}

	seatCount := 0
	for i := 0; i < req.RowCount; i++ {
		rowNumber := i + 1
		row := &Row{
			RowNumber: rowNumber,
			Seats:     make([]*SeatTemplate, req.SeatsPerRow),
		}

		aisleSeats := make(map[int]bool)
		if req.Aisles != nil {
			for _, idx := range req.Aisles[rowNumber] {
				aisleSeats[idx] = true
			}
		}

		seatNum := 1
		for j := 0; j < req.SeatsPerRow; j++ {
			seatIndex := j + 1
			isAisle := aisleSeats[seatIndex]

			seat := &SeatTemplate{
				Index:   seatIndex,
				IsAisle: isAisle,
			}

			if !isAisle {
				seat.SeatNumber = fmt.Sprintf("%d", seatNum)
				seatNum++
				seatCount++
			}

			row.Seats[j] = seat
		}

		section.Rows[i] = row
	}

	section.SeatCount = seatCount
	return section, nil
}

func GetVenue(venueID string) (*Venue, error) {
	venueMutex.RLock()
	defer venueMutex.RUnlock()

	venue, exists := venues[venueID]
	if !exists {
		return nil, ErrVenueNotFound
	}
	return venue, nil
}

func GetAllVenues() []*Venue {
	venueMutex.RLock()
	defer venueMutex.RUnlock()

	result := make([]*Venue, 0, len(venues))
	for _, v := range venues {
		result = append(result, v)
	}
	return result
}

func GetAvailableSeatsForView(venue *Venue, sectionName string) ([]*SeatTemplate, error) {
	venueMutex.RLock()
	defer venueMutex.RUnlock()

	section, exists := venue.Sections[sectionName]
	if !exists {
		return nil, ErrSectionNotFound
	}

	var availableSeats []*SeatTemplate
	for _, row := range section.Rows {
		for _, seat := range row.Seats {
			if !seat.IsAisle {
				availableSeats = append(availableSeats, seat)
			}
		}
	}

	return availableSeats, nil
}

func ValidateSeatLocation(venue *Venue, sectionName string, rowNumber int, seatIndex int) (*SeatTemplate, error) {
	section, exists := venue.Sections[sectionName]
	if !exists {
		return nil, ErrSectionNotFound
	}

	if rowNumber < 1 || rowNumber > len(section.Rows) {
		return nil, ErrRowNotFound
	}

	row := section.Rows[rowNumber-1]

	if seatIndex < 1 || seatIndex > len(row.Seats) {
		return nil, ErrSeatNotFound
	}

	seat := row.Seats[seatIndex-1]
	if seat.IsAisle {
		return nil, ErrIsAisle
	}

	return seat, nil
}
