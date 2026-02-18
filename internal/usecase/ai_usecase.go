// internal/usecase/ai_usecase.go (UPDATED)

package usecase

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/hensybex/soulwi_go_back/internal/model"
	"github.com/hensybex/soulwi_go_back/internal/repository"
	"github.com/hensybex/soulwi_go_back/internal/service"
	"github.com/hensybex/soulwi_go_back/internal/utils"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// SSE data format from OpenAI
type openAIChatChunk struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Index        int    `json:"index"`
		FinishReason string `json:"finish_reason"`
		Delta        struct {
			Content string `json:"content,omitempty"`
			Role    string `json:"role,omitempty"`
		} `json:"delta"`
	} `json:"choices"`
}

type AIUsecase interface {
	StartChatStream(ctx context.Context, chatID uint, userUID, newUserMessage string, parentMessageID *uint) (<-chan string, <-chan error)
	GenerateWisePhrases(ctx context.Context, promptID uint, count int, existing []string) ([]string, error)
	GenerateChatName(ctx context.Context, userMessage, aiReply string) (string, error)
}

type aiUsecase struct {
	chatRepo       repository.ChatRepository
	messageRepo    repository.MessageRepository
	promptRepo     repository.PromptRepository
	basePromptRepo repository.BasePromptRepository
	openAiKey      string
}

func NewAIUsecase(
	chatRepo repository.ChatRepository,
	messageRepo repository.MessageRepository,
	promptRepo repository.PromptRepository,
	basePromptRepo repository.BasePromptRepository,
	openAiKey string,
) AIUsecase {
	return &aiUsecase{
		chatRepo:       chatRepo,
		messageRepo:    messageRepo,
		promptRepo:     promptRepo,
		basePromptRepo: basePromptRepo,
		openAiKey:      openAiKey,
	}
}

func (u *aiUsecase) StartChatStream(ctx context.Context, chatID uint, userUID, newUserMessage string, parentMessageID *uint) (<-chan string, <-chan error) {
	outChan := make(chan string)
	errChan := make(chan error, 1)

	go func() {
		defer close(outChan)
		defer close(errChan)

		// Validate chat ownership
		chat, err := u.chatRepo.GetByID(ctx, chatID, userUID)
		if err != nil || chat == nil {
			errChan <- fmt.Errorf("chat not found or not owned by user")
			return
		}

		// The handler (sse.go) is now solely responsible for creating the user message.

		// Fetch system prompt
		if chat.PromptID == nil {
			errChan <- fmt.Errorf("chat.PromptID is nil")
			return
		}

		prompt, err := u.promptRepo.GetByID(ctx, *chat.PromptID)
		if err != nil || prompt == nil {
			errChan <- fmt.Errorf("failed to fetch system prompt")
			return
		}

		// Possibly fetch subgroup base prompt
		var basePrompt *model.BasePrompt
		if prompt.SubGroupID != nil {
			subGroup, _ := u.promptRepo.GetSubGroupByID(ctx, *prompt.SubGroupID)
			var basePromptID uint
			if subGroup.BasePromptID == nil {
				basePromptID = 1
			} else {
				basePromptID = *subGroup.BasePromptID
			}

			basePrompt, err = u.basePromptRepo.GetBasePromptByID(ctx, basePromptID)
			if err != nil {
				errChan <- fmt.Errorf("failed to fetch base prompt")
				return
			}
		} else {
			basePrompt, _ = u.basePromptRepo.GetBasePromptByID(ctx, 1)
		}

		// Extract temperature, max_tokens, and model
		temperature := prompt.Temperature
		maxTokens := prompt.MaxTokens
		apiModel, err := u.mapModelName(prompt.ModelName)
		if err != nil {
			errChan <- err
			return
		}

		// Fetch prior messages
		var basePromptContent string
		if basePrompt != nil {
			basePromptContent = basePrompt.Prompt
		}

		// --- UPDATE ---
		// Decide which prompt content to use. For the main chat (PromptID 1),
		// we only want the base system prompt, not its own specific content.
		finalPromptContent := prompt.Content
		if prompt.ID == 1 {
			finalPromptContent = ""
		}

		// Pass the potentially empty finalPromptContent to the builder.
		messages, err := u.prepareChatMessages(ctx, chatID, basePromptContent, finalPromptContent)
		if err != nil {
			errChan <- err
			return
		}

		// Call OpenAI API for streaming
		err = u.streamOpenAI(ctx, messages, apiModel, temperature, maxTokens, outChan, chatID, parentMessageID)
		if err != nil {
			errChan <- err
		}
	}()

	return outChan, errChan
}

/* func (u *aiUsecase) mapModelName(model string) (string, error) {
	switch model {
	case "GPT-4o-mini":
		return "gpt-4o-mini", nil
	case "GPT-4o":
		return "gpt-4o", nil
	case "o3-mini-2025-01-31":
		return "o3-mini-2025-01-31", nil
	default:
		return "", fmt.Errorf("unknown model: %s", model)
	}
} */

