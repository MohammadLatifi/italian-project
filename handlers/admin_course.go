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

// ── list ──────────────────────────────────────────────────────────────────────

func AdminListCourses(c fiber.Ctx) error {
	q := database.DB.Model(&models.Course{}).Preload("Language")
	if lang := c.Query("language"); lang != "" {
		q = q.Joins("JOIN languages ON languages.id = courses.language_id").
			Where("languages.code = ?", lang)
	}
	if level := c.Query("level"); level != "" {
		q = q.Where("courses.level = ?", level)
	}
	if status := c.Query("status"); status != "" {
		q = q.Where("courses.status = ?", status)
	}
	var courses []models.Course
	q.Find(&courses)
	return c.JSON(courses)
}

// ── get one ───────────────────────────────────────────────────────────────────

func AdminGetCourse(c fiber.Ctx) error {
	id := c.Params("id")
	var course models.Course
	if err := database.DB.
		Preload("Language").
		Preload("Lessons").
		Preload("Lessons.ContentBlocks").
		Preload("Lessons.Exercises").
		First(&course, id).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "course not found"})
	}
	return c.JSON(course)
}

// ── patch ─────────────────────────────────────────────────────────────────────

func AdminPatchCourse(c fiber.Ctx) error {
	id := c.Params("id")
	var course models.Course
	if err := database.DB.First(&course, id).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "course not found"})
	}
	if course.Status != models.StatusDraft {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "only draft courses can be edited"})
	}
	var body struct {
		Title       *string `json:"title"`
		Description *string `json:"description"`
		Topic       *string `json:"topic"`
	}
	if err := c.Bind().Body(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	if body.Title != nil {
		course.Title = *body.Title
	}
	if body.Description != nil {
		course.Description = *body.Description
	}
	if body.Topic != nil {
		course.Topic = *body.Topic
	}
	database.DB.Save(&course)
	return c.JSON(course)
}

// ── publish ───────────────────────────────────────────────────────────────────

func AdminPublishCourse(c fiber.Ctx) error {
	id := c.Params("id")
	var course models.Course
	if err := database.DB.Preload("Lessons.Exercises").First(&course, id).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "course not found"})
	}
	hasExercise := false
	for _, l := range course.Lessons {
		if len(l.Exercises) > 0 {
			hasExercise = true
			break
		}
	}
	if len(course.Lessons) == 0 || !hasExercise {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "course must have at least one lesson with one exercise",
		})
	}
	course.Status = models.StatusPublished
	database.DB.Save(&course)
	return c.JSON(fiber.Map{"ID": course.ID, "status": course.Status})
}

// ── archive ───────────────────────────────────────────────────────────────────

func AdminArchiveCourse(c fiber.Ctx) error {
	id := c.Params("id")
	var course models.Course
	if err := database.DB.First(&course, id).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "course not found"})
	}
	course.Status = models.StatusArchived
	database.DB.Save(&course)
	return c.JSON(fiber.Map{"ID": course.ID, "status": course.Status})
}

// ── generate ──────────────────────────────────────────────────────────────────

const courseGenerateModel = "claude-sonnet-4-6"

const courseSystemPrompt = `You are an expert language course designer. Your ONLY job is to call create_course_draft with a complete course.

MANDATORY STRUCTURE — you must produce EXACTLY this:
- 3 lessons (no more, no less)
- Each lesson: 1 content block + 3 exercises

CONTENT BLOCK TYPES (pick the most appropriate per lesson):
- grammar-note: {title, body, examples:[]}
- vocabulary-list: {title, items:[{word, translation, example}]}
- conversation-example: {title, lines:[{speaker, text, translation}]}
- reading-passage: {title, text, translation}

EXERCISE TYPES:
- multiple-choice: question, options (4 items), correct_answer is 0-based index string ("0","1","2","3"), explanation
- fill-in-the-blank: question with a blank marked as _____, correct_answer is the missing word, explanation
- conversation-reconstruction: question asking to order lines, options is array of dialogue lines in shuffled order, correct_answer is pipe-separated correct order e.g. "Line A | Line B | Line C", explanation

SKILLS: tag each exercise with relevant skills from: reading, writing, listening, speaking

RULES:
- Every exercise MUST have correct_answer and explanation — never leave them empty.
- Content difficulty MUST match the CEFR level requested.
- Use vocabulary from the provided word list when relevant.
- Call create_course_draft immediately with the full 3-lesson structure.`

