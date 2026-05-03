package models

import "gorm.io/gorm"

type Role struct {
	gorm.Model
	Name string `json:"name" gorm:"not null;uniqueIndex"`
}
