package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"smtp-client/common"
	"smtp-client/smtpclient"
)

var (
	smtpHost     string
	smtpPort     int
	smtpUser     string
	smtpPassword string
	smtpFrom     string
	useTLS       bool

	queue      = make(map[string]*common.QueueItem)
	queueMutex sync.RWMutex
	queueOrder []string
)

func main() {
	flag.StringVar(&smtpHost, "smtp-host", getEnvOrDefault("SMTP_HOST", ""), "SMTP server host")
	flag.IntVar(&smtpPort, "smtp-port", getEnvIntOrDefault("SMTP_PORT", 587), "SMTP server port")
	flag.StringVar(&smtpUser, "smtp-user", getEnvOrDefault("SMTP_USER", ""), "SMTP username")
	flag.StringVar(&smtpPassword, "smtp-password", getEnvOrDefault("SMTP_PASSWORD", ""), "SMTP password")
	flag.StringVar(&smtpFrom, "smtp-from", getEnvOrDefault("SMTP_FROM", ""), "SMTP from address")
	flag.BoolVar(&useTLS, "smtp-tls", getEnvBoolOrDefault("SMTP_TLS", true), "Use STARTTLS")
	
	port := flag.String("port", getEnvOrDefault("PORT", "8080"), "HTTP server port")
	flag.Parse()

	if smtpHost == "" {
		log.Fatal("SMTP host is required. Set via -smtp-host or SMTP_HOST environment variable")
	}

	http.HandleFunc("/api/emails", handleEmails)
	http.HandleFunc("/api/emails/", handleEmailByID)

	log.Printf("Server starting on port %s", *port)
	if err := http.ListenAndServe(":"+*port, nil); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}

func getEnvOrDefault(key, defaultValue string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultValue
}

func getEnvIntOrDefault(key string, defaultValue int) int {
	if v := os.Getenv(key); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}
	return defaultValue
}

func getEnvBoolOrDefault(key string, defaultValue bool) bool {
	if v := os.Getenv(key); v != "" {
		if b, err := strconv.ParseBool(v); err == nil {
			return b
		}
	}
	return defaultValue
}

func handleEmails(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		handleSendEmail(w, r)
	case http.MethodGet:
		handleGetQueue(w, r)
	default:
		sendError(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func handleEmailByID(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		sendError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	id := strings.TrimPrefix(r.URL.Path, "/api/emails/")
	if id == "" {
		sendError(w, "Email ID is required", http.StatusBadRequest)
		return
	}

	queueMutex.Lock()
	defer queueMutex.Unlock()

	if _, exists := queue[id]; !exists {
		sendError(w, "Email not found", http.StatusNotFound)
		return
	}

	delete(queue, id)
	for i, qid := range queueOrder {
		if qid == id {
			queueOrder = append(queueOrder[:i], queueOrder[i+1:]...)
			break
		}
	}

	sendJSON(w, common.SendEmailResponse{
		Success: true,
		Message: "Email removed from queue",
	}, http.StatusOK)
}

func handleSendEmail(w http.ResponseWriter, r *http.Request) {
	var req common.SendEmailRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	if req.From == "" {
		req.From = smtpFrom
	}

	id := generateID()
	item := &common.QueueItem{
		ID:        id,
		Request:   req,
		Status:    "pending",
		CreatedAt: time.Now().Format(time.RFC3339),
	}

	queueMutex.Lock()
	queue[id] = item
	queueOrder = append(queueOrder, id)
	queueMutex.Unlock()

	go sendEmailAsync(item)

	sendJSON(w, common.SendEmailResponse{
		Success: true,
		ID:      id,
		Message: "Email queued for sending",
	}, http.StatusAccepted)
}

func handleGetQueue(w http.ResponseWriter, r *http.Request) {
	queueMutex.RLock()
	defer queueMutex.RUnlock()

	items := make([]common.QueueItem, 0, len(queueOrder))
	for _, id := range queueOrder {
		if item, exists := queue[id]; exists {
			items = append(items, *item)
		}
	}

	sendJSON(w, common.GetQueueResponse{Items: items}, http.StatusOK)
}

func sendEmailAsync(item *common.QueueItem) {
	defer func() {
		if r := recover(); r != nil {
			queueMutex.Lock()
			item.Status = "failed"
			item.ErrorMessage = fmt.Sprintf("Panic: %v", r)
			queueMutex.Unlock()
		}
	}()

	queueMutex.Lock()
	item.Status = "sending"
	queueMutex.Unlock()

	cfg := &smtpclient.Config{
		Host:     smtpHost,
		Port:     smtpPort,
		Username: smtpUser,
		Password: smtpPassword,
		UseTLS:   useTLS,
		From:     item.Request.From,
	}

	client := smtpclient.NewClient(cfg)
	if err := client.Connect(); err != nil {
		queueMutex.Lock()
		item.Status = "failed"
		item.ErrorMessage = err.Error()
		queueMutex.Unlock()
		return
	}
	defer client.Close()

	msg := smtpclient.NewMessage()
	msg.From = item.Request.From
	msg.To = item.Request.To
	msg.Cc = item.Request.Cc
	msg.Bcc = item.Request.Bcc
	msg.Subject = item.Request.Subject
	msg.TextBody = item.Request.TextBody
	msg.HTMLBody = item.Request.HTMLBody

	for _, att := range item.Request.Attachments {
		msg.Attachments = append(msg.Attachments, smtpclient.Attachment{
			Filename:    att.Filename,
			Data:        att.Data,
			ContentType: att.ContentType,
		})
	}

	if err := client.Send(msg); err != nil {
		queueMutex.Lock()
		item.Status = "failed"
		item.ErrorMessage = err.Error()
		queueMutex.Unlock()
		return
	}

	queueMutex.Lock()
	item.Status = "sent"
	queueMutex.Unlock()
}

func generateID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

func sendJSON(w http.ResponseWriter, data interface{}, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func sendError(w http.ResponseWriter, message string, status int) {
	sendJSON(w, common.ErrorResponse{
		Success: false,
		Message: message,
	}, status)
}
