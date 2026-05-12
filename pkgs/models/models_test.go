package models

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestModelsEmbedGormModel(t *testing.T) {
	models := []interface{}{
		Article{},
		ArticleCategory{},
		Country{},
		Product{},
		ProductCategory{},
		ProductStorefront{},
		Role{},
		Storefront{},
		ArticleStorefront{},
		User{},
	}

	for _, model := range models {
		modelType := reflect.TypeOf(model)
		field, ok := modelType.FieldByName("Model")

		assert.True(t, ok, "%s should embed gorm.Model", modelType.Name())
		assert.Equal(t, reflect.TypeOf(gorm.Model{}), field.Type)
		assert.True(t, field.Anonymous, "%s gorm.Model should be embedded", modelType.Name())
	}
}

func TestPublicModelJSONFields(t *testing.T) {
	shipping := int64(499)
	productID := uint(7)

	tests := []struct {
		name     string
		model    interface{}
		expected map[string]interface{}
	}{
		{
			name:  "country",
			model: Country{Code: "uk", Name: "United Kingdom", Currency: "GBP"},
			expected: map[string]interface{}{
				"code":     "uk",
				"name":     "United Kingdom",
				"currency": "GBP",
			},
		},
		{
			name:     "product category",
			model:    ProductCategory{Name: "Laptops"},
			expected: map[string]interface{}{"name": "Laptops"},
		},
		{
			name:     "article category",
			model:    ArticleCategory{Name: "Guides"},
			expected: map[string]interface{}{"name": "Guides"},
		},
		{
			name:     "role",
			model:    Role{Name: "admin"},
			expected: map[string]interface{}{"name": "admin"},
		},
		{
			name: "storefront",
			model: Storefront{
				Name:             "Lead Storefront",
				Slug:             "lead-storefront",
				Domain:           "leadstorefront.com",
				Description:      "Hosted storefront",
				LogoURL:          "/logo.png",
				LogoWidthPx:      180,
				HeroTitle:        "Hero",
				HeroSubtitle:     "Subtitle",
				HeroImageURL:     "/hero.png",
				AboutTitle:       "About",
				AboutBody:        "About copy",
				IsActive:         true,
				PrimaryCountryID: 1,
				OwnerID:          &productID,
			},
			expected: map[string]interface{}{
				"name":               "Lead Storefront",
				"slug":               "lead-storefront",
				"domain":             "leadstorefront.com",
				"description":        "Hosted storefront",
				"logo_url":           "/logo.png",
				"logo_width_px":      float64(180),
				"hero_title":         "Hero",
				"hero_subtitle":      "Subtitle",
				"hero_image_url":     "/hero.png",
				"about_title":        "About",
				"about_body":         "About copy",
				"is_active":          true,
				"primary_country_id": float64(1),
				"owner_id":           float64(7),
			},
		},
		{
			name:  "user hides password",
			model: User{Email: "admin@example.com", Password: "secret", RoleID: 2},
			expected: map[string]interface{}{
				"email":   "admin@example.com",
				"role_id": float64(2),
			},
		},
		{
			name: "product",
			model: Product{
				Name:               "Laptop",
				Slug:               "laptop",
				Description:        "Fast laptop",
				Brand:              "Brand",
				ModelNumber:        "ABC",
				ImageURL:           "/img.jpg",
				ProductURL:         "https://example.com/product",
				AffiliateURL:       "https://example.com/affiliate",
				RetailerName:       "Retailer",
				RetailerURL:        "https://example.com",
				Source:             "manual",
				ExternalID:         "external-1",
				Currency:           "GBP",
				CurrentPriceCents:  10000,
				OriginalPriceCents: 12000,
				ShippingCostCents:  &shipping,
				DiscountPercent:    20,
				CouponCode:         "SAVE",
				DealScore:          90,
				Rating:             4.5,
				ReviewCount:        12,
				IsAvailable:        true,
				IsFeatured:         true,
				CountryID:          1,
				CategoryID:         2,
			},
			expected: map[string]interface{}{
				"name":                 "Laptop",
				"slug":                 "laptop",
				"description":          "Fast laptop",
				"brand":                "Brand",
				"model_number":         "ABC",
				"image_url":            "/img.jpg",
				"product_url":          "https://example.com/product",
				"affiliate_url":        "https://example.com/affiliate",
				"retailer_name":        "Retailer",
				"retailer_url":         "https://example.com",
				"source":               "manual",
				"external_id":          "external-1",
				"currency":             "GBP",
				"current_price_cents":  float64(10000),
				"original_price_cents": float64(12000),
				"shipping_cost_cents":  float64(499),
				"discount_percent":     float64(20),
				"coupon_code":          "SAVE",
				"deal_score":           float64(90),
				"rating":               4.5,
				"review_count":         float64(12),
				"is_available":         true,
				"is_featured":          true,
				"country_id":           float64(1),
				"category_id":          float64(2),
			},
		},
		{
			name: "article",
			model: Article{
				Author:            "Editor",
				Title:             "Guide",
				Slug:              "guide",
				Subtitle:          "Short",
				Body:              "Body",
				MainImage:         "main.jpg",
				ImageURL:          "/main.jpg",
				MetaTitle:         "Meta",
				MetaDescription:   "Description",
				MetaKeywords:      "keywords",
				CanonicalURL:      "https://example.com/guide",
				IsPublished:       true,
				ArticleCategoryID: 3,
				ProductID:         &productID,
			},
			expected: map[string]interface{}{
				"author":              "Editor",
				"title":               "Guide",
				"slug":                "guide",
				"subtitle":            "Short",
				"body":                "Body",
				"main_image":          "main.jpg",
				"image_url":           "/main.jpg",
				"meta_title":          "Meta",
				"meta_description":    "Description",
				"meta_keywords":       "keywords",
				"canonical_url":       "https://example.com/guide",
				"is_published":        true,
				"article_category_id": float64(3),
				"product_id":          float64(7),
			},
		},
		{
			name:  "product storefront",
			model: ProductStorefront{ProductID: 4, StorefrontID: 9},
			expected: map[string]interface{}{
				"product_id":    float64(4),
				"storefront_id": float64(9),
			},
		},
		{
			name:  "article storefront",
			model: ArticleStorefront{ArticleID: 5, StorefrontID: 9},
			expected: map[string]interface{}{
				"article_id":    float64(5),
				"storefront_id": float64(9),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.Marshal(tt.model)
			require.NoError(t, err)

			var actual map[string]interface{}
			require.NoError(t, json.Unmarshal(data, &actual))
			for key, value := range tt.expected {
				assert.Equal(t, value, actual[key], "json field %s", key)
			}
			assert.NotContains(t, actual, "password")
		})
	}
}
