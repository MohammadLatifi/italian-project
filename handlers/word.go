package handlers

import (
	"italian_project/database"
	"italian_project/models"

	"github.com/gofiber/fiber/v3"
)

func GetWords(c fiber.Ctx) error {
	var words []models.Word
	database.DB.Find(&words)
	return c.JSON(words)
}

func GetWord(c fiber.Ctx) error {
	id := c.Params("id")
	var word models.Word

	if err := database.DB.First(&word, id).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "word not found"})
	}

	return c.JSON(word)
}

func CreateWord(c fiber.Ctx) error {
	word := new(models.Word)

	if err := c.Bind().Body(word); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	database.DB.Create(word)
	return c.Status(fiber.StatusCreated).JSON(word)
}

func UpdateWord(c fiber.Ctx) error {
	id := c.Params("id")
	var word models.Word

	if err := database.DB.First(&word, id).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "word not found"})
	}

	if err := c.Bind().Body(&word); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	database.DB.Save(&word)
	return c.JSON(word)
}

func DeleteWord(c fiber.Ctx) error {
	id := c.Params("id")
	var word models.Word

	if err := database.DB.First(&word, id).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "word not found"})
	}

	database.DB.Delete(&word)
	return c.SendStatus(fiber.StatusNoContent)
}
