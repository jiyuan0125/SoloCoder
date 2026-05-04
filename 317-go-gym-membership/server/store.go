package server

import (
	"crypto/md5"
	"crypto/rand"
	"encoding/hex"
	"sync"
	"sync/atomic"
	"time"

	"gym-membership/types"
)

var idCounter uint64

func generateID() string {
	counter := atomic.AddUint64(&idCounter, 1)
	timestamp := time.Now().UnixNano()
	
	randomBytes := make([]byte, 8)
	rand.Read(randomBytes)
	
	id := make([]byte, 16)
	for i := 0; i < 8; i++ {
		id[i] = byte(timestamp >> (i * 8))
		id[i+8] = byte(counter >> (i * 8)) ^ randomBytes[i]
	}
	
	return hex.EncodeToString(id)
}

type Store struct {
	mu              sync.RWMutex
	members         map[string]*types.Member
	membersByPhone  map[string]string
	cards           map[string]*types.MembershipCard
	cardsByMember   map[string]*types.MembershipCard
	classes         map[string]*types.Class
	classInstances  map[string]*types.ClassInstance
	classInstancesByClass map[string][]*types.ClassInstance
	bookings        map[string]*types.Booking
	bookingsByMember map[string][]*types.Booking
	bookingsByInstance map[string][]*types.Booking
	attendances     map[string]*types.Attendance
	attendancesByMember map[string][]*types.Attendance
	attendancesByInstance map[string][]*types.Attendance
	cardPrices      map[types.CardType]*types.CardPrice
}

func NewStore() *Store {
	return &Store{
		members:              make(map[string]*types.Member),
		membersByPhone:       make(map[string]string),
		cards:                make(map[string]*types.MembershipCard),
		cardsByMember:        make(map[string]*types.MembershipCard),
		classes:              make(map[string]*types.Class),
		classInstances:       make(map[string]*types.ClassInstance),
		classInstancesByClass: make(map[string][]*types.ClassInstance),
		bookings:             make(map[string]*types.Booking),
		bookingsByMember:     make(map[string][]*types.Booking),
		bookingsByInstance:   make(map[string][]*types.Booking),
		attendances:          make(map[string]*types.Attendance),
		attendancesByMember:  make(map[string][]*types.Attendance),
		attendancesByInstance: make(map[string][]*types.Attendance),
		cardPrices:           make(map[types.CardType]*types.CardPrice),
	}
}

func hashPassword(password string) string {
	return hex.EncodeToString(md5.New().Sum([]byte(password)))
}

func (s *Store) CreateMember(name, phone, password string) (*types.Member, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.membersByPhone[phone]; exists {
		return nil, ErrMemberExists
	}

	member := &types.Member{
		ID:        generateID(),
		Name:      name,
		Phone:     phone,
		Password:  hashPassword(password),
		CreatedAt: time.Now(),
	}

	s.members[member.ID] = member
	s.membersByPhone[phone] = member.ID
	return member, nil
}

func (s *Store) GetMemberByID(id string) (*types.Member, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	member, exists := s.members[id]
	if !exists {
		return nil, ErrMemberNotFound
	}
	return member, nil
}

func (s *Store) GetMemberByPhone(phone string) (*types.Member, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	id, exists := s.membersByPhone[phone]
	if !exists {
		return nil, ErrMemberNotFound
	}

	return s.members[id], nil
}

func (s *Store) PurchaseCard(memberID string, cardType types.CardType) (*types.MembershipCard, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	member, exists := s.members[memberID]
	if !exists {
		return nil, ErrMemberNotFound
	}

	if currentCard, exists := s.cardsByMember[memberID]; exists {
		if currentCard.Status == types.CardActive {
			return nil, ErrActiveCardExists
		}
	}

	price, exists := s.cardPrices[cardType]
	if !exists {
		return nil, ErrInvalidCardType
	}

	startDate := time.Now()
	endDate := startDate.AddDate(0, 0, price.ValidDays)

	card := &types.MembershipCard{
		ID:        generateID(),
		MemberID:  memberID,
		Type:      cardType,
		StartDate: startDate,
		EndDate:   endDate,
		Price:     price.Price,
		Status:    types.CardActive,
		CreatedAt: time.Now(),
	}

	s.cards[card.ID] = card
	s.cardsByMember[memberID] = card
	member.CurrentCard = card

	return card, nil
}

