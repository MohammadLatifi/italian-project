package models

import "gorm.io/gorm"

type Word struct {
	gorm.Model
	LanguageID        *uint  `gorm:"index"` // NULL = Italian (backward compat)
	Word              string `gorm:"not null"`
	TranslationEN     string `gorm:"not null"`
	ExampleSentence   string
	PracticeCount     int  `gorm:"default:0"`
	HowManyFalse      int  `gorm:"default:0"`
	LastPracticeFalse bool `gorm:"default:false"`
}
