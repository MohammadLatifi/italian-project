package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	"italian_project/database"
	"italian_project/models"

	"github.com/gofiber/fiber/v3"
)

const anthropicAPIURL = "https://api.anthropic.com/v1/messages"
const generateModel = "claude-sonnet-4-6"

// systemPrompt is static — marked for prompt caching so repeated calls reuse it.
const systemPrompt = `You are an Italian language teacher creating vocabulary practice paragraphs for learners.
Write 2-4 natural Italian sentences at B1-B2 level that incorporate ALL the provided vocabulary words.
The paragraph must feel like real, authentic Italian — not a forced exercise.
Every word in the list must appear in the paragraph (conjugated or inflected as appropriate).
You must return your answer exclusively via the create_paragraph tool.`

// ── Anthropic API wire types ──────────────────────────────────────────────────

type cacheControl struct {
	Type string `json:"type"`
}

type systemBlock struct {
	Type         string        `json:"type"`
	Text         string        `json:"text"`
	CacheControl *cacheControl `json:"cache_control,omitempty"`
}

type contentBlock struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type userMessage struct {
	Role    string         `json:"role"`
	Content []contentBlock `json:"content"`
}

type toolInputSchema struct {
	Type       string         `json:"type"`
	Properties map[string]any `json:"properties"`
	Required   []string       `json:"required"`
}

type claudeTool struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	InputSchema toolInputSchema `json:"input_schema"`
}

type toolChoice struct {
	Type string `json:"type"`
	Name string `json:"name"`
}

type claudeRequest struct {
	Model      string        `json:"model"`
	MaxTokens  int           `json:"max_tokens"`
	System     []systemBlock `json:"system"`
	Messages   []userMessage `json:"messages"`
	Tools      []claudeTool  `json:"tools"`
	ToolChoice toolChoice    `json:"tool_choice"`
}

type responseBlock struct {
	Type  string          `json:"type"`
	Name  string          `json:"name,omitempty"`
	Input json.RawMessage `json:"input,omitempty"`
}

type claudeResponse struct {
	Content []responseBlock `json:"content"`
	Error   *struct {
		Type    string `json:"type"`
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

// paragraphToolInput is the exact shape Claude must return via the tool.
// related_word_ids is populated here but we override it with DB-verified IDs.
type paragraphToolInput struct {
	Paragraph     string `json:"paragraph"`
	TranslationEN string `json:"translation_en"`
}

// ── Handler ───────────────────────────────────────────────────────────────────

type GenerateRequest struct {
	WordIDs []uint `json:"word_ids"`
	Context string `json:"context"`
}

var validContexts = map[string]bool{
	"general": true, "sport": true, "art": true, "software": true,
	"finance": true, "law": true, "medical": true,
	"science": true, "family": true, "nature": true,
}

// createParagraphTool returns the tool definition. The schema matches the DB
// schema exactly; context is an enum so Claude cannot invent new values.
func createParagraphTool() claudeTool {
	return claudeTool{
		Name:        "create_paragraph",
		Description: "Returns a structured Italian learning paragraph that matches the database schema exactly.",
		InputSchema: toolInputSchema{
			Type: "object",
			Properties: map[string]any{
				"paragraph": map[string]any{
					"type":        "string",
					"description": "The Italian paragraph text (2–6 sentences)",
				},
				"translation_en": map[string]any{
					"type":        "string",
					"description": "Accurate English translation of the paragraph",
				},
				"context": map[string]any{
					"type": "string",
					"enum": []string{
						"general", "sport", "art", "software", "finance",
						"law", "medical", "science", "family", "nature",
					},
				},
				"related_word_ids": map[string]any{
					"type":        "array",
					"items":       map[string]any{"type": "integer"},
					"description": "IDs of the vocabulary words that appear in the paragraph",
				},
			},
			Required: []string{"paragraph", "translation_en", "context", "related_word_ids"},
		},
	}
}

func GenerateParagraph(c fiber.Ctx) error {
	var req GenerateRequest
	if err := c.Bind().Body(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	if len(req.WordIDs) < 3 || len(req.WordIDs) > 7 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "word_ids must contain 3–7 IDs"})
	}
	if !validContexts[req.Context] {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid context"})
	}

	// Load words from DB — this validates IDs and gives us the real Italian/English text.
	var words []models.Word
	if err := database.DB.Where("id IN ?", req.WordIDs).Find(&words).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "db error"})
	}
	if len(words) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "no valid word IDs found"})
	}

	// Build the user prompt.
	lines := make([]string, len(words))
	for i, w := range words {
		lines[i] = fmt.Sprintf("- %s (%s)", w.Word, w.TranslationEN)
	}
	userText := fmt.Sprintf(
		"Context: %s\n\nVocabulary words to include:\n%s\n\nUse the create_paragraph tool to return your answer.",
		req.Context, strings.Join(lines, "\n"),
	)

	apiReq := claudeRequest{
		Model:     generateModel,
		MaxTokens: 1024,
		// System prompt is always identical → ideal cache target.
		System: []systemBlock{{
			Type:         "text",
			Text:         systemPrompt,
			CacheControl: &cacheControl{Type: "ephemeral"},
		}},
		Messages:   []userMessage{{Role: "user", Content: []contentBlock{{Type: "text", Text: userText}}}},
		Tools:      []claudeTool{createParagraphTool()},
		// tool_choice forces Claude to always call the tool — no free-text fallback.
		ToolChoice: toolChoice{Type: "tool", Name: "create_paragraph"},
	}

	body, _ := json.Marshal(apiReq)

	httpReq, err := http.NewRequestWithContext(context.Background(), http.MethodPost, anthropicAPIURL, bytes.NewReader(body))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to build request"})
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-api-key", os.Getenv("ANTHROPIC_API_KEY"))
	httpReq.Header.Set("anthropic-version", "2023-06-01")
	httpReq.Header.Set("anthropic-beta", "prompt-caching-2024-07-31")

	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Claude API call failed"})
	}
	defer resp.Body.Close()

	var apiResp claudeResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to decode Claude response"})
	}
	if apiResp.Error != nil {
		return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{"error": apiResp.Error.Message})
	}

	// Extract the tool_use block — guaranteed to exist because of tool_choice.
	var toolInput paragraphToolInput
	found := false
	for _, block := range apiResp.Content {
		if block.Type == "tool_use" && block.Name == "create_paragraph" {
			if err := json.Unmarshal(block.Input, &toolInput); err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "tool response parse failed"})
			}
			found = true
			break
		}
	}
	if !found {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "no tool call in response"})
	}

	// Use DB-verified IDs — never trust Claude's word IDs to prevent hallucination.
	relatedIDs := make(models.IntArray, len(words))
	for i, w := range words {
		relatedIDs[i] = w.ID
	}

	paragraph := models.Paragraph{
		Paragraph:      toolInput.Paragraph,
		TranslationEN:  toolInput.TranslationEN,
		Context:        models.ParagraphContext(req.Context),
		RelatedWordIDs: relatedIDs,
	}

	if err := database.DB.Create(&paragraph).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to save paragraph"})
	}

	return c.Status(fiber.StatusCreated).JSON(paragraph)
}
