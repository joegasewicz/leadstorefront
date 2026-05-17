package models

import "gorm.io/gorm"

type OutboundClick struct {
	gorm.Model
	StorefrontID    uint       `json:"storefront_id" gorm:"not null;index"`
	Storefront      Storefront `json:"storefront" gorm:"foreignKey:StorefrontID"`
	ProductID       *uint      `json:"product_id" gorm:"index"`
	Product         *Product   `json:"product" gorm:"foreignKey:ProductID"`
	VisitorID       string     `json:"visitor_id" gorm:"index"`
	CountryCode     string     `json:"country_code" gorm:"size:2;index"`
	DestinationURL  string     `json:"destination_url" gorm:"type:text;not null"`
	AttributionJSON JSONB      `json:"attribution_json" gorm:"type:jsonb"`
	CampaignSource  string     `json:"campaign_source" gorm:"index"`
	CampaignMedium  string     `json:"campaign_medium"`
	CampaignName    string     `json:"campaign_name" gorm:"index"`
	AffiliateSource string     `json:"affiliate_source" gorm:"index"`
	MarketSource    string     `json:"market_source" gorm:"index"`
	ClickID         string     `json:"click_id" gorm:"index"`
	LandingPath     string     `json:"landing_path"`
	Referrer        string     `json:"referrer"`
	UserAgent       string     `json:"user_agent"`
}
