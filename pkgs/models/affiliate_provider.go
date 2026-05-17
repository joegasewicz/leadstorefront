package models

import "gorm.io/gorm"

type AffiliateProvider struct {
	gorm.Model
	Name                    string `json:"name" gorm:"not null;uniqueIndex"`
	Slug                    string `json:"slug" gorm:"not null;uniqueIndex"`
	Description             string `json:"description" gorm:"type:text"`
	RegistrationURL         string `json:"registration_url"`
	ApprovalRequirements    string `json:"approval_requirements" gorm:"type:text"`
	SupportedMarketsJSON    JSONB  `json:"supported_markets_json" gorm:"type:jsonb"`
	TrackingParameterFormat string `json:"tracking_parameter_format"`
	APIAvailable            bool   `json:"api_available" gorm:"not null;default:false"`
	DeepLinkSupport         bool   `json:"deep_link_support" gorm:"not null;default:false"`
}
