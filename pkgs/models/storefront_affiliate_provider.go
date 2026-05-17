package models

import "gorm.io/gorm"

type StorefrontAffiliateProvider struct {
	gorm.Model
	StorefrontID        uint              `json:"storefront_id" gorm:"not null;uniqueIndex:idx_storefront_affiliate_provider;index"`
	Storefront          Storefront        `json:"storefront" gorm:"foreignKey:StorefrontID"`
	AffiliateProviderID uint              `json:"affiliate_provider_id" gorm:"not null;uniqueIndex:idx_storefront_affiliate_provider;index"`
	AffiliateProvider   AffiliateProvider `json:"affiliate_provider" gorm:"foreignKey:AffiliateProviderID"`
	AffiliateID         string            `json:"affiliate_id"`
	PartnerID           string            `json:"partner_id"`
	AID                 string            `json:"aid"`
	CID                 string            `json:"cid"`
	SubIDFormat         string            `json:"sub_id_format"`
	ClickRefFormat      string            `json:"click_ref_format"`
	TrackingDomain      string            `json:"tracking_domain"`
	DeepLinkBaseURL     string            `json:"deep_link_base_url"`
	APIKey              string            `json:"api_key"`
	DefaultMarket       string            `json:"default_market" gorm:"index"`
	CommissionType      string            `json:"commission_type"`
	ApprovalStatus      string            `json:"approval_status" gorm:"not null;default:'registration_required';index"`
	IsActive            bool              `json:"is_active" gorm:"not null;default:false;index"`
}