func (s *Store) RenewCard(memberID string, cardType types.CardType) (*types.MembershipCard, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	price, exists := s.cardPrices[cardType]
	if !exists {
		return nil, ErrInvalidCardType
	}

	oldCard, hasOldCard := s.cardsByMember[memberID]
	
	var startDate time.Time
	if hasOldCard && oldCard.Status == types.CardActive {
		startDate = oldCard.EndDate.AddDate(0, 0, 1)
	} else {
		startDate = time.Now()
	}

	endDate := startDate.AddDate(0, 0, price.ValidDays)

	newCard := &types.MembershipCard{
		ID:        generateID(),
		MemberID:  memberID,
		Type:      cardType,
		StartDate: startDate,
		EndDate:   endDate,
		Price:     price.Price,
		Status:    types.CardActive,
		CreatedAt: time.Now(),
	}

	if hasOldCard {
		oldCard.Status = types.CardExpired
	}

	s.cards[newCard.ID] = newCard
	s.cardsByMember[memberID] = newCard
	
	if member, exists := s.members[memberID]; exists {
		member.CurrentCard = newCard
	}

	return newCard, nil
}

func (s *Store) GetMemberCards(memberID string) (*types.MembershipCard, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	card, exists := s.cardsByMember[memberID]
	if !exists {
		return nil, ErrCardNotFound
	}
	return card, nil
}

func (s *Store) CreateClass(req *types.CreateClassRequest) (*types.Class, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if req.Weekday < 0 || req.Weekday > 6 {
		return nil, ErrInvalidClassWeekday
	}

	if !validateStartTime(req.StartTime) {
		return nil, ErrInvalidStartTime
	}

	class := &types.Class{
		ID:          generateID(),
		Name:        req.Name,
		Weekday:     req.Weekday,
		StartTime:   req.StartTime,
		Duration:    req.Duration,
		MaxCapacity: req.MaxCapacity,
		MinCapacity: req.MinCapacity,
		Instructor:  req.Instructor,
		Location:    req.Location,
		CreatedAt:   time.Now(),
	}

	s.classes[class.ID] = class
	return class, nil
}

func (s *Store) GetClassByID(id string) (*types.Class, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	class, exists := s.classes[id]
	if !exists {
		return nil, ErrClassNotFound
	}
	return class, nil
}

func (s *Store) ListClasses(weekday *time.Weekday) []types.Class {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var classes []types.Class
	for _, c := range s.classes {
		if weekday != nil && c.Weekday != *weekday {
			continue
		}
		classes = append(classes, *c)
	}
	return classes
}

func (s *Store) GenerateClassInstances(startDate, endDate time.Time) ([]types.ClassInstance, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	var instances []types.ClassInstance

	for _, class := range s.classes {
		current := startDate
		for current.Before(endDate) || current.Equal(endDate) {
			if current.Weekday() == class.Weekday {
				instance := &types.ClassInstance{
					ID:        generateID(),
					ClassID:   class.ID,
					Date:      current,
					Status:    types.ClassScheduled,
					CreatedAt: time.Now(),
				}

				s.classInstances[instance.ID] = instance
				s.classInstancesByClass[class.ID] = append(s.classInstancesByClass[class.ID], instance)
				instances = append(instances, *instance)
			}
			current = current.AddDate(0, 0, 1)
		}
	}

	return instances, nil
}

func (s *Store) GetClassInstance(classID string, date time.Time) (*types.ClassInstance, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	instances, exists := s.classInstancesByClass[classID]
	if !exists {
		return nil, ErrClassInstanceNotFound
	}

	targetDate := date.Truncate(24 * time.Hour)
	for _, inst := range instances {
		if inst.Date.Truncate(24 * time.Hour).Equal(targetDate) {
			return inst, nil
		}
	}

	return nil, ErrClassInstanceNotFound
}