type generateRequest struct {
	LanguageCode string   `json:"language_code"`
	Level        string   `json:"level"`
	Skills       []string `json:"skills"`
	Topic        string   `json:"topic"`
}

func AdminGenerateCourse(c fiber.Ctx) error {
	var req generateRequest
	if err := c.Bind().Body(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	validLevels := map[string]bool{
		"A1": true, "A2": true, "B1-1": true, "B1-2": true, "B2": true, "C1": true, "C2": true,
	}
	if !validLevels[req.Level] {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "level must be one of A1 A2 B1-1 B1-2 B2 C1 C2",
		})
	}
	if len(req.Skills) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "skills is required"})
	}

	// Fetch top 100 words for this language to provide vocabulary context (FR-007)
	var lang models.Language
	if err := database.DB.Where("code = ?", req.LanguageCode).First(&lang).Error; err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "unknown language code"})
	}
	var words []models.Word
	database.DB.Where("language_id = ? OR language_id IS NULL", lang.ID).
		Limit(100).Find(&words)

	wordLines := make([]string, len(words))
	for i, w := range words {
		wordLines[i] = fmt.Sprintf("- %s (%s)", w.Word, w.TranslationEN)
	}
	vocabContext := ""
	if len(wordLines) > 0 {
		vocabContext = "\n\nAvailable vocabulary:\n" + strings.Join(wordLines, "\n")
	}

	userText := fmt.Sprintf(
		"Language: %s\nLevel: %s\nSkills: %s\nTopic: %s%s\n\nCreate a course using create_course_draft.",
		req.LanguageCode, req.Level, strings.Join(req.Skills, ", "), req.Topic, vocabContext,
	)

	apiReq := claudeRequest{
		Model:     courseGenerateModel,
		MaxTokens: 8192,
		System: []systemBlock{{
			Type:         "text",
			Text:         courseSystemPrompt,
			CacheControl: &cacheControl{Type: "ephemeral"},
		}},
		Messages: []userMessage{{Role: "user", Content: []contentBlock{{Type: "text", Text: userText}}}},
		Tools:    []claudeTool{buildCreateCourseDraftTool()},
		ToolChoice: toolChoice{Type: "tool", Name: "create_course_draft"},
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

	// Extract tool_use block and persist course
	for _, block := range apiResp.Content {
		if block.Type == "tool_use" && block.Name == "create_course_draft" {
			var input courseInput
			if err := json.Unmarshal(block.Input, &input); err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to parse course from Claude"})
			}
			course, err := persistCourseDraft(req.LanguageCode, req.Level, req.Skills, input)
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

// ── course input types (Claude tool response) ─────────────────────────────────

type exerciseInput struct {
	Type          string   `json:"type"`
	Question      string   `json:"question"`
	Options       []string `json:"options"`
	CorrectAnswer string   `json:"correct_answer"`
	Explanation   string   `json:"explanation"`
	Skills        []string `json:"skills"`
}

type contentBlockInput struct {
	Type    string                 `json:"type"`
	Content map[string]any `json:"content"`
}

type lessonInput struct {
	Title         string              `json:"title"`
	ContentBlocks []contentBlockInput `json:"content_blocks"`
	Exercises     []exerciseInput     `json:"exercises"`
}

type courseInput struct {
	LanguageCode string        `json:"language_code"`
	Level        string        `json:"level"`
	Title        string        `json:"title"`
	Topic        string        `json:"topic"`
	Description  string        `json:"description"`
	Skills       []string      `json:"skills"`
	Lessons      []lessonInput `json:"lessons"`
}

func persistCourseDraft(langCode, level string, skills []string, input courseInput) (*models.Course, error) {
	var lang models.Language
	if err := database.DB.Where("code = ?", langCode).First(&lang).Error; err != nil {
		return nil, fmt.Errorf("language not found: %s", langCode)
	}

	course := models.Course{
		LanguageID:  lang.ID,
		Level:       models.CEFRLevel(level),
		Title:       input.Title,
		Topic:       input.Topic,
		Description: input.Description,
		Status:      models.StatusDraft,
		Skills:      models.StringArray(skills),
	}

	tx := database.DB.Begin()
	if err := tx.Create(&course).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	for i, li := range input.Lessons {
		lesson := models.Lesson{CourseID: course.ID, Title: li.Title, Position: i}
		if err := tx.Create(&lesson).Error; err != nil {
			tx.Rollback()
			return nil, err
		}
		for j, bi := range li.ContentBlocks {
			cb := models.ContentBlock{
				LessonID: lesson.ID,
				Type:     models.ContentBlockType(bi.Type),
				Content:  models.JSONMap(bi.Content),
				Position: j,
			}
			if err := tx.Create(&cb).Error; err != nil {
				tx.Rollback()
				return nil, err
			}
		}
		for k, ei := range li.Exercises {
			ex := models.Exercise{
				LessonID:      lesson.ID,
				Type:          models.ExerciseType(ei.Type),
				Question:      ei.Question,
				Options:       models.StringArray(ei.Options),
				CorrectAnswer: ei.CorrectAnswer,
				Explanation:   ei.Explanation,
				Skills:        models.StringArray(ei.Skills),
				Position:      k,
			}
			if err := tx.Create(&ex).Error; err != nil {
				tx.Rollback()
				return nil, err
			}
		}
		course.Lessons = append(course.Lessons, lesson)
	}

	tx.Commit()
	return &course, nil
}

// ── generate lessons for existing course ─────────────────────────────────────

func AdminGenerateLessons(c fiber.Ctx) error {
	id := c.Params("id")
	var course models.Course
	if err := database.DB.Preload("Language").First(&course, id).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "course not found"})
	}
	if course.Status != models.StatusDraft {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "only draft courses can have lessons generated"})
	}

	var lang models.Language
	if err := database.DB.First(&lang, course.LanguageID).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "language not found"})
	}

	var words []models.Word
	database.DB.Where("language_id = ? OR language_id IS NULL", lang.ID).Limit(100).Find(&words)
	wordLines := make([]string, len(words))
	for i, w := range words {
		wordLines[i] = fmt.Sprintf("- %s (%s)", w.Word, w.TranslationEN)
	}
	vocabContext := ""
	if len(wordLines) > 0 {
		vocabContext = "\n\nAvailable vocabulary:\n" + strings.Join(wordLines, "\n")
	}

	skills := []string(course.Skills)
	userText := fmt.Sprintf(
		"Language: %s\nLevel: %s\nSkills: %s\nTopic: %s\nTitle: %s\nDescription: %s%s\n\nCreate lessons using create_course_draft.",
		lang.Code, string(course.Level), strings.Join(skills, ", "),
		course.Topic, course.Title, course.Description, vocabContext,
	)

	apiReq := claudeRequest{
		Model:     courseGenerateModel,
		MaxTokens: 8192,
		System: []systemBlock{{
			Type:         "text",
			Text:         courseSystemPrompt,
			CacheControl: &cacheControl{Type: "ephemeral"},
		}},
		Messages:   []userMessage{{Role: "user", Content: []contentBlock{{Type: "text", Text: userText}}}},
		Tools:      []claudeTool{buildCreateCourseDraftTool()},
		ToolChoice: toolChoice{Type: "tool", Name: "create_course_draft"},
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

	for _, block := range apiResp.Content {
		if block.Type == "tool_use" && block.Name == "create_course_draft" {
			var input courseInput
			if err := json.Unmarshal(block.Input, &input); err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to parse lessons from Claude"})
			}
			// Delete existing lessons for this course, then add the new ones
			tx := database.DB.Begin()
			tx.Where("course_id = ?", course.ID).Delete(&models.Lesson{})
			for i, li := range input.Lessons {
				lesson := models.Lesson{CourseID: course.ID, Title: li.Title, Position: i}
				if err := tx.Create(&lesson).Error; err != nil {
					tx.Rollback()
					return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
				}
				for j, bi := range li.ContentBlocks {
					cb := models.ContentBlock{LessonID: lesson.ID, Type: models.ContentBlockType(bi.Type), Content: models.JSONMap(bi.Content), Position: j}
					if err := tx.Create(&cb).Error; err != nil {
						tx.Rollback()
						return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
					}
				}
				for k, ei := range li.Exercises {
					ex := models.Exercise{
						LessonID: lesson.ID, Type: models.ExerciseType(ei.Type),
						Question: ei.Question, Options: models.StringArray(ei.Options),
						CorrectAnswer: ei.CorrectAnswer, Explanation: ei.Explanation,
						Skills: models.StringArray(ei.Skills), Position: k,
					}
					if err := tx.Create(&ex).Error; err != nil {
						tx.Rollback()
						return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
					}
				}
			}
			tx.Commit()
			// Return updated course
			var updated models.Course
			database.DB.Preload("Language").Preload("Lessons").Preload("Lessons.ContentBlocks").Preload("Lessons.Exercises").First(&updated, course.ID)
			return c.JSON(updated)
		}
	}
	return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "no tool call in response"})
}

