package routes

import (
	"italian_project/handlers"

	"github.com/gofiber/fiber/v3"
)

func Register(app *fiber.App) {
	api := app.Group("/api")

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
}
