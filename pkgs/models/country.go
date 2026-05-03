package models

import "gorm.io/gorm"

type Country struct {
	gorm.Model
	Code     string `json:"code" gorm:"size:2;not null;uniqueIndex"`
	Name     string `json:"name" gorm:"not null"`
	Currency string `json:"currency" gorm:"size:3;not null"`
}
