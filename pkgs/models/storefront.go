package models

import "gorm.io/gorm"

type Storefront struct {
	gorm.Model
	Name             string  `json:"name" gorm:"not null;index"`
	Slug             string  `json:"slug" gorm:"not null;uniqueIndex"`
	Domain           string  `json:"domain" gorm:"not null;uniqueIndex"`
	Description      string  `json:"description"`
	LogoURL          string  `json:"logo_url"`
	LogoWidthPx      int     `json:"logo_width_px" gorm:"not null;default:305"`
	HeroTitle        string  `json:"hero_title"`
	HeroSubtitle     string  `json:"hero_subtitle"`
	HeroImageURL     string  `json:"hero_image_url"`
	AboutTitle       string  `json:"about_title"`
	AboutBody        string  `json:"about_body"`
	IsActive         bool    `json:"is_active" gorm:"not null;default:true;index"`
	PrimaryCountryID uint    `json:"primary_country_id" gorm:"not null;index"`
	PrimaryCountry   Country `json:"primary_country" gorm:"foreignKey:PrimaryCountryID"`
	OwnerID          *uint   `json:"owner_id" gorm:"index"`
	Owner            *User   `json:"owner" gorm:"foreignKey:OwnerID"`
}
