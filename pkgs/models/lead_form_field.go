package models

import "gorm.io/gorm"

type LeadFormField struct {
	gorm.Model
	StorefrontID uint       `json:"storefront_id" gorm:"not null;index"`
	Storefront   Storefront `json:"storefront" gorm:"foreignKey:StorefrontID"`
	Label        string     `json:"label" gorm:"not null"`
	Name         string     `json:"name" gorm:"not null;index"`
	Type         string     `json:"type" gorm:"not null"`
	Options      string     `json:"options"`
	IsRequired   bool       `json:"is_required" gorm:"not null;default:false"`
	SortOrder    int        `json:"sort_order" gorm:"not null;default:0"`
}