func (s *Store) BookClass(memberID, classInstanceID string) (*types.Booking, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.members[memberID]; !exists {
		return nil, ErrMemberNotFound
	}

	card, hasCard := s.cardsByMember[memberID]
	if !hasCard || card.Status != types.CardActive || time.Now().After(card.EndDate) {
		return nil, ErrNoActiveCard
	}

	instance, exists := s.classInstances[classInstanceID]
	if !exists {
		return nil, ErrClassInstanceNotFound
	}

	if instance.Status != types.ClassScheduled {
		return nil, ErrClassNotAvailable
	}

	for _, booking := range s.bookingsByMember[memberID] {
		if booking.ClassInstanceID == classInstanceID && booking.Status == types.BookingConfirmed {
			return nil, ErrAlreadyBooked
		}
	}

	bookedCount := 0
	for _, booking := range s.bookingsByInstance[classInstanceID] {
		if booking.Status == types.BookingConfirmed {
			bookedCount++
		}
	}

	class := s.classes[instance.ClassID]
	if bookedCount >= class.MaxCapacity {
		return nil, ErrClassFull
	}

	booking := &types.Booking{
		ID:              generateID(),
		MemberID:        memberID,
		ClassInstanceID: classInstanceID,
		Status:          types.BookingConfirmed,
		BookedAt:        time.Now(),
	}

	s.bookings[booking.ID] = booking
	s.bookingsByMember[memberID] = append(s.bookingsByMember[memberID], booking)
	s.bookingsByInstance[classInstanceID] = append(s.bookingsByInstance[classInstanceID], booking)

	return booking, nil
}

func (s *Store) CancelBooking(bookingID, memberID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	booking, exists := s.bookings[bookingID]
	if !exists {
		return ErrBookingNotFound
	}

	if booking.MemberID != memberID {
		return ErrBookingNotFound
	}

	if booking.Status != types.BookingConfirmed {
		return ErrInvalidBookingStatus
	}

	instance := s.classInstances[booking.ClassInstanceID]
	class := s.classes[instance.ClassID]

	classTime, _ := time.Parse("15:04", class.StartTime)
	classDateTime := time.Date(instance.Date.Year(), instance.Date.Month(), instance.Date.Day(),
		classTime.Hour(), classTime.Minute(), 0, 0, time.Local)

	now := time.Now()
	if now.After(classDateTime.Add(-2 * time.Hour)) {
		return ErrCancelTooLate
	}

	now2 := time.Now()
	booking.Status = types.BookingCancelled
	booking.CancelledAt = &now2

	return nil
}

func (s *Store) CheckIn(memberID, classInstanceID string) (*types.Attendance, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.members[memberID]; !exists {
		return nil, ErrMemberNotFound
	}

	card, hasCard := s.cardsByMember[memberID]
	if !hasCard || card.Status != types.CardActive || time.Now().After(card.EndDate) {
		return nil, ErrNoActiveCard
	}

	instance, exists := s.classInstances[classInstanceID]
	if !exists {
		return nil, ErrClassInstanceNotFound
	}

	if instance.Status != types.ClassScheduled {
		return nil, ErrClassNotAvailable
	}

	class := s.classes[instance.ClassID]
	classTime, _ := time.Parse("15:04", class.StartTime)
	classDateTime := time.Date(instance.Date.Year(), instance.Date.Month(), instance.Date.Day(),
		classTime.Hour(), classTime.Minute(), 0, 0, time.Local)

	now := time.Now()
	if now.Before(classDateTime.Add(-15*time.Minute)) || now.After(classDateTime.Add(15*time.Minute)) {
		return nil, ErrCheckInTimeInvalid
	}

	var booking *types.Booking
	for _, b := range s.bookingsByMember[memberID] {
		if b.ClassInstanceID == classInstanceID && b.Status == types.BookingConfirmed {
			booking = b
			break
		}
	}

	if booking == nil {
		return nil, ErrNoBookingForClass
	}

	for _, a := range s.attendancesByMember[memberID] {
		if a.ClassInstanceID == classInstanceID {
			return nil, ErrAlreadyCheckedIn
		}
	}

	attendance := &types.Attendance{
		ID:              generateID(),
		MemberID:        memberID,
		ClassInstanceID: classInstanceID,
		CheckInTime:     now,
		CreatedAt:       now,
	}

	s.attendances[attendance.ID] = attendance
	s.attendancesByMember[memberID] = append(s.attendancesByMember[memberID], attendance)
	s.attendancesByInstance[classInstanceID] = append(s.attendancesByInstance[classInstanceID], attendance)

	return attendance, nil
}

