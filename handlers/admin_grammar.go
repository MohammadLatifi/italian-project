package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"os"
	"strings"

	"italian_project/database"
	"italian_project/models"

	"github.com/gofiber/fiber/v3"
)

const grammarSystemPrompt = `You are a language curriculum analyst.
Given a list of existing course topics for a specific language and CEFR level,
identify the grammar topics already covered and suggest at least 5 important
grammar topics that are missing at that level.
Use the suggest_grammar_topics tool to return your findings.`

// AdminListGrammarSuggestions returns existing pending suggestions — no AI call.
func AdminListGrammarSuggestions(c fiber.Ctx) error {
	langCode := c.Query("language")
	level := c.Query("level")
	if langCode == "" || level == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "language and level are required"})
	}
	var lang models.Language
	if err := database.DB.Where("code = ?", langCode).First(&lang).Error; err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "unknown language"})
	}
	var suggestions []models.GrammarSuggestion
	database.DB.Where("language_id = ? AND level = ? AND status = ?", lang.ID, level, models.SuggestionPending).
		Find(&suggestions)
	return c.JSON(suggestions)
}

// AdminAnalyzeGrammarSuggestions triggers a fresh AI analysis and returns new pending suggestions.
func AdminAnalyzeGrammarSuggestions(c fiber.Ctx) error {
	langCode := c.Query("language")
	level := c.Query("level")
	if langCode == "" || level == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "language and level are required"})
	}
	var lang models.Language
	if err := database.DB.Where("code = ?", langCode).First(&lang).Error; err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "unknown language"})
	}
	suggestions := analyzeGrammarCoverage(lang, models.CEFRLevel(level))
	if suggestions == nil {
		return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{"error": "AI analysis failed"})
	}
	return c.JSON(suggestions)
}

// analyzeGrammarCoverage calls Claude to identify missing grammar topics and
// persists them as GrammarSuggestion rows. This function is shared by both
// the admin HTTP handler and the MCP tool (Principle II compliance).
func analyzeGrammarCoverage(lang models.Language, level models.CEFRLevel) []models.GrammarSuggestion {
	var courses []models.Course
	database.DB.Where("language_id = ? AND level = ? AND status = ?", lang.ID, level, models.StatusPublished).
		Find(&courses)

	topics := make([]string, 0, len(courses))
	for _, course := range courses {
		topics = append(topics, course.Topic)
	}

	userText := "Language: " + lang.Code + "\nLevel: " + string(level) +
		"\nExisting course topics:\n" + strings.Join(topics, "\n") +
		"\n\nUse the suggest_grammar_topics tool to return covered and missing grammar topics."

	suggestTool := claudeTool{
		Name:        "suggest_grammar_topics",
		Description: "Returns covered and missing grammar topics for a language/level",
		InputSchema: toolInputSchema{
			Type: "object",
			Properties: map[string]any{
				"covered_topics": map[string]any{
					"type":  "array",
					"items": map[string]any{"type": "string"},
				},
				"suggested_missing": map[string]any{
					"type": "array",
					"items": map[string]any{
						"type": "object",
						"properties": map[string]any{
							"topic":     map[string]any{"type": "string"},
							"rationale": map[string]any{"type": "string"},
						},
					},
				},
			},
			Required: []string{"covered_topics", "suggested_missing"},
		},
	}

	apiReq := claudeRequest{
		Model:     courseGenerateModel,
		MaxTokens: 1024,
		System: []systemBlock{{
			Type: "text",
			Text: grammarSystemPrompt,
		}},
		Messages:   []userMessage{{Role: "user", Content: []contentBlock{{Type: "text", Text: userText}}}},
		Tools:      []claudeTool{suggestTool},
		ToolChoice: toolChoice{Type: "tool", Name: "suggest_grammar_topics"},
	}

	body, _ := json.Marshal(apiReq)
	httpReq, err := http.NewRequestWithContext(context.Background(), http.MethodPost, anthropicAPIURL, bytes.NewReader(body))
	if err != nil {
		return nil
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-api-key", os.Getenv("ANTHROPIC_API_KEY"))
	httpReq.Header.Set("anthropic-version", "2023-06-01")

	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return nil
	}
	defer resp.Body.Close()

	var apiResp claudeResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil
	}

	var suggestions []models.GrammarSuggestion
	for _, block := range apiResp.Content {
		if block.Type != "tool_use" || block.Name != "suggest_grammar_topics" {
			continue
		}
		var result struct {
			SuggestedMissing []struct {
				Topic     string `json:"topic"`
				Rationale string `json:"rationale"`
			} `json:"suggested_missing"`
		}
		if err := json.Unmarshal(block.Input, &result); err != nil {
			break
		}
		for _, s := range result.SuggestedMissing {
			suggestion := models.GrammarSuggestion{
				LanguageID: lang.ID,
				Level:      level,
				Topic:      s.Topic,
				Rationale:  s.Rationale,
				Status:     models.SuggestionPending,
			}
			database.DB.Create(&suggestion)
			suggestions = append(suggestions, suggestion)
		}
		break
	}
	return suggestions
}

