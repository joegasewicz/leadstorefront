package database

import (
	"gadgetscout/pkgs/models"

	"gorm.io/gorm"
)

func Seed(db *gorm.DB) error {
	countries := []models.Country{
		{Code: "us", Name: "United States", Currency: "USD"},
		{Code: "uk", Name: "United Kingdom", Currency: "GBP"},
		{Code: "au", Name: "Australia", Currency: "AUD"},
		{Code: "za", Name: "South Africa", Currency: "ZAR"},
		{Code: "nz", Name: "New Zealand", Currency: "NZD"},
		{Code: "ca", Name: "Canada", Currency: "CAD"},
		{Code: "ie", Name: "Ireland", Currency: "EUR"},
		{Code: "sg", Name: "Singapore", Currency: "SGD"},
	}

	for _, country := range countries {
		if err := db.Where("code = ?", country.Code).FirstOrCreate(&country).Error; err != nil {
			return err
		}
	}

	return nil
}
