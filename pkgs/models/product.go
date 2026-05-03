package models

import (
	"time"

	"gorm.io/gorm"
)

type Product struct {
	gorm.Model
	Name               string          `json:"name" gorm:"not null;index"`
	Slug               string          `json:"slug" gorm:"not null;uniqueIndex"`
	Description        string          `json:"description"`
	Brand              string          `json:"brand" gorm:"index"`
	ModelNumber        string          `json:"model_number"`
	ImageURL           string          `json:"image_url"`
	ProductURL         string          `json:"product_url" gorm:"not null"`
	AffiliateURL       string          `json:"affiliate_url"`
	RetailerName       string          `json:"retailer_name" gorm:"not null;index"`
	RetailerURL        string          `json:"retailer_url"`
	Source             string          `json:"source" gorm:"index"`
	ExternalID         string          `json:"external_id" gorm:"index"`
	Currency           string          `json:"currency" gorm:"size:3;not null"`
	CurrentPriceCents  int64           `json:"current_price_cents" gorm:"not null;index"`
	OriginalPriceCents int64           `json:"original_price_cents"`
	ShippingCostCents  *int64          `json:"shipping_cost_cents"`
	DiscountPercent    int             `json:"discount_percent" gorm:"index"`
	CouponCode         string          `json:"coupon_code"`
	DealScore          int             `json:"deal_score" gorm:"index"`
	Rating             float32         `json:"rating"`
	ReviewCount        int             `json:"review_count"`
	IsAvailable        bool            `json:"is_available" gorm:"not null;default:true;index"`
	IsFeatured         bool            `json:"is_featured" gorm:"not null;default:false;index"`
	StartsAt           *time.Time      `json:"starts_at"`
	EndsAt             *time.Time      `json:"ends_at" gorm:"index"`
	LastCheckedAt      *time.Time      `json:"last_checked_at"`
	CountryID          uint            `json:"country_id" gorm:"not null;index"`
	Country            Country         `json:"country" gorm:"foreignKey:CountryID"`
	CategoryID         uint            `json:"category_id" gorm:"not null;index"`
	Category           ProductCategory `json:"category" gorm:"foreignKey:CategoryID"`
}
