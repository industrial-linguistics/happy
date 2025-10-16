package main

import (
	"crypto/rand"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"os"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

const (
	dbPath        = "/vhosts/happy.industrial-linguistics.com/data/positive-social.db"
	maxNameLen    = 50
	maxMessageLen = 500
	rateLimit     = 100 // requests per minute per IP
)

type Handler struct {
	db *sql.DB
}

type MessageResponse struct {
	Name      string    `json:"name"`
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
	MessageID string    `json:"message_id"`
	Sequence  int       `json:"sequence"`
}

type PostMessageRequest struct {
	From      string `json:"from"`
	To        string `json:"to"`
	Message   string `json:"message"`
	SessionID string `json:"session_id,omitempty"`
}

type ErrorResponse struct {
	Error     string    `json:"error"`
	Timestamp time.Time `json:"timestamp"`
}

func main() {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	h := &Handler{db: db}

	// Parse CGI request
	method := os.Getenv("REQUEST_METHOD")
	scriptName := os.Getenv("SCRIPT_NAME")
	queryString := os.Getenv("QUERY_STRING")

	// Extract endpoint from SCRIPT_NAME (e.g., "/v1/status" -> "/status")
	endpoint := scriptName
	if strings.HasPrefix(scriptName, "/v1/") {
		endpoint = "/" + strings.TrimPrefix(scriptName, "/v1/")
	}

	startTime := time.Now()

	// Route request
	var statusCode int
	switch {
	case method == "GET" && endpoint == "/message":
		statusCode = h.handleGetMessage(queryString)
	case method == "POST" && endpoint == "/message":
		statusCode = h.handlePostMessage()
	case method == "GET" && endpoint == "/messages":
		statusCode = h.handleGetMessages(queryString)
	case method == "GET" && endpoint == "/status":
		statusCode = h.handleStatus()
	default:
		statusCode = h.handleNotFound()
	}

	// Log activity
	elapsed := time.Since(startTime).Milliseconds()
	h.logActivity(endpoint, statusCode, int(elapsed))
}

func (h *Handler) handleGetMessage(queryString string) int {
	values, err := url.ParseQuery(queryString)
	if err != nil {
		h.sendError(400, "Invalid query string")
		return 400
	}

	name := values.Get("name")
	if name == "" {
		h.sendError(400, "name parameter required")
		return 400
	}

	if len(name) > maxNameLen {
		h.sendError(400, "name too long")
		return 400
	}

	sessionID := values.Get("session_id")

	// Check rate limit
	ip := os.Getenv("REMOTE_ADDR")
	if !h.checkRateLimit(ip) {
		h.sendError(429, "Rate limit exceeded")
		return 429
	}

	// Get random message
	var message string
	err = h.db.QueryRow(`
        SELECT message FROM messages
        ORDER BY RANDOM() LIMIT 1
    `).Scan(&message)

	if err != nil {
		log.Printf("Error fetching message: %v", err)
		h.sendError(500, "Internal server error")
		return 500
	}

	// Get sequence number for this name
	var sequence int
	h.db.QueryRow(`
        SELECT COUNT(*) FROM activity_log
        WHERE name = ? AND endpoint = '/message'
    `, name).Scan(&sequence)
	sequence++ // This is their nth request

	messageID := generateMessageID()

	response := MessageResponse{
		Name:      name,
		Message:   message,
		Timestamp: time.Now(),
		MessageID: messageID,
		Sequence:  sequence,
	}

	h.sendJSON(200, response)

	// Log with session info
	h.logActivityWithDetails("/message", name, sessionID, ip, 200)

	return 200
}

func (h *Handler) handlePostMessage() int {
	var req PostMessageRequest
	if err := json.NewDecoder(os.Stdin).Decode(&req); err != nil {
		h.sendError(400, "Invalid JSON")
		return 400
	}

	// Validation
	if req.From == "" || req.To == "" || req.Message == "" {
		h.sendError(400, "from, to, and message are required")
		return 400
	}

	if len(req.Message) > maxMessageLen {
		h.sendError(400, "message too long")
		return 400
	}

	// Basic positivity check (simple keyword filter)
	if !isPositive(req.Message) {
		h.sendError(400, "message must be positive")
		return 400
	}

	ip := os.Getenv("REMOTE_ADDR")
	messageID := generateMessageID()

	_, err := h.db.Exec(`
        INSERT INTO user_messages (message_id, from_user, to_user, message, ip_address)
        VALUES (?, ?, ?, ?, ?)
    `, messageID, req.From, req.To, req.Message, ip)

	if err != nil {
		log.Printf("Error saving message: %v", err)
		h.sendError(500, "Internal server error")
		return 500
	}

	response := map[string]interface{}{
		"message_id": messageID,
		"timestamp":  time.Now(),
		"status":     "delivered",
	}

	h.sendJSON(201, response)
	h.logActivityWithDetails("/message", req.From, req.SessionID, ip, 201)

	return 201
}

func (h *Handler) handleGetMessages(queryString string) int {
	values, err := url.ParseQuery(queryString)
	if err != nil {
		h.sendError(400, "Invalid query string")
		return 400
	}

	recipient := values.Get("recipient")
	if recipient == "" {
		h.sendError(400, "recipient parameter required")
		return 400
	}

	if len(recipient) > maxNameLen {
		h.sendError(400, "recipient name too long")
		return 400
	}

	sessionID := values.Get("session_id")

	// Check rate limit
	ip := os.Getenv("REMOTE_ADDR")
	if !h.checkRateLimit(ip) {
		h.sendError(429, "Rate limit exceeded")
		return 429
	}

	// Get random message
	var message string
	err = h.db.QueryRow(`
        SELECT message FROM messages
        ORDER BY RANDOM() LIMIT 1
    `).Scan(&message)

	if err != nil {
		log.Printf("Error fetching message: %v", err)
		h.sendError(500, "Internal server error")
		return 500
	}

	// Get sequence number for this recipient
	var sequence int
	h.db.QueryRow(`
        SELECT COUNT(*) FROM activity_log
        WHERE name = ? AND endpoint = '/messages'
    `, recipient).Scan(&sequence)
	sequence++ // This is their nth request

	messageID := generateMessageID()

	response := MessageResponse{
		Name:      recipient,
		Message:   message,
		Timestamp: time.Now(),
		MessageID: messageID,
		Sequence:  sequence,
	}

	h.sendJSON(200, response)

	// Log with recipient name and session info
	h.logActivityWithDetails("/messages", recipient, sessionID, ip, 200)

	return 200
}

func (h *Handler) handleStatus() int {
	var requestsToday int
	today := time.Now().Format("2006-01-02")

	h.db.QueryRow(`
        SELECT COUNT(*) FROM activity_log
        WHERE date(timestamp) = ?
    `, today).Scan(&requestsToday)

	response := map[string]interface{}{
		"status":         "ok",
		"version":        "1.0.0",
		"requests_today": requestsToday,
	}

	h.sendJSON(200, response)
	return 200
}

func (h *Handler) handleNotFound() int {
	h.sendError(404, "Endpoint not found")
	return 404
}

func (h *Handler) sendJSON(code int, data interface{}) {
	fmt.Printf("Status: %d\r\n", code)
	fmt.Printf("Content-Type: application/json\r\n")
	fmt.Printf("Access-Control-Allow-Origin: *\r\n")
	fmt.Printf("\r\n")
	json.NewEncoder(os.Stdout).Encode(data)
}

func (h *Handler) sendError(code int, message string) {
	h.sendJSON(code, ErrorResponse{
		Error:     message,
		Timestamp: time.Now(),
	})
}

func (h *Handler) checkRateLimit(ip string) bool {
	bucket := time.Now().Format("2006-01-02 15:04")

	var count int
	h.db.QueryRow(`
        SELECT request_count FROM request_stats
        WHERE ip_address = ? AND minute_bucket = ?
    `, ip, bucket).Scan(&count)

	if count >= rateLimit {
		return false
	}

	h.db.Exec(`
        INSERT INTO request_stats (ip_address, minute_bucket, request_count)
        VALUES (?, ?, 1)
        ON CONFLICT(ip_address, minute_bucket)
        DO UPDATE SET request_count = request_count + 1
    `, ip, bucket)

	return true
}

func (h *Handler) logActivity(endpoint string, statusCode, responseTimeMs int) {
	ip := os.Getenv("REMOTE_ADDR")
	userAgent := os.Getenv("HTTP_USER_AGENT")

	h.db.Exec(`
        INSERT INTO activity_log
        (endpoint, ip_address, user_agent, response_code, response_time_ms)
        VALUES (?, ?, ?, ?, ?)
    `, endpoint, ip, userAgent, statusCode, responseTimeMs)
}

func (h *Handler) logActivityWithDetails(endpoint, name, sessionID, ip string, statusCode int) {
	userAgent := os.Getenv("HTTP_USER_AGENT")

	h.db.Exec(`
        INSERT INTO activity_log
        (endpoint, name, session_id, ip_address, user_agent, response_code)
        VALUES (?, ?, ?, ?, ?, ?)
    `, endpoint, name, sessionID, ip, userAgent, statusCode)
}

func generateMessageID() string {
	b := make([]byte, 8)
	rand.Read(b)
	return fmt.Sprintf("msg_%x", b)
}

func isPositive(message string) bool {
	// Simple keyword filter - reject obvious negativity
	negative := []string{"hate", "stupid", "idiot", "bad", "terrible", "awful", "suck", "dumb"}
	lower := strings.ToLower(message)

	for _, word := range negative {
		if strings.Contains(lower, word) {
			return false
		}
	}
	return true
}
