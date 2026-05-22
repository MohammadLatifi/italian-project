package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"strings"

	"gorm.io/gorm"
)

type CEFRLevel string

const (
	LevelA1  CEFRLevel = "A1"
	LevelA2  CEFRLevel = "A2"
	LevelB11 CEFRLevel = "B1-1"
	LevelB12 CEFRLevel = "B1-2"
	LevelB2  CEFRLevel = "B2"
	LevelC1  CEFRLevel = "C1"
	LevelC2  CEFRLevel = "C2"
)

type CourseStatus string

const (
	StatusDraft     CourseStatus = "draft"
	StatusPublished CourseStatus = "published"
	StatusArchived  CourseStatus = "archived"
)

type ContentBlockType string

const (
	BlockVocabularyList     ContentBlockType = "vocabulary-list"
	BlockGrammarNote        ContentBlockType = "grammar-note"
	BlockConversationExample ContentBlockType = "conversation-example"
	BlockReadingPassage     ContentBlockType = "reading-passage"
)

type ExerciseType string

const (
	ExerciseMultipleChoice             ExerciseType = "multiple-choice"
	ExerciseFillInTheBlank             ExerciseType = "fill-in-the-blank"
	ExerciseConversationReconstruction ExerciseType = "conversation-reconstruction"
)

// StringArray stores a slice of strings as a JSON string in MySQL.
type StringArray []string

func (a StringArray) Value() (driver.Value, error) {
	if a == nil {
		return "[]", nil
	}
	b, err := json.Marshal(a)
	return string(b), err
}

func (a *StringArray) Scan(value any) error {
	if value == nil {
		*a = StringArray{}
		return nil
	}
	var b []byte
	switch v := value.(type) {
	case []byte:
		b = v
	case string:
		b = []byte(v)
	default:
		return fmt.Errorf("unsupported type for StringArray: %T", v)
	}
	return json.Unmarshal(b, a)
}

func (a StringArray) Contains(s string) bool {
	for _, v := range a {
		if strings.EqualFold(v, s) {
			return true
		}
	}
	return false
}

// JSONMap stores arbitrary JSON content for ContentBlock.
type JSONMap map[string]any

func (m JSONMap) Value() (driver.Value, error) {
	if m == nil {
		return "{}", nil
	}
	b, err := json.Marshal(m)
	return string(b), err
}

func (m *JSONMap) Scan(value any) error {
	if value == nil {
		*m = JSONMap{}
		return nil
	}
	var b []byte
	switch v := value.(type) {
	case []byte:
		b = v
	case string:
		b = []byte(v)
	default:
		return fmt.Errorf("unsupported type for JSONMap: %T", v)
	}
	return json.Unmarshal(b, m)
}

type Course struct {
	gorm.Model
	LanguageID  uint         `gorm:"not null;index"`
	Language    Language     `gorm:"foreignKey:LanguageID"`
	Level       CEFRLevel    `gorm:"type:enum('A1','A2','B1-1','B1-2','B2','C1','C2');not null"`
	Title       string       `gorm:"type:varchar(255);not null"`
	Topic       string       `gorm:"type:text;not null"`
	Description string       `gorm:"type:text"`
	Status      CourseStatus `gorm:"type:enum('draft','published','archived');default:'draft';not null"`
	Skills      StringArray  `gorm:"type:json;not null"`
	Lessons     []Lesson     `gorm:"foreignKey:CourseID"`
}

type Lesson struct {
	gorm.Model
	CourseID      uint           `gorm:"not null;index"`
	Title         string         `gorm:"type:varchar(255);not null"`
	Position      int            `gorm:"default:0;not null"`
	ContentBlocks []ContentBlock `gorm:"foreignKey:LessonID"`
	Exercises     []Exercise     `gorm:"foreignKey:LessonID"`
}

type ContentBlock struct {
	gorm.Model
	LessonID uint             `gorm:"not null;index"`
	Type     ContentBlockType `gorm:"type:enum('vocabulary-list','grammar-note','conversation-example','reading-passage');not null"`
	Content  JSONMap          `gorm:"type:json;not null"`
	Position int              `gorm:"default:0;not null"`
}

type Exercise struct {
	gorm.Model
	LessonID      uint         `gorm:"not null;index"`
	Type          ExerciseType `gorm:"type:enum('multiple-choice','fill-in-the-blank','conversation-reconstruction');not null"`
	Question      string       `gorm:"type:text;not null"`
	Options       StringArray  `gorm:"type:json"`
	CorrectAnswer string       `gorm:"type:text;not null"`
	Explanation   string       `gorm:"type:text"`
	Skills        StringArray  `gorm:"type:json;not null"`
	Position      int          `gorm:"default:0;not null"`
}