func AdminAcceptGrammarSuggestion(c fiber.Ctx) error {
	id := c.Params("id")
	var suggestion models.GrammarSuggestion
	if err := database.DB.Preload("Language").First(&suggestion, id).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "suggestion not found"})
	}
	suggestion.Status = models.SuggestionAccepted
	database.DB.Save(&suggestion)

	// Reuse the generate logic
	var lang models.Language
	database.DB.First(&lang, suggestion.LanguageID)
	var words []models.Word
	database.DB.Where("language_id = ? OR language_id IS NULL", lang.ID).Limit(100).Find(&words)

	wordLines := make([]string, len(words))
	for i, w := range words {
		wordLines[i] = "- " + w.Word + " (" + w.TranslationEN + ")"
	}
	vocabContext := ""
	if len(wordLines) > 0 {
		vocabContext = "\n\nAvailable vocabulary:\n" + strings.Join(wordLines, "\n")
	}

	userText := "Language: " + lang.Code + "\nLevel: " + string(suggestion.Level) +
		"\nSkills: reading, writing, listening, speaking" +
		"\nTopic: " + suggestion.Topic + vocabContext +
		"\n\nCreate a course using create_course_draft."

	apiReq := claudeRequest{
		Model:     courseGenerateModel,
		MaxTokens: 4096,
		System: []systemBlock{{
			Type: "text",
			Text: courseSystemPrompt,
		}},
		Messages:   []userMessage{{Role: "user", Content: []contentBlock{{Type: "text", Text: userText}}}},
		Tools:      []claudeTool{buildCreateCourseDraftTool()},
		ToolChoice: toolChoice{Type: "tool", Name: "create_course_draft"},
	}

	body, _ := json.Marshal(apiReq)
	httpReq, err := http.NewRequestWithContext(context.Background(), http.MethodPost, anthropicAPIURL, bytes.NewReader(body))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "request build failed"})
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-api-key", os.Getenv("ANTHROPIC_API_KEY"))
	httpReq.Header.Set("anthropic-version", "2023-06-01")

	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Claude API call failed"})
	}
	defer resp.Body.Close()

	var apiResp claudeResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "decode failed"})
	}
	if apiResp.Error != nil {
		return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{"error": apiResp.Error.Message})
	}

	for _, block := range apiResp.Content {
		if block.Type == "tool_use" && block.Name == "create_course_draft" {
			var input courseInput
			if err := json.Unmarshal(block.Input, &input); err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "parse failed"})
			}
			skills := []string{"reading", "writing", "listening", "speaking"}
			course, err := persistCourseDraft(lang.Code, string(suggestion.Level), skills, input)
			if err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
			}
			return c.Status(fiber.StatusCreated).JSON(fiber.Map{
				"ID":           course.ID,
				"status":       course.Status,
				"title":        course.Title,
				"lesson_count": len(course.Lessons),
			})
		}
	}
	return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "no tool call in response"})
}

func AdminDismissGrammarSuggestion(c fiber.Ctx) error {
	id := c.Params("id")
	var suggestion models.GrammarSuggestion
	if err := database.DB.First(&suggestion, id).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "suggestion not found"})
	}
	database.DB.Unscoped().Delete(&suggestion)
	return c.SendStatus(fiber.StatusNoContent)
}
