package database

import (
	"fmt"
	"log"
	"os"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"italian_project/models"
)

var DB *gorm.DB

func Connect() {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_NAME"),
	)

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database: ", err)
	}

	db.AutoMigrate(
		&models.Language{},
		&models.Word{},
		&models.Paragraph{},
		&models.Course{},
		&models.Lesson{},
		&models.ContentBlock{},
		&models.Exercise{},
		&models.LessonAttempt{},
		&models.GrammarSuggestion{},
	)

	// Seed Italian language if missing
	var count int64
	db.Model(&models.Language{}).Count(&count)
	if count == 0 {
		db.Create(&models.Language{Code: "it", Name: "Italian"})
	}

	DB = db
	log.Println("Database connected successfully")
}
