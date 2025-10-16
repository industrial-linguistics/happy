package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
)

type PostMessageRequest struct {
	From      string `json:"from"`
	To        string `json:"to"`
	Message   string `json:"message"`
	SessionID string `json:"session_id,omitempty"`
}

type MessageResponse struct {
	MessageID string `json:"message_id"`
	Timestamp string `json:"timestamp"`
	Status    string `json:"status"`
}

type ErrorResponse struct {
	Error     string `json:"error"`
	Timestamp string `json:"timestamp"`
}

func main() {
	from := flag.String("from", "", "Your name (required)")
	to := flag.String("to", "", "Recipient name (required)")
	message := flag.String("message", "", "The message to send (required)")
	sessionID := flag.String("session", "", "Optional session ID")
	baseURL := flag.String("url", "https://happy.industrial-linguistics.com/v1", "Base URL for the API")
	flag.Parse()

	// Validate required fields
	if *from == "" || *to == "" || *message == "" {
		fmt.Fprintf(os.Stderr, "Error: from, to, and message are all required\n\n")
		fmt.Fprintf(os.Stderr, "Usage:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExample:\n")
		fmt.Fprintf(os.Stderr, "  send-message -from Alice -to Bob -message \"You're doing great!\"\n")
		os.Exit(1)
	}

	// Build request
	req := PostMessageRequest{
		From:      *from,
		To:        *to,
		Message:   *message,
		SessionID: *sessionID,
	}

	jsonData, err := json.Marshal(req)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error encoding JSON: %v\n", err)
		os.Exit(1)
	}

	// Send request
	url := fmt.Sprintf("%s/message", *baseURL)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error sending request: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading response: %v\n", err)
		os.Exit(1)
	}

	// Handle response
	if resp.StatusCode == 201 {
		var msgResp MessageResponse
		if err := json.Unmarshal(body, &msgResp); err == nil {
			fmt.Printf("✓ Message sent successfully!\n")
			fmt.Printf("  Message ID: %s\n", msgResp.MessageID)
			fmt.Printf("  Status: %s\n", msgResp.Status)
			fmt.Printf("  From: %s\n", *from)
			fmt.Printf("  To: %s\n", *to)
			fmt.Printf("  Message: %s\n", *message)
		} else {
			fmt.Printf("✓ Message sent (HTTP %d)\n", resp.StatusCode)
		}
	} else {
		var errResp ErrorResponse
		if err := json.Unmarshal(body, &errResp); err == nil {
			fmt.Fprintf(os.Stderr, "❌ Error: %s (HTTP %d)\n", errResp.Error, resp.StatusCode)
		} else {
			fmt.Fprintf(os.Stderr, "❌ HTTP %d: %s\n", resp.StatusCode, string(body))
		}
		os.Exit(1)
	}
}
