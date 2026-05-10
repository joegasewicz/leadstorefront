package models

import "gorm.io/gorm"

type Storefront struct {
	gorm.Model
	Name             string  `json:"name" gorm:"not null;index"`
	Slug             string  `json:"slug" gorm:"not null;uniqueIndex"`
	Domain           string  `json:"domain" gorm:"not null;uniqueIndex"`
	Description      string  `json:"description"`
	LogoURL          string  `json:"logo_url"`
	IsActive         bool    `json:"is_active" gorm:"not null;default:true;index"`
	PrimaryCountryID uint    `json:"primary_country_id" gorm:"not null;index"`
	PrimaryCountry   Country `json:"primary_country" gorm:"foreignKey:PrimaryCountryID"`
	OwnerID          *uint   `json:"owner_id" gorm:"index"`
	Owner            *User   `json:"owner" gorm:"foreignKey:OwnerID"`
}
