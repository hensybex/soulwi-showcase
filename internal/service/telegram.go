package service

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
)

func SendLogToTelegram(logMessage string) {
	const maxMessageLength = 4096

	fmt.Println("Sending tg log")
	botToken := os.Getenv("TG_BOT_TOKEN")
	if botToken == "" {
		fmt.Println("Bot token not set!")
		return
	}

	// List of Telegram user IDs
	chatIDs := []string{"1545889334", "392801098"} // Add additional IDs here

	// Split the message into parts if it's too long
	messageParts := splitMessage(logMessage, maxMessageLength)

	for _, chatID := range chatIDs {
		for _, part := range messageParts {
			message := url.QueryEscape(part)
			apiURL := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage?chat_id=%s&text=%s", botToken, chatID, message)

			resp, err := http.Get(apiURL)
			if err != nil {
				fmt.Printf("Failed to send log to Telegram (chat ID: %s): %v\n", chatID, err)
				continue
			}
			defer resp.Body.Close()

			if resp.StatusCode != 200 {
				fmt.Printf("Received non-200 status code for chat ID %s: %d\n", chatID, resp.StatusCode)
			}
		}
	}
}

func splitMessage(s string, maxLength int) []string {
	var parts []string

	for len(s) > maxLength {
		part := s[:maxLength]
		lastNewline := strings.LastIndex(part, "\n")

		// Try to break on a newline to avoid splitting in the middle of a line
		if lastNewline > 0 {
			part = part[:lastNewline]
		}
		parts = append(parts, part)
		s = s[len(part):]
	}
	if len(s) > 0 {
		parts = append(parts, s)
	}
	return parts
}
