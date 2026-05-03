package models

import "gorm.io/gorm"

type ArticleCategory struct {
	gorm.Model
	Name string `json:"name" gorm:"not null;uniqueIndex"`
}
