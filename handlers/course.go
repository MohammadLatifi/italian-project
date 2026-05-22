package handlers

import (
	"italian_project/database"
	"italian_project/models"

	"github.com/gofiber/fiber/v3"
)

func ListPublishedCourses(c fiber.Ctx) error {
	q := database.DB.Model(&models.Course{}).
		Preload("Language").
		Where("status = ?", models.StatusPublished)
	if lang := c.Query("language"); lang != "" {
		q = q.Joins("JOIN languages ON languages.id = courses.language_id").
			Where("languages.code = ?", lang)
	}
	if level := c.Query("level"); level != "" {
		q = q.Where("courses.level = ?", level)
	}
	var courses []models.Course
	q.Find(&courses)

	type courseWithCount struct {
		models.Course
		LessonCount int64 `json:"lesson_count"`
	}
	result := make([]courseWithCount, len(courses))
	for i, course := range courses {
		var count int64
		database.DB.Model(&models.Lesson{}).Where("course_id = ?", course.ID).Count(&count)
		result[i] = courseWithCount{Course: course, LessonCount: count}
	}
	return c.JSON(result)
}

func GetCourse(c fiber.Ctx) error {
	id := c.Params("courseId")
	var course models.Course
	if err := database.DB.
		Preload("Language").
		Preload("Lessons").
		Where("status = ?", models.StatusPublished).
		First(&course, id).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "course not found"})
	}
	return c.JSON(course)
}

func GetLesson(c fiber.Ctx) error {
	lessonID := c.Params("lessonId")
	var lesson models.Lesson
	if err := database.DB.
		Preload("ContentBlocks").
		Preload("Exercises").
		First(&lesson, lessonID).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "lesson not found"})
	}

	// Strip correct_answer and explanation to prevent client-side cheating
	type safeExercise struct {
		ID       uint                `json:"ID"`
		LessonID uint                `json:"LessonID"`
		Type     models.ExerciseType `json:"Type"`
		Question string              `json:"Question"`
		Options  models.StringArray  `json:"Options"`
		Skills   models.StringArray  `json:"Skills"`
		Position int                 `json:"Position"`
	}
	safeExercises := make([]safeExercise, len(lesson.Exercises))
	for i, ex := range lesson.Exercises {
		safeExercises[i] = safeExercise{
			ID: ex.ID, LessonID: ex.LessonID, Type: ex.Type,
			Question: ex.Question, Options: ex.Options, Skills: ex.Skills, Position: ex.Position,
		}
	}

	type response struct {
		ID            uint                   `json:"ID"`
		CourseID      uint                   `json:"CourseID"`
		Title         string                 `json:"Title"`
		Position      int                    `json:"Position"`
		ContentBlocks []models.ContentBlock  `json:"ContentBlocks"`
		Exercises     []safeExercise         `json:"Exercises"`
		CurrentAttempt *models.LessonAttempt `json:"current_attempt,omitempty"`
	}

	res := response{
		ID:            lesson.ID,
		CourseID:      lesson.CourseID,
		Title:         lesson.Title,
		Position:      lesson.Position,
		ContentBlocks: lesson.ContentBlocks,
		Exercises:     safeExercises,
	}

	if token := c.Get("X-Learner-Token"); token != "" {
		var attempts []models.LessonAttempt
		database.DB.
			Where("lesson_id = ? AND learner_token = ? AND completed = false", lesson.ID, token).
			Order("id DESC").Limit(1).
			Find(&attempts)
		if len(attempts) > 0 {
			res.CurrentAttempt = &attempts[0]
		}
	}

	return c.JSON(res)
}