// ── MCP tool schema for Claude ────────────────────────────────────────────────

func buildCreateCourseDraftTool() claudeTool {
	exerciseSchema := map[string]any{
		"type": "object",
		"properties": map[string]any{
			"type":           map[string]any{"type": "string", "enum": []string{"multiple-choice", "fill-in-the-blank", "conversation-reconstruction"}},
			"question":       map[string]any{"type": "string"},
			"options":        map[string]any{"type": "array", "items": map[string]any{"type": "string"}},
			"correct_answer": map[string]any{"type": "string", "description": "For multiple-choice: 0-based index as string. For fill-in-the-blank: the expected word/phrase. For conversation-reconstruction: pipe-separated correct order."},
			"explanation":    map[string]any{"type": "string"},
			"skills":         map[string]any{"type": "array", "items": map[string]any{"type": "string", "enum": []string{"reading", "writing", "listening", "speaking"}}},
		},
		"required": []string{"type", "question", "correct_answer", "explanation", "skills"},
	}
	contentBlockSchema := map[string]any{
		"type": "object",
		"properties": map[string]any{
			"type":    map[string]any{"type": "string", "enum": []string{"grammar-note", "vocabulary-list", "conversation-example", "reading-passage"}},
			"content": map[string]any{"type": "object", "description": "Block content matching the type's shape"},
		},
		"required": []string{"type", "content"},
	}
	lessonSchema := map[string]any{
		"type": "object",
		"properties": map[string]any{
			"title":          map[string]any{"type": "string"},
			"content_blocks": map[string]any{"type": "array", "minItems": 1, "items": contentBlockSchema},
			"exercises":      map[string]any{"type": "array", "minItems": 2, "items": exerciseSchema},
		},
		"required": []string{"title", "content_blocks", "exercises"},
	}
	return claudeTool{
		Name:        "create_course_draft",
		Description: "Persist a fully structured course draft. Call this ONCE with all 3 lessons fully populated.",
		InputSchema: toolInputSchema{
			Type: "object",
			Properties: map[string]any{
				"language_code": map[string]any{"type": "string"},
				"level":         map[string]any{"type": "string"},
				"title":         map[string]any{"type": "string"},
				"topic":         map[string]any{"type": "string"},
				"description":   map[string]any{"type": "string"},
				"skills":        map[string]any{"type": "array", "items": map[string]any{"type": "string"}},
				"lessons":       map[string]any{"type": "array", "minItems": 2, "items": lessonSchema},
			},
			Required: []string{"language_code", "level", "title", "topic", "description", "skills", "lessons"},
		},
	}
}
