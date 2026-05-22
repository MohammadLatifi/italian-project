package handlers

import (
	"italian_project/database"
	"italian_project/models"

	"github.com/gofiber/fiber/v3"
)

func GetLanguages(c fiber.Ctx) error {
	var languages []models.Language
	database.DB.Find(&languages)
	return c.JSON(languages)
}

func AdminCreateLanguage(c fiber.Ctx) error {
	var body struct {
		Code string `json:"code"`
		Name string `json:"name"`
	}
	if err := c.Bind().Body(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	if body.Code == "" || body.Name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "code and name are required"})
	}
	lang := models.Language{Code: body.Code, Name: body.Name}
	if err := database.DB.Create(&lang).Error; err != nil {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": "language code already exists"})
	}
	return c.Status(fiber.StatusCreated).JSON(lang)
}

func AdminUpdateLanguage(c fiber.Ctx) error {
	id := c.Params("id")
	var lang models.Language
	if err := database.DB.First(&lang, id).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "language not found"})
	}
	var body struct {
		Name *string `json:"name"`
	}
	if err := c.Bind().Body(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	if body.Name != nil {
		lang.Name = *body.Name
	}
	database.DB.Save(&lang)
	return c.JSON(lang)
}

func AdminDeleteLanguage(c fiber.Ctx) error {
	id := c.Params("id")
	var lang models.Language
	if err := database.DB.First(&lang, id).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "language not found"})
	}
	var count int64
	database.DB.Model(&models.Course{}).Where("language_id = ?", lang.ID).Count(&count)
	if count > 0 {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": "cannot delete language with existing courses"})
	}
	database.DB.Delete(&lang)
	return c.SendStatus(fiber.StatusNoContent)
}
