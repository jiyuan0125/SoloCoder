package main

import (
	"os"

	"github.com/gin-gonic/gin"

	"json-protobuf-gateway/handler"
	"json-protobuf-gateway/types"
)

func main() {
	store := types.NewSchemaStore()
	h := handler.NewHandler(store)

	r := gin.Default()

	r.POST("/schemas/:message_type", h.RegisterSchema)
	r.PUT("/schemas/:message_type", h.UpdateSchema)
	r.POST("/convert/:message_type", h.ConvertForward)
	r.POST("/convert/:message_type/reverse", h.ConvertReverse)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	r.Run(":" + port)
}