func (s *Store) GetMemberBookings(memberID string, activeOnly bool) []types.Booking {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var bookings []types.Booking
	for _, b := range s.bookingsByMember[memberID] {
		if activeOnly && b.Status != types.BookingConfirmed {
			continue
		}
		bookings = append(bookings, *b)
	}
	return bookings
}

func (s *Store) GetMemberAttendances(memberID string) []types.Attendance {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var attendances []types.Attendance
	for _, a := range s.attendancesByMember[memberID] {
		attendances = append(attendances, *a)
	}
	return attendances
}

func (s *Store) GetInstanceBookings(instanceID string) []types.Booking {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var bookings []types.Booking
	for _, b := range s.bookingsByInstance[instanceID] {
		bookings = append(bookings, *b)
	}
	return bookings
}

func (s *Store) GetInstanceAttendances(instanceID string) []types.Attendance {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var attendances []types.Attendance
	for _, a := range s.attendancesByInstance[instanceID] {
		attendances = append(attendances, *a)
	}
	return attendances
}

func (s *Store) SetCardPrice(cardType types.CardType, price float64, validDays int) (*types.CardPrice, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	cardPrice := &types.CardPrice{
		ID:        generateID(),
		CardType:  cardType,
		Price:     price,
		ValidDays: validDays,
		UpdatedAt: time.Now(),
	}

	s.cardPrices[cardType] = cardPrice
	return cardPrice, nil
}

func (s *Store) GetCardPrices() []types.CardPrice {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var prices []types.CardPrice
	for _, p := range s.cardPrices {
		prices = append(prices, *p)
	}
	return prices
}

func (s *Store) GetCardPrice(cardType types.CardType) (*types.CardPrice, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	price, exists := s.cardPrices[cardType]
	if !exists {
		return nil, ErrInvalidCardType
	}
	return price, nil
}

func (s *Store) CheckAndCancelExpiredBookings() {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	for memberID, card := range s.cardsByMember {
		if card.Status == types.CardActive && now.After(card.EndDate) {
			card.Status = types.CardExpired
			
			for _, booking := range s.bookingsByMember[memberID] {
				if booking.Status == types.BookingConfirmed {
					instance := s.classInstances[booking.ClassInstanceID]
					if instance.Date.After(now) {
						booking.Status = types.BookingCancelled
						now2 := now
						booking.CancelledAt = &now2
					}
				}
			}
		}
	}
}

func (s *Store) CheckAndCancelLowCapacityClasses() {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	for _, instance := range s.classInstances {
		if instance.Status != types.ClassScheduled {
			continue
		}

		class := s.classes[instance.ClassID]
		classTime, _ := time.Parse("15:04", class.StartTime)
		classDateTime := time.Date(instance.Date.Year(), instance.Date.Month(), instance.Date.Day(),
			classTime.Hour(), classTime.Minute(), 0, 0, time.Local)

		if now.Before(classDateTime.Add(-1*time.Hour)) || now.After(classDateTime) {
			continue
		}

		bookedCount := 0
		for _, booking := range s.bookingsByInstance[instance.ID] {
			if booking.Status == types.BookingConfirmed {
				bookedCount++
			}
		}

		if bookedCount < class.MinCapacity {
			instance.Status = types.ClassCancelled
			
			for _, booking := range s.bookingsByInstance[instance.ID] {
				if booking.Status == types.BookingConfirmed {
					booking.Status = types.BookingCancelled
					now2 := now
					booking.CancelledAt = &now2
				}
			}
		}
	}
}

func (s *Store) GetAllMembers() []*types.Member {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var members []*types.Member
	for _, m := range s.members {
		members = append(members, m)
	}
	return members
}

