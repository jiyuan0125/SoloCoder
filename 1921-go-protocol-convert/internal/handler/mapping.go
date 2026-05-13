package handler

import (
	"log"

	"github.com/gofiber/fiber/v2"
	
	"xml-json-converter/internal/mapping"
)

type MappingHandler struct {
	store *mapping.Store
}

func NewMappingHandler(store *mapping.Store) *MappingHandler {
	log.Printf("MappingHandler: Store at %p", store)
	return &MappingHandler{store: store}
}

func (h *MappingHandler) Register(c *fiber.Ctx) error {
	messageType := c.Params("message_type")
	log.Printf("[Register] Store: %p, messageType: %s", h.store, messageType)
	
	var m mapping.Mapping
	if err := c.BodyParser(&m); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid mapping: " + err.Error(),
		})
	}
	
	m.MessageType = messageType
	h.store.Register(m)
	
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "mapping registered",
		"message_type": messageType,
	})
}

func (h *MappingHandler) Update(c *fiber.Ctx) error {
	messageType := c.Params("message_type")
	
	var m mapping.Mapping
	if err := c.BodyParser(&m); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid mapping: " + err.Error(),
		})
	}
	
	if !h.store.Update(messageType, m) {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "message_type not found",
			"message_type": messageType,
		})
	}
	
	return c.JSON(fiber.Map{
		"message": "mapping updated",
		"message_type": messageType,
	})
}

func (h *MappingHandler) Get(c *fiber.Ctx) error {
	messageType := c.Params("message_type")
	
	m, ok := h.store.Get(messageType)
	if !ok {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "message_type not found",
			"message_type": messageType,
		})
	}
	
	return c.JSON(m)
}
