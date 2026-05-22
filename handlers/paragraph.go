package handlers

import (
	"italian_project/database"
	"italian_project/models"

	"github.com/gofiber/fiber/v3"
)

func GetParagraphs(c fiber.Ctx) error {
	var paragraphs []models.Paragraph
	database.DB.Find(&paragraphs)
	return c.JSON(paragraphs)
}

func GetParagraph(c fiber.Ctx) error {
	id := c.Params("id")
	var paragraph models.Paragraph

	if err := database.DB.First(&paragraph, id).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "paragraph not found"})
	}

	return c.JSON(paragraph)
}

func CreateParagraph(c fiber.Ctx) error {
	paragraph := new(models.Paragraph)

	if err := c.Bind().Body(paragraph); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	if paragraph.Context == "" {
		paragraph.Context = models.ContextGeneral
	}

	database.DB.Create(paragraph)
	return c.Status(fiber.StatusCreated).JSON(paragraph)
}

func UpdateParagraph(c fiber.Ctx) error {
	id := c.Params("id")
	var paragraph models.Paragraph

	if err := database.DB.First(&paragraph, id).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "paragraph not found"})
	}

	if err := c.Bind().Body(&paragraph); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	database.DB.Save(&paragraph)
	return c.JSON(paragraph)
}

func DeleteParagraph(c fiber.Ctx) error {
	id := c.Params("id")
	var paragraph models.Paragraph

	if err := database.DB.First(&paragraph, id).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "paragraph not found"})
	}

	database.DB.Delete(&paragraph)
	return c.SendStatus(fiber.StatusNoContent)
}

// GetParagraphWords returns the paragraph along with its related Word records.
func GetParagraphWords(c fiber.Ctx) error {
	id := c.Params("id")
	var paragraph models.Paragraph

	if err := database.DB.First(&paragraph, id).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "paragraph not found"})
	}

	var words []models.Word
	if len(paragraph.RelatedWordIDs) > 0 {
		database.DB.Where("id IN ?", []uint(paragraph.RelatedWordIDs)).Find(&words)
	}

	return c.JSON(fiber.Map{
		"paragraph": paragraph,
		"words":     words,
	})
}