func (u *aiUsecase) mapModelName(model string) (string, error) {
	// Temporary override to always use specific model
	return "gpt-4.1-nano-2025-04-14", nil
}

func (u *aiUsecase) prepareChatMessages(ctx context.Context, chatID uint, basePrompt, promptContent string) ([]map[string]string, error) {
	messages := make([]map[string]string, 0)

	// 1. Base prompt as system message
	if basePrompt != "" {
		messages = append(messages, map[string]string{
			"role":    "system",
			"content": basePrompt,
		})
	}

	// --- UPDATE ---
	// 2. Prompt's content as an assistant message, ONLY IF it's not empty.
	if promptContent != "" {
		messages = append(messages, map[string]string{
			"role":    "assistant",
			"content": promptContent,
		})
	}

	// 3. Include prior conversation messages
	oldMessages, err := u.messageRepo.ListActiveByChat(ctx, chatID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch chat messages: %w", err)
	}
	for _, m := range oldMessages {
		messages = append(messages, map[string]string{
			"role":    m.Role,
			"content": m.Content,
		})
	}

	// Convert messages to formatted log output
	formattedLog := formatTelegramLog(messages)

	log.Println("Generated AI Prompt:\n", formattedLog)
	// Send log to Telegram
	go service.SendLogToTelegram(formattedLog)

	// Trim messages if necessary, using the "Golden Mean" budget
	return u.trimMessagesToFitTokens(messages, 8192)
}

func formatTelegramLog(messages []map[string]string) string {
	var logBuilder strings.Builder
	firstAssistantLogged := false // Track if first assistant message has been logged as "ACTUAL PROMPT"

	for _, msg := range messages {
		role := msg["role"]
		content := msg["content"]

		switch role {
		case "system":
			logBuilder.WriteString(fmt.Sprintf("——SYSTEM PROMPT——\n%s\n\n", content))
		case "assistant":
			if !firstAssistantLogged {
				logBuilder.WriteString(fmt.Sprintf("——ACTUAL PROMPT——\n%s\n\n", content))
				firstAssistantLogged = true
			} else {
				logBuilder.WriteString(fmt.Sprintf("🤖 Assistant: %s\n", content))
			}
		case "user":
			logBuilder.WriteString(fmt.Sprintf("👤 User: %s\n", content))
		default:
			// Catch any unexpected roles, just in case
			logBuilder.WriteString(fmt.Sprintf("🔹 %s: %s\n", cases.Title(language.English).String(role), content))
		}
	}

	return logBuilder.String()
}

func (u *aiUsecase) trimMessagesToFitTokens(messages []map[string]string, maxTokens int) ([]map[string]string, error) {
	countTokens := func(msgs []map[string]string) (int, error) {
		return utils.CountTokensTiktoken(msgs, "o200k_base")
	}

	for {
		tokenCount, err := countTokens(messages)
		if err != nil {
			return nil, err
		}
		if tokenCount <= maxTokens {
			break // Контекст в норме, выходим
		}

		// Мы должны сохранить [system, assistant] промпты. Их 2.
		// Если сообщений 3 или меньше, удалять уже нечего (кроме самой истории).
		if len(messages) <= 3 { // <-- ИЗМЕНЕНО
			break
		}

		// Сохраняем первые два элемента (system, assistant) и удаляем третий (самое старое сообщение чата)
		messages = append(messages[:2], messages[3:]...) // <-- ГЛАВНОЕ ИЗМЕНЕНИЕ
	}

	return messages, nil
}

func (u *aiUsecase) streamOpenAI(
	ctx context.Context,
	messages []map[string]string,
	llmModel string,
	temperature float64,
	maxTokens int,
	outChan chan<- string,
	chatID uint,
	parentMessageID *uint,
) error {
	reqBody := map[string]interface{}{
		"model":                 llmModel,
		"stream":                true,
		"messages":              messages,
		"max_completion_tokens": maxTokens,
	}

	jsonBody, _ := json.Marshal(reqBody)

	req, _ := http.NewRequestWithContext(ctx, "POST", "https://api.openai.com/v1/chat/completions", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+u.openAiKey)

	resp, err := (&http.Client{}).Do(req)
	if err != nil {
		return fmt.Errorf("openai call failed: %w", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Printf("Failed to close response body: %v", err)
		}
	}()
	if resp.StatusCode != http.StatusOK {
		data, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("openai error: %s", string(data))
	}

	reader := bufio.NewReader(resp.Body)
	for {
		line, err := reader.ReadString('\n')
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("error reading SSE: %w", err)
		}

		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "data: ") {
			data := strings.TrimPrefix(line, "data: ")
			if data == "[DONE]" {
				break
			}

			var chunk openAIChatChunk
			if err := json.Unmarshal([]byte(data), &chunk); err == nil && len(chunk.Choices) > 0 {
				content := chunk.Choices[0].Delta.Content
				outChan <- content
			}
		}
	}

	// **FIX: REMOVED aI response saving from the use case.**
	// The handler is now responsible for assembling the final response and saving it.
	// This prevents race conditions and makes the logic more robust.

	return nil
}

