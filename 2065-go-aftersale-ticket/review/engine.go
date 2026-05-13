package review

import (
	"errors"
	"time"

	"aftersale-ticket/db"
	"aftersale-ticket/models"
)

type Engine struct {
	store *db.Store
}

func NewEngine(store *db.Store) *Engine {
	return &Engine{store: store}
}

type ReviewResult struct {
	Approved     bool
	OverdueDays  int
	ErrorMessage string
}

func (e *Engine) Review(ticket *models.Ticket) (*ReviewResult, error) {
	if ticket.Status != models.StatusPendingReview {
		return nil, errors.New("ticket is not in pending review status")
	}

	order, err := e.store.GetOrder(ticket.OrderID)
	if err != nil {
		return nil, err
	}

	warrantyDays := ticket.GetWarrantyDays()
	elapsedDays := int(time.Since(order.CompletedAt).Hours() / 24)

	if elapsedDays > warrantyDays {
		return &ReviewResult{
			Approved:     false,
			OverdueDays:  elapsedDays - warrantyDays,
			ErrorMessage: "warranty period expired",
		}, nil
	}

	return &ReviewResult{
		Approved:     true,
		OverdueDays:  0,
		ErrorMessage: "",
	}, nil
}

func (e *Engine) CheckExchangeInventory(ticket *models.Ticket) (*ReviewResult, error) {
	if ticket.Type != models.TicketTypeExchange {
		return nil, errors.New("not an exchange ticket")
	}

	order, err := e.store.GetOrder(ticket.OrderID)
	if err != nil {
		return nil, err
	}

	item, err := e.store.GetInventory(order.ProductSKU)
	if err != nil {
		return nil, err
	}

	if item.Stock < order.Quantity {
		return &ReviewResult{
			Approved:     false,
			ErrorMessage: "out of stock",
		}, nil
	}

	return &ReviewResult{
		Approved: true,
	}, nil
}
