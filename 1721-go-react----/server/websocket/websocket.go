package websocket

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"time"

	"education-platform/database"
	"education-platform/models"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type Client struct {
	conn      *websocket.Conn
	courseID  int64
	studentID int64
	send      chan []byte
}

type Hub struct {
	courses map[int64]map[*Client]bool
	mutex   sync.RWMutex
}

var hub = &Hub{
	courses: make(map[int64]map[*Client]bool),
}

func (h *Hub) addClient(client *Client) {
	h.mutex.Lock()
	defer h.mutex.Unlock()
	if h.courses[client.courseID] == nil {
		h.courses[client.courseID] = make(map[*Client]bool)
	}
	h.courses[client.courseID][client] = true

	database.DB.Exec(`INSERT INTO live_sessions (course_id, student_id, join_time) VALUES (?, ?, ?)`,
		client.courseID, client.studentID, time.Now().Format(time.RFC3339))
}

func (h *Hub) removeClient(client *Client) {
	h.mutex.Lock()
	defer h.mutex.Unlock()
	if clients, ok := h.courses[client.courseID]; ok {
		if _, exists := clients[client]; exists {
			delete(clients, client)
			close(client.send)
		}
	}

	now := time.Now().Format(time.RFC3339)
	var joinTimeStr string
	database.DB.QueryRow(`SELECT join_time FROM live_sessions WHERE course_id = ? AND student_id = ? AND leave_time IS NULL ORDER BY join_time DESC LIMIT 1`,
		client.courseID, client.studentID).Scan(&joinTimeStr)

	if joinTime, err := time.Parse(time.RFC3339, joinTimeStr); err == nil {
		duration := int(time.Since(joinTime).Seconds())
		database.DB.Exec(`UPDATE live_sessions SET leave_time = ?, duration_seconds = ? WHERE course_id = ? AND student_id = ? AND leave_time IS NULL`,
			now, duration, client.courseID, client.studentID)
	}
}

func (h *Hub) broadcast(courseID int64, message []byte) {
	h.mutex.RLock()
	defer h.mutex.RUnlock()
	if clients, ok := h.courses[courseID]; ok {
		for client := range clients {
			select {
			case client.send <- message:
			default:
				close(client.send)
				delete(clients, client)
			}
		}
	}
}

func (h *Hub) getOnlineCount(courseID int64) int {
	h.mutex.RLock()
	defer h.mutex.RUnlock()
	if clients, ok := h.courses[courseID]; ok {
		return len(clients)
	}
	return 0
}

type WebSocketMessage struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
}

type DanmakuPayload struct {
	StudentID int64  `json:"student_id"`
	Content   string `json:"content"`
	Time      string `json:"time"`
}

func HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	courseID, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		http.Error(w, "Invalid course ID", http.StatusBadRequest)
		return
	}

	studentIDStr := r.URL.Query().Get("student_id")
	studentID, err := strconv.ParseInt(studentIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid student ID", http.StatusBadRequest)
		return
	}

	var courseType string
	var liveStatus string
	var startTimeStr string
	err = database.DB.QueryRow(`SELECT course_type, live_status, start_time FROM courses WHERE id = ?`, courseID).Scan(&courseType, &liveStatus, &startTimeStr)
	if err != nil {
		http.Error(w, "Course not found", http.StatusNotFound)
		return
	}

	if courseType != "live" {
		http.Error(w, "Not a live course", http.StatusBadRequest)
		return
	}

	var startTime time.Time
	if startTimeStr != "" {
		startTime, _ = time.Parse(time.RFC3339, startTimeStr)
	}

	now := time.Now()
	canJoin := now.After(startTime.Add(-30 * time.Minute))
	if !canJoin {
		http.Error(w, "Cannot join yet - room opens 30 minutes before start", http.StatusBadRequest)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println("WebSocket upgrade error:", err)
		return
	}

	client := &Client{
		conn:      conn,
		courseID:  courseID,
		studentID: studentID,
		send:      make(chan []byte, 256),
	}

	hub.addClient(client)
	defer hub.removeClient(client)

	go writePump(client)
	readPump(client, liveStatus, startTime)
}