func (s *Store) GetAllCards() []*types.MembershipCard {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var cards []*types.MembershipCard
	for _, c := range s.cards {
		cards = append(cards, c)
	}
	return cards
}

func (s *Store) GetAllClasses() []*types.Class {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var classes []*types.Class
	for _, c := range s.classes {
		classes = append(classes, c)
	}
	return classes
}

func (s *Store) GetAllClassInstances() []*types.ClassInstance {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var instances []*types.ClassInstance
	for _, i := range s.classInstances {
		instances = append(instances, i)
	}
	return instances
}

func (s *Store) GetAllBookings() []*types.Booking {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var bookings []*types.Booking
	for _, b := range s.bookings {
		bookings = append(bookings, b)
	}
	return bookings
}

func (s *Store) GetAllAttendances() []*types.Attendance {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var attendances []*types.Attendance
	for _, a := range s.attendances {
		attendances = append(attendances, a)
	}
	return attendances
}

func (s *Store) LoadData(data *StoreData) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.members = make(map[string]*types.Member)
	s.membersByPhone = make(map[string]string)
	for _, m := range data.Members {
		s.members[m.ID] = m
		s.membersByPhone[m.Phone] = m.ID
	}

	s.cards = make(map[string]*types.MembershipCard)
	s.cardsByMember = make(map[string]*types.MembershipCard)
	for _, c := range data.Cards {
		s.cards[c.ID] = c
		s.cardsByMember[c.MemberID] = c
	}

	s.classes = make(map[string]*types.Class)
	for _, c := range data.Classes {
		s.classes[c.ID] = c
	}

	s.classInstances = make(map[string]*types.ClassInstance)
	s.classInstancesByClass = make(map[string][]*types.ClassInstance)
	for _, i := range data.ClassInstances {
		s.classInstances[i.ID] = i
		s.classInstancesByClass[i.ClassID] = append(s.classInstancesByClass[i.ClassID], i)
	}

	s.bookings = make(map[string]*types.Booking)
	s.bookingsByMember = make(map[string][]*types.Booking)
	s.bookingsByInstance = make(map[string][]*types.Booking)
	for _, b := range data.Bookings {
		s.bookings[b.ID] = b
		s.bookingsByMember[b.MemberID] = append(s.bookingsByMember[b.MemberID], b)
		s.bookingsByInstance[b.ClassInstanceID] = append(s.bookingsByInstance[b.ClassInstanceID], b)
	}

	s.attendances = make(map[string]*types.Attendance)
	s.attendancesByMember = make(map[string][]*types.Attendance)
	s.attendancesByInstance = make(map[string][]*types.Attendance)
	for _, a := range data.Attendances {
		s.attendances[a.ID] = a
		s.attendancesByMember[a.MemberID] = append(s.attendancesByMember[a.MemberID], a)
		s.attendancesByInstance[a.ClassInstanceID] = append(s.attendancesByInstance[a.ClassInstanceID], a)
	}

	s.cardPrices = make(map[types.CardType]*types.CardPrice)
	for i := range data.CardPrices {
		s.cardPrices[data.CardPrices[i].CardType] = &data.CardPrices[i]
	}
}

func (s *Store) GetData() *StoreData {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return &StoreData{
		Members:        s.GetAllMembers(),
		Cards:          s.GetAllCards(),
		Classes:        s.GetAllClasses(),
		ClassInstances: s.GetAllClassInstances(),
		Bookings:       s.GetAllBookings(),
		Attendances:    s.GetAllAttendances(),
		CardPrices:     s.GetCardPrices(),
	}
}

type StoreData struct {
	Members        []*types.Member           `json:"members"`
	Cards          []*types.MembershipCard   `json:"cards"`
	Classes        []*types.Class            `json:"classes"`
	ClassInstances []*types.ClassInstance    `json:"class_instances"`
	Bookings       []*types.Booking          `json:"bookings"`
	Attendances    []*types.Attendance       `json:"attendances"`
	CardPrices     []types.CardPrice         `json:"card_prices"`
}
