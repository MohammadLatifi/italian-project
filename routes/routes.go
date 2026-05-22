package routes

import (
	"italian_project/handlers"

	"github.com/gofiber/fiber/v3"
)

func Register(app *fiber.App) {
	api := app.Group("/api")

	// ── existing ──────────────────────────────────────────────────────────────

	words := api.Group("/words")
	words.Get("/", handlers.GetWords)
	words.Get("/:id", handlers.GetWord)
	words.Post("/", handlers.CreateWord)
	words.Put("/:id", handlers.UpdateWord)
	words.Delete("/:id", handlers.DeleteWord)

	paragraphs := api.Group("/paragraphs")
	paragraphs.Get("/", handlers.GetParagraphs)
	paragraphs.Get("/:id", handlers.GetParagraph)
	paragraphs.Get("/:id/words", handlers.GetParagraphWords)
	paragraphs.Post("/", handlers.CreateParagraph)
	paragraphs.Put("/:id", handlers.UpdateParagraph)
	paragraphs.Delete("/:id", handlers.DeleteParagraph)

	generate := api.Group("/generate")
	generate.Post("/paragraph", handlers.GenerateParagraph)

	// ── languages (public) ───────────────────────────────────────────────────
	api.Get("/languages", handlers.GetLanguages)

	// ── admin (protected) ────────────────────────────────────────────────────
	api.Post("/admin/login", handlers.AdminLogin)

	admin := api.Group("/admin", handlers.AdminAuth)
	admin.Get("/courses", handlers.AdminListCourses)
	admin.Post("/courses/generate", handlers.AdminGenerateCourse)
	admin.Get("/courses/:id", handlers.AdminGetCourse)
	admin.Patch("/courses/:id", handlers.AdminPatchCourse)
	admin.Post("/courses/:id/publish", handlers.AdminPublishCourse)
	admin.Post("/courses/:id/archive", handlers.AdminArchiveCourse)
	admin.Post("/courses/:id/generate-lessons", handlers.AdminGenerateLessons)
	admin.Get("/languages", handlers.GetLanguages)
	admin.Post("/languages", handlers.AdminCreateLanguage)
	admin.Patch("/languages/:id", handlers.AdminUpdateLanguage)
	admin.Delete("/languages/:id", handlers.AdminDeleteLanguage)
	admin.Get("/grammar-suggestions", handlers.AdminListGrammarSuggestions)
	admin.Post("/grammar-suggestions/analyze", handlers.AdminAnalyzeGrammarSuggestions)
	admin.Post("/grammar-suggestions/:id/accept", handlers.AdminAcceptGrammarSuggestion)
	admin.Post("/grammar-suggestions/:id/dismiss", handlers.AdminDismissGrammarSuggestion)

	// ── learner (public) ─────────────────────────────────────────────────────
	api.Get("/courses", handlers.ListPublishedCourses)
	api.Get("/courses/:courseId", handlers.GetCourse)
	api.Get("/courses/:courseId/lessons/:lessonId", handlers.GetLesson)
	api.Post("/courses/:courseId/lessons/:lessonId/attempts", handlers.StartAttempt)
	api.Post("/courses/:courseId/lessons/:lessonId/attempts/:attemptId/answer", handlers.SubmitAnswer)
	api.Post("/courses/:courseId/lessons/:lessonId/attempts/:attemptId/complete", handlers.CompleteAttempt)
}
