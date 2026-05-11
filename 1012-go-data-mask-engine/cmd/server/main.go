package main

import (
"flag"
"fmt"
"log"
"net/http"

"mask-engine/internal/config"
)

func main() {
port := flag.Int("port", 8102, "Server port")
configPath := flag.String("config", "rules.json", "Path to rules configuration file")
flag.Parse()

cm, err := config.NewManager(*configPath)
if err != nil {
log.Fatalf("Failed to initialize config manager: %v", err)
}

handler := NewHandler(cm)

http.HandleFunc("/mask", handler.HandleMask)
http.HandleFunc("/mask/type", handler.HandleMaskByType)
http.HandleFunc("/mask/rules", handler.HandleRules)

addr := fmt.Sprintf(":%d", *port)
log.Printf("Server starting on %s", addr)
log.Printf("Config file: %s", *configPath)
if err := http.ListenAndServe(addr, nil); err != nil {
log.Fatalf("Server failed: %v", err)
}
}
