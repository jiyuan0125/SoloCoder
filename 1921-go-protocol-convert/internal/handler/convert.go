package handler

import (
	"encoding/json"

	"github.com/gofiber/fiber/v2"
	
	"xml-json-converter/internal/converter"
	"xml-json-converter/internal/mapping"
)

type ConvertHandler struct {
	store *mapping.Store
}

func NewConvertHandler(store *mapping.Store) *ConvertHandler {
	return &ConvertHandler{store: store}
}

func getMessageType(c *fiber.Ctx, body []byte) string {
	if mt := c.Query("message_type"); mt != "" {
		return mt
	}
	if mt := c.Get("X-Message-Type"); mt != "" {
		return mt
	}
	
	var bodyMap map[string]interface{}
	if err := json.Unmarshal(body, &bodyMap); err == nil {
		if mt, ok := bodyMap["message_type"].(string); ok && mt != "" {
			return mt
		}
	}
	
	return ""
}

func (h *ConvertHandler) XML2JSON(c *fiber.Ctx) error {
	body := c.Body()
	messageType := getMessageType(c, body)
	
	if messageType != "" {
		if !h.store.Exists(messageType) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "message_type not registered",
				"message_type": messageType,
			})
		}
	}
	
	jsonData, err := converter.XmlToJson(body)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "failed to parse XML: " + err.Error(),
		})
	}
	
	if messageType != "" {
		if m, ok := h.store.Get(messageType); ok {
			jsonData = mapping.ApplyXML2JSONMappings(jsonData, m.XML2JSON)
		}
	}
	
	jsonOutput, err := converter.MapToJson(jsonData)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to marshal JSON: " + err.Error(),
		})
	}
	
	c.Set("Content-Type", "application/json")
	return c.Send(jsonOutput)
}

func (h *ConvertHandler) JSON2XML(c *fiber.Ctx) error {
	body := c.Body()
	messageType := getMessageType(c, body)
	
	var req map[string]interface{}
	if err := json.Unmarshal(body, &req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "failed to parse JSON: " + err.Error(),
		})
	}
	
	delete(req, "message_type")
	
	if messageType != "" {
		if !h.store.Exists(messageType) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "message_type not registered",
				"message_type": messageType,
			})
		}
		
		if m, ok := h.store.Get(messageType); ok {
			req = mapping.ApplyJSON2XMLMappings(req, m.JSON2XML)
		}
	}
	
	xmlOutput, err := converter.MapToXml(req)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to marshal XML: " + err.Error(),
		})
	}
	
	c.Set("Content-Type", "application/xml")
	return c.Send(xmlOutput)
}