func readPump(client *Client, liveStatus string, startTime time.Time) {
	defer client.conn.Close()

	for {
		_, message, err := client.conn.ReadMessage()
		if err != nil {
			break
		}

		var msg WebSocketMessage
		if err := json.Unmarshal(message, &msg); err != nil {
			continue
		}

		canInteract := liveStatus == "in_progress"

		switch msg.Type {
		case "danmaku":
			if !canInteract {
				sendError(client, "Cannot send danmaku - live not started or already ended")
				continue
			}

			var payload DanmakuPayload
			payloadBytes, _ := json.Marshal(msg.Payload)
			json.Unmarshal(payloadBytes, &payload)

			payload.StudentID = client.studentID
			payload.Time = time.Now().Format(time.RFC3339)

			database.DB.Exec(`INSERT INTO danmakus (course_id, student_id, content, sent_at) VALUES (?, ?, ?, ?)`,
				client.courseID, client.studentID, payload.Content, payload.Time)

			broadcastMessage, _ := json.Marshal(WebSocketMessage{
				Type:    "danmaku",
				Payload: payload,
			})
			hub.broadcast(client.courseID, broadcastMessage)

		case "vote":
			if !canInteract {
				sendError(client, "Cannot vote - live not started or already ended")
				continue
			}

			var payload struct {
				VoteID   int64 `json:"vote_id"`
				OptionID int64 `json:"option_id"`
			}
			payloadBytes, _ := json.Marshal(msg.Payload)
			json.Unmarshal(payloadBytes, &payload)

			var isActive bool
			database.DB.QueryRow(`SELECT is_active FROM votes WHERE id = ?`, payload.VoteID).Scan(&isActive)
			if !isActive {
				sendError(client, "Vote is not active")
				continue
			}

			_, err := database.DB.Exec(`INSERT INTO vote_participants (vote_id, student_id, option_id) VALUES (?, ?, ?)`,
				payload.VoteID, client.studentID, payload.OptionID)
			if err != nil {
				sendError(client, "Already voted")
				continue
			}

			database.DB.Exec(`UPDATE vote_options SET votes = votes + 1 WHERE id = ?`, payload.OptionID)

			broadcastVoteResults(client.courseID, payload.VoteID)

		case "ping":
			response, _ := json.Marshal(WebSocketMessage{Type: "pong", Payload: map[string]int{"online": hub.getOnlineCount(client.courseID)}})
			client.send <- response
		}
	}
}

func writePump(client *Client) {
	defer client.conn.Close()
	for message := range client.send {
		if err := client.conn.WriteMessage(websocket.TextMessage, message); err != nil {
			break
		}
	}
}

func sendError(client *Client, message string) {
	errorMsg, _ := json.Marshal(WebSocketMessage{
		Type:    "error",
		Payload: map[string]string{"message": message},
	})
	client.send <- errorMsg
}

func broadcastVoteResults(courseID int64, voteID int64) {
	var vote models.Vote
	database.DB.QueryRow(`SELECT id, course_id, title, start_time, is_active FROM votes WHERE id = ?`, voteID).Scan(
		&vote.ID, &vote.CourseID, &vote.Title, &vote.StartTime, &vote.IsActive)

	rows, _ := database.DB.Query(`SELECT id, vote_id, text, votes FROM vote_options WHERE vote_id = ?`, voteID)
	vote.Options = []models.VoteOption{}
	var totalVotes int
	for rows.Next() {
		var opt models.VoteOption
		rows.Scan(&opt.ID, &opt.VoteID, &opt.Text, &opt.Votes)
		vote.Options = append(vote.Options, opt)
		totalVotes += opt.Votes
	}

	results := []models.VoteResult{}
	for _, opt := range vote.Options {
		percentage := 0.0
		if totalVotes > 0 {
			percentage = float64(opt.Votes) / float64(totalVotes) * 100
		}
		results = append(results, models.VoteResult{
			OptionID:   opt.ID,
			OptionText: opt.Text,
			Votes:      opt.Votes,
			Percentage: percentage,
		})
	}

	msg, _ := json.Marshal(WebSocketMessage{
		Type: "vote_result",
		Payload: map[string]interface{}{
			"vote_id": voteID,
			"title":   vote.Title,
			"results": results,
		},
	})
	hub.broadcast(courseID, msg)
}

func CreateVote(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	courseID, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		http.Error(w, "Invalid course ID", http.StatusBadRequest)
		return
	}

	var req struct {
		Title   string   `json:"title"`
		Options []string `json:"options"`
	}
	json.NewDecoder(r.Body).Decode(&req)

	if len(req.Options) < 2 || len(req.Options) > 5 {
		http.Error(w, "Vote must have 2-5 options", http.StatusBadRequest)
		return
	}

	tx, err := database.DB.Begin()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	result, err := tx.Exec(`INSERT INTO votes (course_id, title, is_active) VALUES (?, ?, 1)`, courseID, req.Title)
	if err != nil {
		tx.Rollback()
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	voteID, _ := result.LastInsertId()

	for _, opt := range req.Options {
		_, err = tx.Exec(`INSERT INTO vote_options (vote_id, text, votes) VALUES (?, ?, 0)`, voteID, opt)
		if err != nil {
			tx.Rollback()
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	tx.Commit()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"vote_id": voteID})
}

func GetActiveVotes(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	courseID, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		http.Error(w, "Invalid course ID", http.StatusBadRequest)
		return
	}

	rows, err := database.DB.Query(`SELECT id, course_id, title, start_time, end_time, is_active FROM votes WHERE course_id = ? AND is_active = 1`, courseID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	votes := []models.Vote{}
	for rows.Next() {
		var v models.Vote
		var endTime sql.NullString
		err := rows.Scan(&v.ID, &v.CourseID, &v.Title, &v.StartTime, &endTime, &v.IsActive)
		if err != nil {
			continue
		}
		if endTime.Valid {
			if t, err := time.Parse(time.RFC3339, endTime.String); err == nil {
				v.EndTime = &t
			}
		}

		optRows, _ := database.DB.Query(`SELECT id, vote_id, text, votes FROM vote_options WHERE vote_id = ?`, v.ID)
		v.Options = []models.VoteOption{}
		for optRows.Next() {
			var opt models.VoteOption
			optRows.Scan(&opt.ID, &opt.VoteID, &opt.Text, &opt.Votes)
			v.Options = append(v.Options, opt)
		}
		optRows.Close()

		votes = append(votes, v)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(votes)
}
