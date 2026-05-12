package server

import (
	"conn-pool/internal/manager"
	"errors"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

type MaxOpenRequest struct {
	MaxOpen int `json:"max_open"`
}

func SetupRoutes(app *fiber.App, mgr *manager.Manager) {
	app.Get("/api/pools", listPoolsHandler(mgr))
	app.Get("/api/pools/:address", getPoolHandler(mgr))
	app.Post("/api/pools/:address/max-open", setMaxOpenHandler(mgr))
}

func listPoolsHandler(mgr *manager.Manager) fiber.Handler {
	return func(c *fiber.Ctx) error {
		info := mgr.GetAllInfo()
		return c.JSON(fiber.Map{
			"success": true,
			"data":    info,
		})
	}
}

func getPoolHandler(mgr *manager.Manager) fiber.Handler {
	return func(c *fiber.Ctx) error {
		address := c.Params("address")
		decoded, err := decodeURLParam(address)
		if err != nil {
			decoded = address
		}
		
		allInfo := mgr.GetAllInfo()
		info, exists := allInfo[decoded]
		if !exists {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"success": false,
				"error":   "pool not found",
			})
		}
		
		return c.JSON(fiber.Map{
			"success": true,
			"data":    info,
		})
	}
}

func setMaxOpenHandler(mgr *manager.Manager) fiber.Handler {
	return func(c *fiber.Ctx) error {
		address := c.Params("address")
		decoded, err := decodeURLParam(address)
		if err != nil {
			decoded = address
		}
		
		var req MaxOpenRequest
		if err := c.BodyParser(&req); err != nil {
			maxOpenStr := c.Query("max_open")
			if maxOpenStr == "" {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"success": false,
					"error":   "missing max_open parameter",
				})
			}
			parsed, err := strconv.Atoi(maxOpenStr)
			if err != nil {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"success": false,
					"error":   "invalid max_open value",
				})
			}
			req.MaxOpen = parsed
		}
		
		if req.MaxOpen < 1 {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"success": false,
				"error":   "max_open must be at least 1",
			})
		}
		
		ok := mgr.SetMaxOpen(decoded, req.MaxOpen)
		if !ok {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"success": false,
				"error":   "pool not found",
			})
		}
		
		return c.JSON(fiber.Map{
			"success": true,
			"message": "max_open updated",
		})
	}
}

func decodeURLParam(s string) (string, error) {
	result := ""
	for i := 0; i < len(s); i++ {
		if s[i] == '%' && i+2 < len(s) {
			hi := hexToByte(s[i+1])
			lo := hexToByte(s[i+2])
			if hi == 255 || lo == 255 {
				return "", errors.New("invalid hex")
			}
			result += string(hi*16 + lo)
			i += 2
		} else {
			result += string(s[i])
		}
	}
	return result, nil
}

func hexToByte(c byte) byte {
	switch {
	case c >= '0' && c <= '9':
		return c - '0'
	case c >= 'a' && c <= 'f':
		return 10 + c - 'a'
	case c >= 'A' && c <= 'F':
		return 10 + c - 'A'
	default:
		return 255
	}
}
