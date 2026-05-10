package models

import "gorm.io/gorm"

type ProductStorefront struct {
	gorm.Model
	ProductID    uint       `json:"product_id" gorm:"not null;uniqueIndex:idx_product_storefront"`
	Product      Product    `json:"product" gorm:"foreignKey:ProductID"`
	StorefrontID uint       `json:"storefront_id" gorm:"not null;uniqueIndex:idx_product_storefront;index"`
	Storefront   Storefront `json:"storefront" gorm:"foreignKey:StorefrontID"`
}
