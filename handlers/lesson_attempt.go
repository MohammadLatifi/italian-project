package handlers

import (
	"fmt"
	"time"

	"italian_project/database"
	"italian_project/models"

	"github.com/gofiber/fiber/v3"
)

func StartAttempt(c fiber.Ctx) error {
	token := c.Get("X-Learner-Token")
	if token == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "X-Learner-Token header required"})
	}
	lessonID := c.Params("lessonId")
	attempt := models.LessonAttempt{
		LearnerToken:    token,
		CurrentPosition: 0,
		Answers:         models.AnswerMap{},
		Completed:       false,
	}
	// Parse lessonID into uint
	var lesson models.Lesson
	if err := database.DB.First(&lesson, lessonID).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "lesson not found"})
	}
	attempt.LessonID = lesson.ID
	database.DB.Create(&attempt)
	return c.Status(fiber.StatusCreated).JSON(attempt)
}

func SubmitAnswer(c fiber.Ctx) error {
	token := c.Get("X-Learner-Token")
	if token == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "X-Learner-Token header required"})
	}

	attemptID := c.Params("attemptId")
	var attempt models.LessonAttempt
	if err := database.DB.First(&attempt, attemptID).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "attempt not found"})
	}
	if attempt.LearnerToken != token {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "token mismatch"})
	}

	var body struct {
		ExerciseID uint   `json:"exercise_id"`
		Answer     string `json:"answer"`
	}
	if err := c.Bind().Body(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	var exercise models.Exercise
	if err := database.DB.First(&exercise, body.ExerciseID).Error; err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "exercise not found"})
	}
	if exercise.LessonID != attempt.LessonID {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "exercise does not belong to this lesson"})
	}

	correct := body.Answer == exercise.CorrectAnswer
	if attempt.Answers == nil {
		attempt.Answers = models.AnswerMap{}
	}
	attempt.Answers[fmt.Sprintf("%d", body.ExerciseID)] = body.Answer
	attempt.CurrentPosition++
	database.DB.Save(&attempt)

	var explanation *string
	if !correct && exercise.Explanation != "" {
		explanation = &exercise.Explanation
	}

	return c.JSON(fiber.Map{
		"correct":        correct,
		"correct_answer": exercise.CorrectAnswer,
		"explanation":    explanation,
		"next_position":  attempt.CurrentPosition,
	})
}

func CompleteAttempt(c fiber.Ctx) error {
	token := c.Get("X-Learner-Token")
	attemptID := c.Params("attemptId")
	var attempt models.LessonAttempt
	if err := database.DB.First(&attempt, attemptID).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "attempt not found"})
	}
	if token != "" && attempt.LearnerToken != token {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "token mismatch"})
	}

	var exercises []models.Exercise
	database.DB.Where("lesson_id = ?", attempt.LessonID).Find(&exercises)

	correct := 0
	for _, ex := range exercises {
		if given, ok := attempt.Answers[fmt.Sprintf("%d", ex.ID)]; ok {
			if given == ex.CorrectAnswer {
				correct++
			}
		}
	}

	score := correct
	now := time.Now()
	attempt.Score = &score
	attempt.Completed = true
	attempt.CompletedAt = &now
	database.DB.Save(&attempt)

	return c.JSON(fiber.Map{
		"ID":           attempt.ID,
		"score":        score,
		"total":        len(exercises),
		"completed":    true,
		"completed_at": now,
	})
}
