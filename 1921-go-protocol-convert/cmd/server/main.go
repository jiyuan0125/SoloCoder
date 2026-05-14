package main

import (
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	
	"xml-json-converter/internal/handler"
	"xml-json-converter/internal/mapping"
)

func main() {
	app := fiber.New()
	
	store := mapping.NewStore()
	
	convertHandler := handler.NewConvertHandler(store)
	mappingHandler := handler.NewMappingHandler(store)
	
	api := app.Group("/convert")
	api.Post("/xml2json", convertHandler.XML2JSON)
	api.Post("/json2xml", convertHandler.JSON2XML)
	
	mappings := app.Group("/mappings")
	mappings.Post("/:message_type", mappingHandler.Register)
	mappings.Put("/:message_type", mappingHandler.Update)
	mappings.Get("/:message_type", mappingHandler.Get)
	
	port := os.Getenv("PORT")
	if port == "" {
		port = "9204"
	}
	
	log.Printf("Server starting on port %s...", port)
	if err := app.Listen(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
