package main

import (
	"flag"
	"log"
	"net/http"
	"os"

	"github.com/mqtt/parser/api"
	"github.com/mqtt/parser/mqtt"
)

type Server struct {
	willConfig *api.WillConfig
}

func main() {
	port := flag.String("port", "", "Server port")
	flag.Parse()

	if *port == "" {
		if envPort := os.Getenv("MQTT_PARSER_PORT"); envPort != "" {
			*port = envPort
		} else {
			*port = "8080"
		}
	}

	srv := &Server{
		willConfig: &api.WillConfig{
			Enabled: false,
			QoS:     0,
		},
	}

	http.HandleFunc("/parse", srv.handleParse)
	http.HandleFunc("/build", srv.handleBuild)
	http.HandleFunc("/will", srv.handleWill)

	log.Printf("MQTT Parser Server starting on port %s", *port)
	if err := http.ListenAndServe(":"+*port, nil); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}

func init() {
	_ = mqtt.Parse
}
