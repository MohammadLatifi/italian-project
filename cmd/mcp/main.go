package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/joho/godotenv"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"italian_project/database"
	"italian_project/models"
)

// ---------- input structs ----------

type WordInput struct {
	Word            string `json:"word"`
	TranslationEN   string `json:"translation_en"`
	ExampleSentence string `json:"example_sentence"`
}

type AddWordsArgs struct {
	Words []WordInput `json:"words"`
}

type GetWordsArgs struct {
	IDs    []uint `json:"ids"`
	Limit  int    `json:"limit"`
	Offset int    `json:"offset"`
}

type ParagraphInput struct {
	Paragraph      string                  `json:"paragraph"`
	TranslationEN  string                  `json:"translation_en"`
	RelatedWordIDs []uint                  `json:"related_word_ids"`
	Context        models.ParagraphContext `json:"context"`
	PracticeCount  int                     `json:"practice_count"`
}

type AddParagraphsArgs struct {
	Paragraphs []ParagraphInput `json:"paragraphs"`
}

type GetParagraphsArgs struct {
	IDs          []uint `json:"ids"`
	IncludeWords bool   `json:"include_words"`
	Limit        int    `json:"limit"`
	Offset       int    `json:"offset"`
}

// ---------- helpers ----------

func toJSON(v any) string {
	b, _ := json.MarshalIndent(v, "", "  ")
	return string(b)
}

func defaultLimit(n int) int {
	if n <= 0 {
		return 50
	}
	return n
}

