package models

import "gorm.io/gorm"

type ArticleStorefront struct {
	gorm.Model
	ArticleID    uint       `json:"article_id" gorm:"not null;uniqueIndex:idx_article_storefront"`
	Article      Article    `json:"article" gorm:"foreignKey:ArticleID"`
	StorefrontID uint       `json:"storefront_id" gorm:"not null;uniqueIndex:idx_article_storefront;index"`
	Storefront   Storefront `json:"storefront" gorm:"foreignKey:StorefrontID"`
}