func (u *aiUsecase) GenerateWisePhrases(ctx context.Context, promptID uint, count int, existing []string) ([]string, error) {
	// 1. Fetch the prompt from DB
	prompt, err := u.promptRepo.GetByID(ctx, promptID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch wise-words prompt: %w", err)
	}
	if prompt == nil {
		return nil, fmt.Errorf("prompt with ID %d not found", promptID)
	}

	// 2. Build the chat messages for OpenAI request.
	modelName := prompt.ModelName
	if strings.TrimSpace(modelName) == "" {
		modelName = "gpt-4.1-nano"
	}

	baseInstruction := fmt.Sprintf(
		"%s\n\nGenerate exactly %d short, unique wise phrases in Russian or English depending on the original prompt. Output each phrase on a separate line with no numbering. Avoid repetition and maintain brevity (max 12 words).",
		prompt.Content, count,
	)

	messages := []map[string]string{
		{
			"role":    "system",
			"content": baseInstruction,
		},
	}

	if len(existing) > 0 {
		var builder strings.Builder
		builder.Grow(len(existing) * 16)
		for _, phrase := range existing {
			trimmed := strings.TrimSpace(phrase)
			if trimmed == "" {
				continue
			}
			builder.WriteString(trimmed)
			builder.WriteString("\n")
		}
		existingBlock := builder.String()
		if existingBlock != "" {
			messages = append(messages, map[string]string{
				"role": "user",
				"content": fmt.Sprintf(
					"Список существующих мудрых фраз (не повторяй их и не создавай близких по смыслу дубликатов):\n%s", existingBlock,
				),
			})
		}
	} else {
		messages = append(messages, map[string]string{
			"role":    "user",
			"content": "Создай новых уникальных мудрых фраз, избегая повторов.",
		})
	}

	// 3. Prepare the request body for a single Chat Completion
	reqBody := map[string]interface{}{
		"model":                 modelName,
		"stream":                false,
		"max_completion_tokens": prompt.MaxTokens,
		"messages":              messages,
	}

	// 4. Marshal and call OpenAI
	jsonBody, _ := json.Marshal(reqBody)
	req, _ := http.NewRequestWithContext(ctx, "POST", "https://api.openai.com/v1/chat/completions", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+u.openAiKey)

	resp, err := (&http.Client{}).Do(req)
	if err != nil {
		return nil, fmt.Errorf("OpenAI call failed: %w", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Printf("Failed to close response body: %v", err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		data, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("OpenAI error %d: %s", resp.StatusCode, string(data))
	}

	// 5. Parse OpenAI’s response (should be a single chunk)
	var response struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode OpenAI response: %w", err)
	}
	if len(response.Choices) == 0 {
		return nil, fmt.Errorf("no choices returned from OpenAI")
	}

	// 6. The GPT output text
	rawText := strings.TrimSpace(response.Choices[0].Message.Content)
	if rawText == "" {
		return nil, fmt.Errorf("OpenAI returned empty wise phrases")
	}

	// 7. Split text into lines
	lines := strings.Split(rawText, "\n")
	// Trim each line, remove empties
	var phrases []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			phrases = append(phrases, line)
		}
	}

	return phrases, nil
}

// GenerateChatName generates a descriptive name for a chat based on the initial messages.
func (u *aiUsecase) GenerateChatName(ctx context.Context, userMessage, aiReply string) (string, error) {
	var prompt string
	if userMessage != "" {
		prompt = fmt.Sprintf("Craft a name for the given chat of user in my app, related to mental health. First messages:\nUser message: %s\nAI reply: %s", userMessage, aiReply)
	} else {
		prompt = fmt.Sprintf("Craft a name for the given chat of user in my app, related to mental health. First message:\nAI greeting: %s", aiReply)
	}

	// Use a specific model for this task, as per requirements
	modelName := "gpt-4.1-nano"

	reqBody := map[string]interface{}{
		"model":  modelName,
		"stream": false,
		"messages": []map[string]string{
			{"role": "system", "content": "You are an expert at creating concise, descriptive, and emotionally resonant chat titles. The title should be very short, ideally 2-4 words. Do not use quotes."},
			{"role": "user", "content": prompt},
		},
		"max_completion_tokens": 20, // A short limit for a title
		"temperature":           0.7,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request body: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.openai.com/v1/chat/completions", bytes.NewReader(jsonBody))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+u.openAiKey)

	resp, err := (&http.Client{}).Do(req)
	if err != nil {
		return "", fmt.Errorf("openai call failed: %w", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Printf("Failed to close response body in GenerateChatName: %v", err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		data, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("openai error %d: %s", resp.StatusCode, string(data))
	}

	var response struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return "", fmt.Errorf("failed to decode openai response: %w", err)
	}
	if len(response.Choices) == 0 || response.Choices[0].Message.Content == "" {
		return "", fmt.Errorf("no chat name returned from openai")
	}

	// Clean up the response: remove quotes and trim whitespace
	chatName := strings.Trim(response.Choices[0].Message.Content, "\" ")
	return chatName, nil
}
