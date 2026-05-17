package database

import (
	"leadstorefront/pkgs/models"

	"golang.org/x/crypto/bcrypt"
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

	roles := []models.Role{
		{Name: "super"},
		{Name: "admin"},
		{Name: "editor"},
		{Name: "user"},
	}

	for _, role := range roles {
		if err := db.Where("name = ?", role.Name).FirstOrCreate(&role).Error; err != nil {
			return err
		}
	}

	affiliateProviders := []models.AffiliateProvider{
		{
			Name:                    "Booking.com",
			Slug:                    "booking-com",
			Description:             "Accommodation affiliate program with market-aware AID tracking and deep-link support.",
			RegistrationURL:         "https://www.booking.com/affiliate-program/v2/index.html",
			ApprovalRequirements:    "Requires affiliate registration, website/storefront details, tax details, and program approval before tracking is active.",
			SupportedMarketsJSON:    models.JSONB(`["uk","us","au","za","nz","ca","ie","sg"]`),
			TrackingParameterFormat: "aid={aid}&label={subid}",
			APIAvailable:            false,
			DeepLinkSupport:         true,
		},
		{
			Name:                    "Awin",
			Slug:                    "awin",
			Description:             "Affiliate network for retail and travel advertisers with click reference tracking.",
			RegistrationURL:         "https://www.awin.com/gb/publishers",
			ApprovalRequirements:    "Requires publisher account approval and advertiser program approval.",
			SupportedMarketsJSON:    models.JSONB(`["uk","us","au","ca","ie"]`),
			TrackingParameterFormat: "clickref={clickref}",
			APIAvailable:            true,
			DeepLinkSupport:         true,
		},
		{
			Name:                    "CJ Affiliate",
			Slug:                    "cj-affiliate",
			Description:             "Affiliate network for commerce advertisers with SID tracking support.",
			RegistrationURL:         "https://www.cj.com/publishers",
			ApprovalRequirements:    "Requires publisher signup, profile completion, and advertiser approval.",
			SupportedMarketsJSON:    models.JSONB(`["uk","us","au","ca","ie","sg"]`),
			TrackingParameterFormat: "sid={subid}",
			APIAvailable:            true,
			DeepLinkSupport:         true,
		},
		{
			Name:                    "Amazon Associates UK",
			Slug:                    "amazon-associates-uk",
			Description:             "Retail affiliate program for Amazon UK product links.",
			RegistrationURL:         "https://affiliate-program.amazon.co.uk/",
			ApprovalRequirements:    "Requires Associates account approval and qualifying sales within Amazon's review window.",
			SupportedMarketsJSON:    models.JSONB(`["uk"]`),
			TrackingParameterFormat: "tag={affiliate_id}&ascsubtag={subid}",
			APIAvailable:            true,
			DeepLinkSupport:         true,
		},
		{
			Name:                    "Travelpayouts",
			Slug:                    "travelpayouts",
			Description:             "Travel affiliate platform supporting markers for flights, hotels, tours, and travel products.",
			RegistrationURL:         "https://www.travelpayouts.com/",
			ApprovalRequirements:    "Requires publisher signup and partner program activation.",
			SupportedMarketsJSON:    models.JSONB(`["uk","us","au","za","nz","ca","ie","sg"]`),
			TrackingParameterFormat: "marker={marker}",
			APIAvailable:            true,
			DeepLinkSupport:         true,
		},
	}
	for _, provider := range affiliateProviders {
		if err := db.Where("slug = ?", provider.Slug).FirstOrCreate(&provider).Error; err != nil {
			return err
		}
	}

	var superRole models.Role
	if err := db.Where("name = ?", "super").First(&superRole).Error; err != nil {
		return err
	}

	adminPasswordHash, err := bcrypt.GenerateFromPassword([]byte("Status1234!"), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	adminUser := models.User{
		Email:    "joegoosebass@gmail.com",
		Password: string(adminPasswordHash),
		RoleID:   superRole.ID,
	}

	if err := db.Where("email = ?", adminUser.Email).FirstOrCreate(&adminUser).Error; err != nil {
		return err
	}
	if err := db.Model(&models.User{}).Where("email = ?", adminUser.Email).Update("role_id", superRole.ID).Error; err != nil {
		return err
	}

	productCategories := []models.ProductCategory{
		{Name: "Smartphones"},
		{Name: "Laptops"},
		{Name: "Tablets"},
		{Name: "Desktop Computers"},
		{Name: "Monitors"},
		{Name: "Headphones"},
		{Name: "Earbuds"},
		{Name: "Smartwatches"},
		{Name: "Fitness Trackers"},
		{Name: "Gaming Consoles"},
		{Name: "Gaming Accessories"},
		{Name: "Cameras"},
		{Name: "Drones"},
		{Name: "Smart Home"},
		{Name: "Speakers"},
		{Name: "TVs"},
		{Name: "Streaming Devices"},
		{Name: "Networking"},
		{Name: "Storage"},
		{Name: "Computer Components"},
		{Name: "Keyboards"},
		{Name: "Mice"},
		{Name: "Printers"},
		{Name: "Chargers"},
		{Name: "Power Banks"},
		{Name: "Cables"},
		{Name: "VR Headsets"},
		{Name: "Projectors"},
		{Name: "E-Readers"},
		{Name: "Software"},
		{Name: "Smart Rings"},
		{Name: "Action Cameras"},
		{Name: "Dash Cams"},
		{Name: "Security Cameras"},
		{Name: "Video Doorbells"},
		{Name: "Robot Vacuums"},
		{Name: "Smart Lighting"},
		{Name: "Smart Thermostats"},
		{Name: "Smart Plugs"},
		{Name: "Portable Speakers"},
		{Name: "Soundbars"},
		{Name: "Microphones"},
		{Name: "Webcams"},
		{Name: "Docking Stations"},
		{Name: "USB Hubs"},
		{Name: "Memory Cards"},
		{Name: "External SSDs"},
		{Name: "NAS Drives"},
		{Name: "Routers"},
		{Name: "Mesh WiFi"},
		{Name: "Modems"},
		{Name: "Graphics Cards"},
		{Name: "Processors"},
		{Name: "Motherboards"},
		{Name: "RAM"},
		{Name: "PC Cases"},
		{Name: "Power Supplies"},
		{Name: "Cooling"},
		{Name: "Drawing Tablets"},
		{Name: "3D Printers"},
	}

	for _, category := range productCategories {
		if err := db.Where("name = ?", category.Name).FirstOrCreate(&category).Error; err != nil {
			return err
		}
	}

	articleCategories := []models.ArticleCategory{
		{Name: "Buying Guides"},
		{Name: "Deal Roundups"},
		{Name: "Reviews"},
		{Name: "Comparisons"},
		{Name: "How To"},
		{Name: "News"},
	}

	for _, category := range articleCategories {
		if err := db.Where("name = ?", category.Name).FirstOrCreate(&category).Error; err != nil {
			return err
		}
	}

	return nil
}
