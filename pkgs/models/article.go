package models

import (
	"time"

	"gorm.io/gorm"
)

type Article struct {
	gorm.Model
	Author            string          `json:"author" gorm:"not null;index"`
	Title             string          `json:"title" gorm:"not null;index"`
	Slug              string          `json:"slug" gorm:"not null;uniqueIndex"`
	Subtitle          string          `json:"subtitle"`
	Body              string          `json:"body" gorm:"type:text;not null"`
	MainImage         string          `json:"main_image"`
	ImageURL          string          `json:"image_url"`
	MetaTitle         string          `json:"meta_title"`
	MetaDescription   string          `json:"meta_description"`
	MetaKeywords      string          `json:"meta_keywords"`
	CanonicalURL      string          `json:"canonical_url"`
	IsPublished       bool            `json:"is_published" gorm:"not null;default:false;index"`
	PublishedAt       *time.Time      `json:"published_at" gorm:"index"`
	ArticleCategoryID uint            `json:"article_category_id" gorm:"not null;index"`
	ArticleCategory   ArticleCategory `json:"article_category" gorm:"foreignKey:ArticleCategoryID"`
	ProductID         *uint           `json:"product_id" gorm:"index"`
	Product           *Product        `json:"product" gorm:"foreignKey:ProductID"`
}
