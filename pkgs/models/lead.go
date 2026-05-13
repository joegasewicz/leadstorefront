package models

import "gorm.io/gorm"

type Lead struct {
	gorm.Model
	StorefrontID uint       `json:"storefront_id" gorm:"not null;index"`
	Storefront   Storefront `json:"storefront" gorm:"foreignKey:StorefrontID"`
	Source       string     `json:"source"`
	Tracking     string     `json:"tracking"`
	ValuesJSON   string     `json:"values_json" gorm:"type:text"`
}
