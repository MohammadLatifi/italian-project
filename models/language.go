package models

import "gorm.io/gorm"

type Language struct {
	gorm.Model
	Code string `gorm:"type:varchar(10);not null;uniqueIndex"`
	Name string `gorm:"type:varchar(100);not null"`
}
