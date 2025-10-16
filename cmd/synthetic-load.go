package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"time"
)

var users = []string{
	"Alice",
	"Bob",
	"Charlie",
	"Diana",
	"Eve",
	"Frank",
	"Grace",
	"Henry",
	"Iris",
	"Jack",
}

type MessageResponse struct {
	Name      string    `json:"name"`
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
	MessageID string    `json:"message_id"`
	Sequence  int       `json:"sequence"`
}

func main() {
	baseURL := flag.String("url", "http://localhost/v1", "Base URL for the API")
	delay := flag.Int("delay", 2000, "Average delay between requests in milliseconds")
	flag.Parse()

	rand.Seed(time.Now().UnixNano())

	fmt.Printf("Starting synthetic load generator...\n")
	fmt.Printf("API URL: %s\n", *baseURL)
	fmt.Printf("Average delay: %dms\n\n", *delay)

	requestCount := 0

	for {
		// Pick a random user
		user := users[rand.Intn(len(users))]

		// Pick a random endpoint (message or messages)
		var endpoint, param string
		if rand.Float32() < 0.5 {
			endpoint = "message"
			param = "name"
		} else {
			endpoint = "messages"
			param = "recipient"
		}

		// Make request
		url := fmt.Sprintf("%s/%s?%s=%s", *baseURL, endpoint, param, user)

		resp, err := http.Get(url)
		if err != nil {
			fmt.Printf("❌ Error requesting %s: %v\n", user, err)
			time.Sleep(time.Duration(*delay) * time.Millisecond)
			continue
		}

		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()

		requestCount++

		if resp.StatusCode == 200 {
			var msgResp MessageResponse
			if err := json.Unmarshal(body, &msgResp); err == nil {
				fmt.Printf("✓ [%d] %s (#%d): %s\n",
					requestCount,
					msgResp.Name,
					msgResp.Sequence,
					truncate(msgResp.Message, 50))
			} else {
				fmt.Printf("✓ [%d] %s: 200 OK\n", requestCount, user)
			}
		} else {
			fmt.Printf("❌ [%d] %s: HTTP %d\n", requestCount, user, resp.StatusCode)
		}

		// Random delay (between delay/2 and delay*1.5)
		minDelay := *delay / 2
		maxDelay := (*delay * 3) / 2
		randomDelay := minDelay + rand.Intn(maxDelay-minDelay)
		time.Sleep(time.Duration(randomDelay) * time.Millisecond)
	}
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
