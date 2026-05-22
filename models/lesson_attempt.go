package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"

	"gorm.io/gorm"
)

// AnswerMap stores exercise_id → answer given as JSON.
type AnswerMap map[string]string

func (m AnswerMap) Value() (driver.Value, error) {
	if m == nil {
		return "{}", nil
	}
	b, err := json.Marshal(m)
	return string(b), err
}

func (m *AnswerMap) Scan(value any) error {
	if value == nil {
		*m = AnswerMap{}
		return nil
	}
	var b []byte
	switch v := value.(type) {
	case []byte:
		b = v
	case string:
		b = []byte(v)
	default:
		return fmt.Errorf("unsupported type for AnswerMap: %T", v)
	}
	return json.Unmarshal(b, m)
}

type LessonAttempt struct {
	gorm.Model
	LessonID        uint       `gorm:"not null;index"`
	LearnerToken    string     `gorm:"type:varchar(64);not null;index"`
	CurrentPosition int        `gorm:"default:0;not null"`
	Answers         AnswerMap  `gorm:"type:json"`
	Score           *int       `gorm:"default:null"`
	Completed       bool       `gorm:"default:false;not null"`
	CompletedAt     *time.Time `gorm:"default:null"`
}

type GrammarSuggestionStatus string

const (
	SuggestionPending   GrammarSuggestionStatus = "pending"
	SuggestionAccepted  GrammarSuggestionStatus = "accepted"
	SuggestionDismissed GrammarSuggestionStatus = "dismissed"
)

type GrammarSuggestion struct {
	gorm.Model
	LanguageID uint                    `gorm:"not null;index"`
	Language   Language                `gorm:"foreignKey:LanguageID"`
	Level      CEFRLevel               `gorm:"type:enum('A1','A2','B1-1','B1-2','B2','C1','C2');not null"`
	Topic      string                  `gorm:"type:text;not null"`
	Rationale  string                  `gorm:"type:text"`
	Status     GrammarSuggestionStatus `gorm:"type:enum('pending','accepted','dismissed');default:'pending';not null"`
}