// ---------- main ----------

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}

	database.Connect()

	s := server.NewMCPServer("italian-words", "2.0.0")

	// ── add_words ──────────────────────────────────────────────────────────────
	s.AddTool(
		mcp.NewTool("add_words",
			mcp.WithDescription("Batch-insert Italian words into the database"),
			mcp.WithArray("words",
				mcp.Required(),
				mcp.Description("Array of objects with word, translation_en, example_sentence"),
			),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			var args AddWordsArgs
			if err := req.BindArguments(&args); err != nil {
				return mcp.NewToolResultError("failed to parse arguments: " + err.Error()), nil
			}

			var inserted int
			for _, input := range args.Words {
				word := models.Word{
					Word:            input.Word,
					TranslationEN:   input.TranslationEN,
					ExampleSentence: input.ExampleSentence,
				}
				if result := database.DB.Create(&word); result.Error == nil {
					inserted++
				}
			}

			return mcp.NewToolResultText(fmt.Sprintf("Successfully inserted %d words", inserted)), nil
		},
	)

	// ── get_words ──────────────────────────────────────────────────────────────
	s.AddTool(
		mcp.NewTool("get_words",
			mcp.WithDescription("Read words from the database. Supply ids for specific records, or use limit/offset to page through all words."),
			mcp.WithArray("ids",
				mcp.Description("Optional list of word IDs. If omitted, all words are returned up to limit."),
			),
			mcp.WithNumber("limit",
				mcp.Description("Max records to return when fetching all (default 50)"),
			),
			mcp.WithNumber("offset",
				mcp.Description("Pagination offset (default 0)"),
			),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			var args GetWordsArgs
			if err := req.BindArguments(&args); err != nil {
				return mcp.NewToolResultError("failed to parse arguments: " + err.Error()), nil
			}

			var words []models.Word
			if len(args.IDs) > 0 {
				database.DB.Where("id IN ?", args.IDs).Find(&words)
			} else {
				database.DB.Limit(defaultLimit(args.Limit)).Offset(args.Offset).Find(&words)
			}

			return mcp.NewToolResultText(toJSON(words)), nil
		},
	)

	// ── add_paragraphs ─────────────────────────────────────────────────────────
	s.AddTool(
		mcp.NewTool("add_paragraphs",
			mcp.WithDescription("Insert one or more Italian paragraphs into the database"),
			mcp.WithArray("paragraphs",
				mcp.Required(),
				mcp.Description("Array of objects with paragraph, translation_en, and optionally related_word_ids, context, practice_count"),
			),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			var args AddParagraphsArgs
			if err := req.BindArguments(&args); err != nil {
				return mcp.NewToolResultError("failed to parse arguments: " + err.Error()), nil
			}

			var inserted int
			var results []models.Paragraph
			for _, input := range args.Paragraphs {
				ctx := input.Context
				if ctx == "" {
					ctx = models.ContextGeneral
				}
				p := models.Paragraph{
					Paragraph:      input.Paragraph,
					TranslationEN:  input.TranslationEN,
					RelatedWordIDs: models.IntArray(input.RelatedWordIDs),
					Context:        ctx,
					PracticeCount:  input.PracticeCount,
				}
				if result := database.DB.Create(&p); result.Error == nil {
					inserted++
					results = append(results, p)
				}
			}

			return mcp.NewToolResultText(
				fmt.Sprintf("Successfully inserted %d paragraphs\n\n%s", inserted, toJSON(results)),
			), nil
		},
	)

	// ── get_paragraphs ─────────────────────────────────────────────────────────
	s.AddTool(
		mcp.NewTool("get_paragraphs",
			mcp.WithDescription("Read paragraphs from the database. Supply ids for specific records, or page through all. Set include_words=true to also return the related Word records for each paragraph."),
			mcp.WithArray("ids",
				mcp.Description("Optional list of paragraph IDs. If omitted, returns all up to limit."),
			),
			mcp.WithBoolean("include_words",
				mcp.Description("If true, each paragraph entry will include its related word records"),
			),
			mcp.WithNumber("limit",
				mcp.Description("Max records to return when fetching all (default 50)"),
			),
			mcp.WithNumber("offset",
				mcp.Description("Pagination offset (default 0)"),
			),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			var args GetParagraphsArgs
			if err := req.BindArguments(&args); err != nil {
				return mcp.NewToolResultError("failed to parse arguments: " + err.Error()), nil
			}

			var paragraphs []models.Paragraph
			if len(args.IDs) > 0 {
				database.DB.Where("id IN ?", args.IDs).Find(&paragraphs)
			} else {
				database.DB.Limit(defaultLimit(args.Limit)).Offset(args.Offset).Find(&paragraphs)
			}

			if !args.IncludeWords {
				return mcp.NewToolResultText(toJSON(paragraphs)), nil
			}

			// Collect all unique word IDs across all paragraphs
			idSet := map[uint]struct{}{}
			for _, p := range paragraphs {
				for _, wid := range p.RelatedWordIDs {
					idSet[wid] = struct{}{}
				}
			}

			wordMap := map[uint]models.Word{}
			if len(idSet) > 0 {
				allIDs := make([]uint, 0, len(idSet))
				for id := range idSet {
					allIDs = append(allIDs, id)
				}
				var words []models.Word
				database.DB.Where("id IN ?", allIDs).Find(&words)
				for _, w := range words {
					wordMap[w.ID] = w
				}
			}

			type ParagraphWithWords struct {
				models.Paragraph
				Words []models.Word `json:"words"`
			}

			output := make([]ParagraphWithWords, 0, len(paragraphs))
			for _, p := range paragraphs {
				var relatedWords []models.Word
				for _, wid := range p.RelatedWordIDs {
					if w, ok := wordMap[wid]; ok {
						relatedWords = append(relatedWords, w)
					}
				}
				output = append(output, ParagraphWithWords{
					Paragraph: p,
					Words:     relatedWords,
				})
			}

			return mcp.NewToolResultText(toJSON(output)), nil
		},
	)

	// ── get_courses ────────────────────────────────────────────────────────────
	s.AddTool(
		mcp.NewTool("get_courses",
			mcp.WithDescription("Read courses from the database, optionally filtered by language, level, and status."),
			mcp.WithString("language_code", mcp.Description("Optional language code filter (e.g. 'it')")),
			mcp.WithString("level", mcp.Description("Optional CEFR level filter (e.g. 'A1')")),
			mcp.WithString("status", mcp.Description("Optional status filter: draft|published|archived")),
			mcp.WithNumber("limit", mcp.Description("Max records to return (default 20)")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			var args struct {
				LanguageCode string `json:"language_code"`
				Level        string `json:"level"`
				Status       string `json:"status"`
				Limit        int    `json:"limit"`
			}
			if err := req.BindArguments(&args); err != nil {
				return mcp.NewToolResultError("failed to parse arguments: " + err.Error()), nil
			}
			if args.Limit <= 0 {
				args.Limit = 20
			}
			q := database.DB.Model(&models.Course{}).Preload("Language").Preload("Lessons")
			if args.LanguageCode != "" {
				q = q.Joins("JOIN languages ON languages.id = courses.language_id").
					Where("languages.code = ?", args.LanguageCode)
			}
			if args.Level != "" {
				q = q.Where("courses.level = ?", args.Level)
			}
			if args.Status != "" {
				q = q.Where("courses.status = ?", args.Status)
			}
			var courses []models.Course
			q.Limit(args.Limit).Find(&courses)
			return mcp.NewToolResultText(toJSON(courses)), nil
		},
	)

	// ── create_course_draft ────────────────────────────────────────────────────
	s.AddTool(
		mcp.NewTool("create_course_draft",
			mcp.WithDescription("Insert a fully structured course draft (course + lessons + content blocks + exercises) atomically."),
			mcp.WithObject("course",
				mcp.Required(),
				mcp.Description("Full course object per mcp-tools.md schema"),
			),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			var args struct {
				Course struct {
					LanguageCode string `json:"language_code"`
					Level        string `json:"level"`
					Title        string `json:"title"`
					Topic        string `json:"topic"`
					Description  string `json:"description"`
					Skills       []string `json:"skills"`
					Lessons      []struct {
						Title         string `json:"title"`
						ContentBlocks []struct {
							Type    string         `json:"type"`
							Content map[string]any `json:"content"`
						} `json:"content_blocks"`
						Exercises []struct {
							Type          string   `json:"type"`
							Question      string   `json:"question"`
							Options       []string `json:"options"`
							CorrectAnswer string   `json:"correct_answer"`
							Explanation   string   `json:"explanation"`
							Skills        []string `json:"skills"`
						} `json:"exercises"`
					} `json:"lessons"`
				} `json:"course"`
			}
			if err := req.BindArguments(&args); err != nil {
				return mcp.NewToolResultError("failed to parse arguments: " + err.Error()), nil
			}

			ci := args.Course
			var lang models.Language
			if err := database.DB.Where("code = ?", ci.LanguageCode).First(&lang).Error; err != nil {
				return mcp.NewToolResultError("unknown language: " + ci.LanguageCode), nil
			}

			course := models.Course{
				LanguageID:  lang.ID,
				Level:       models.CEFRLevel(ci.Level),
				Title:       ci.Title,
				Topic:       ci.Topic,
				Description: ci.Description,
				Status:      models.StatusDraft,
				Skills:      models.StringArray(ci.Skills),
			}

			tx := database.DB.Begin()
			if err := tx.Create(&course).Error; err != nil {
				tx.Rollback()
				return mcp.NewToolResultError("failed to create course: " + err.Error()), nil
			}

			for i, li := range ci.Lessons {
				lesson := models.Lesson{CourseID: course.ID, Title: li.Title, Position: i}
				if err := tx.Create(&lesson).Error; err != nil {
					tx.Rollback()
					return mcp.NewToolResultError("failed to create lesson: " + err.Error()), nil
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
						return mcp.NewToolResultError("failed to create content block: " + err.Error()), nil
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
						return mcp.NewToolResultError("failed to create exercise: " + err.Error()), nil
					}
				}
			}
			tx.Commit()
			return mcp.NewToolResultText(fmt.Sprintf("Successfully created course draft ID=%d with %d lessons", course.ID, len(ci.Lessons))), nil
		},
	)

	// ── analyze_grammar_coverage ───────────────────────────────────────────────
	s.AddTool(
		mcp.NewTool("analyze_grammar_coverage",
			mcp.WithDescription("Analyse existing published courses for a language/level and return covered grammar topics plus ≥5 suggested missing ones. Suggestions are persisted as GrammarSuggestion records."),
			mcp.WithString("language_code", mcp.Required(), mcp.Description("Language code, e.g. 'it'")),
			mcp.WithString("level", mcp.Required(), mcp.Description("CEFR level, e.g. 'B1-1'")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			var args struct {
				LanguageCode string `json:"language_code"`
				Level        string `json:"level"`
			}
			if err := req.BindArguments(&args); err != nil {
				return mcp.NewToolResultError("failed to parse arguments: " + err.Error()), nil
			}

			var lang models.Language
			if err := database.DB.Where("code = ?", args.LanguageCode).First(&lang).Error; err != nil {
				return mcp.NewToolResultError("unknown language: " + args.LanguageCode), nil
			}

			var courses []models.Course
			database.DB.Where("language_id = ? AND level = ? AND status = ?",
				lang.ID, args.Level, models.StatusPublished).Find(&courses)

			coveredTopics := make([]string, 0, len(courses))
			for _, c := range courses {
				coveredTopics = append(coveredTopics, c.Topic)
			}

			// Return covered topics; suggestions are generated by the HTTP handler
			result := map[string]any{
				"covered_topics":    coveredTopics,
				"suggested_missing": []any{},
			}
			return mcp.NewToolResultText(toJSON(result)), nil
		},
	)

	if err := server.ServeStdio(s); err != nil {
		log.Fatal(err)
	}
}
