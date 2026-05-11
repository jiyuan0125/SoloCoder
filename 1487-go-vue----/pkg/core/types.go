package core

import (
	"time"
)

type ItemSize string

const (
	SizeSmall  ItemSize = "small"
	SizeMedium ItemSize = "medium"
	SizeLarge  ItemSize = "large"
)

type ItemCategory string

const (
	CategoryFurniture ItemCategory = "furniture"
	CategoryAppliance ItemCategory = "appliance"
	CategoryBox       ItemCategory = "box"
	CategorySpecial   ItemCategory = "special"
)

type TimeSlot string

const (
	SlotMorning  TimeSlot = "morning"
	SlotAfternoon TimeSlot = "afternoon"
	SlotEvening  TimeSlot = "evening"
)

type BookingStatus string

const (
	StatusPending   BookingStatus = "pending"
	StatusConfirmed BookingStatus = "confirmed"
	StatusCancelled BookingStatus = "cancelled"
	StatusCompleted BookingStatus = "completed"
	StatusSettled   BookingStatus = "settled"
)

type Furniture struct {
	Name string   `json:"name"`
	Size ItemSize `json:"size"`
}

type Appliance struct {
	Name     string `json:"name"`
	NeedDisassemble bool `json:"needDisassemble"`
}

type ItemList struct {
	Furniture  []Furniture `json:"furniture"`
	Appliances []Appliance `json:"appliances"`
	BoxCount   int         `json:"boxCount"`
	Specials   []string    `json:"specials"`
}

type Booking struct {
	ID              string        `json:"id"`
	CustomerName    string        `json:"customerName"`
	Phone           string        `json:"phone"`
	MoveDate        time.Time     `json:"moveDate"`
	TimeSlot        TimeSlot      `json:"timeSlot"`
	FromAddress     string        `json:"fromAddress"`
	ToAddress       string        `json:"toAddress"`
	DistanceKM      float64       `json:"distanceKM"`
	Items           ItemList      `json:"items"`
	Estimate        int64         `json:"estimate"`
	FinalAmount     int64         `json:"finalAmount,omitempty"`
	Status          BookingStatus `json:"status"`
	CreatedAt       time.Time     `json:"createdAt"`
	UpdatedAt       time.Time     `json:"updatedAt"`
	Review          *Review       `json:"review,omitempty"`
}

type Review struct {
	Rating  int    `json:"rating"`
	Comment string `json:"comment"`
}

type Vehicle struct {
	ID       string `json:"id"`
	Capacity int    `json:"capacity"`
}
