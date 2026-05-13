package main

import (
	"log"
	"net/http"
	"pdf-watermark/internal/handler"
)

func main() {
	h, err := handler.New()
	if err != nil {
		log.Fatal(err)
	}
	defer h.Close()

	mux := http.NewServeMux()
	mux.HandleFunc("/", h.ServeHome)
	mux.HandleFunc("/upload", h.HandleUpload)
	mux.HandleFunc("/jobs", h.ListJobs)
	mux.HandleFunc("/jobs/", h.GetJob)
	mux.HandleFunc("/download/", h.Download)

	log.Println("PDF水印服务启动在 http://localhost:9400")
	if err := http.ListenAndServe(":9400", mux); err != nil {
		log.Fatal(err)
	}
}
