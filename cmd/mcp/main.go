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

	if err := server.ServeStdio(s); err != nil {
		log.Fatal(err)
	}
}
