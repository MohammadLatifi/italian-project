package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"

	"gorm.io/gorm"
)

type ParagraphContext string

const (
	ContextGeneral  ParagraphContext = "general"
	ContextSport    ParagraphContext = "sport"
	ContextArt      ParagraphContext = "art"
	ContextSoftware ParagraphContext = "software"
	ContextFinance  ParagraphContext = "finance"
	ContextLaw      ParagraphContext = "law"
	ContextMedical  ParagraphContext = "medical"
	ContextScience  ParagraphContext = "science"
	ContextFamily   ParagraphContext = "family"
	ContextNature   ParagraphContext = "nature"
)

// IntArray stores a slice of uints as a JSON string in MySQL.
type IntArray []uint

func (a IntArray) Value() (driver.Value, error) {
	if a == nil {
		return "[]", nil
	}
	b, err := json.Marshal(a)
	return string(b), err
}

func (a *IntArray) Scan(value any) error {
	if value == nil {
		*a = IntArray{}
		return nil
	}
	var bytes []byte
	switch v := value.(type) {
	case []byte:
		bytes = v
	case string:
		bytes = []byte(v)
	default:
		return fmt.Errorf("unsupported type for IntArray: %T", v)
	}
	return json.Unmarshal(bytes, a)
}

type Paragraph struct {
	gorm.Model
	Paragraph      string          `gorm:"not null"`
	TranslationEN  string          `gorm:"not null"`
	RelatedWordIDs IntArray        `gorm:"type:json"`
	PracticeCount  int             `gorm:"default:0"`
	Context        ParagraphContext `gorm:"type:enum('general','sport','art','software','finance','law','medical','science','family','nature');default:'general'"`
}
