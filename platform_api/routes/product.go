package routes

import (
	"leadstorefront/pkgs/models"
	"leadstorefront/pkgs/utils"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Product struct {
	DB *gorm.DB
}

func (product *Product) Get(c *gin.Context) {
	if c.Param("id") != "" {
		product.getByID(c)
		return
	}
	if c.Param("slug") != "" {
		product.getByCountrySlug(c)
		return
	}
	if utils.CountryCodeFromRequest(c) != "" && !strings.HasPrefix(c.FullPath(), utils.APIVersion+"/admin/") {
		product.getByCountry(c)
		return
	}
	product.getAdminList(c)
}

func (product *Product) Post(c *gin.Context) {
	record, ok := product.bindJSON(c)
	if !ok {
		return
	}
	if err := product.DB.Create(&record).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not create product"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"product": record})
}

func (product *Product) Put(c *gin.Context) {
	record, ok := product.bindJSON(c)
	if !ok {
		return
	}
	var existing models.Product
	if err := product.DB.First(&existing, c.Param("id")).Error; err != nil {
		utils.WriteRecordError(c, err, "could not load product")
		return
	}
	if err := product.DB.Model(&existing).Updates(productUpdateMap(record)).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not update product"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"product": existing})
}

func (product *Product) Delete(c *gin.Context) {
	if err := product.DB.Delete(&models.Product{}, c.Param("id")).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not delete product"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"deleted": true})
}

func (product *Product) getByCountry(c *gin.Context) {
	country := Country{DB: product.DB}
	record, ok := country.byCode(c)
	if !ok {
		return
	}
	var products []models.Product
	query := product.DB.Preload("Country").Preload("Category").Where("country_id = ?", record.ID).Order("is_featured desc, deal_score desc, created_at desc")
	if storefrontID := strings.TrimSpace(c.Query("storefront_id")); storefrontID != "" {
		query = query.Joins("JOIN product_storefronts ON product_storefronts.product_id = products.id").Where("product_storefronts.storefront_id = ?", storefrontID)
	}
	if err := query.Find(&products).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not load products"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"products": products})
}

func (product *Product) getByCountrySlug(c *gin.Context) {
	var record models.Product
	query := product.DB.Preload("Country").Preload("Category").Joins("JOIN countries ON countries.id = products.country_id").Where("countries.code = ? AND products.slug = ?", strings.ToLower(utils.CountryCodeFromRequest(c)), c.Param("slug"))
	if storefrontID := strings.TrimSpace(c.Query("storefront_id")); storefrontID != "" {
		query = query.Joins("JOIN product_storefronts ON product_storefronts.product_id = products.id").Where("product_storefronts.storefront_id = ?", storefrontID)
	}
	err := query.First(&record).Error
	if err != nil {
		utils.WriteRecordError(c, err, "could not load product")
		return
	}
	response := gin.H{"product": record}
	if storefront := product.associatedStorefront(record.ID, strings.TrimSpace(c.Query("storefront_id"))); storefront.ID != 0 {
		response["storefront"] = storefront
	}
	c.JSON(http.StatusOK, response)
}

func (product *Product) associatedStorefront(productID uint, storefrontID string) models.Storefront {
	var association models.ProductStorefront
	query := product.DB.Preload("Storefront").Preload("Storefront.PrimaryCountry").Joins("JOIN storefronts ON storefronts.id = product_storefronts.storefront_id").Where("product_storefronts.product_id = ? AND storefronts.is_active = ?", productID, true)
	if storefrontID != "" {
		query = query.Where("product_storefronts.storefront_id = ?", storefrontID)
	}
	if err := query.Order("product_storefronts.created_at asc").First(&association).Error; err != nil {
		return models.Storefront{}
	}
	return association.Storefront
}

func (product *Product) getAdminList(c *gin.Context) {
	var products []models.Product
	var total int64
	page, limit, offset := utils.GetPagination(c)
	query := product.DB.Model(&models.Product{})
	if storefrontID := strings.TrimSpace(c.Query("storefront_id")); storefrontID != "" {
		query = query.Joins("JOIN product_storefronts ON product_storefronts.product_id = products.id").Where("product_storefronts.storefront_id = ?", storefrontID)
	}
	if err := query.Count(&total).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not count products"})
		return
	}
	listQuery := product.DB.Preload("Country").Preload("Category").Order("created_at desc").Limit(limit).Offset(offset)
	if storefrontID := strings.TrimSpace(c.Query("storefront_id")); storefrontID != "" {
		listQuery = listQuery.Joins("JOIN product_storefronts ON product_storefronts.product_id = products.id").Where("product_storefronts.storefront_id = ?", storefrontID)
	}
	if err := listQuery.Find(&products).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not load products"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"products": products, "pagination": utils.NewPagination(page, limit, total)})
}

func (product *Product) getByID(c *gin.Context) {
	var record models.Product
	err := product.DB.First(&record, c.Param("id")).Error
	if err != nil {
		utils.WriteRecordError(c, err, "could not load product")
		return
	}
	c.JSON(http.StatusOK, gin.H{"product": record})
}

func (product *Product) bindJSON(c *gin.Context) (models.Product, bool) {
	var record models.Product
	if err := c.ShouldBindJSON(&record); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid product"})
		return models.Product{}, false
	}
	record.Slug = utils.Slugify(record.Name)
	return record, true
}

func productUpdateMap(product models.Product) map[string]interface{} {
	return map[string]interface{}{
		"name": product.Name, "slug": product.Slug, "description": product.Description,
		"brand": product.Brand, "model_number": product.ModelNumber, "image_url": product.ImageURL,
		"product_url": product.ProductURL, "affiliate_url": product.AffiliateURL,
		"retailer_name": product.RetailerName, "retailer_url": product.RetailerURL,
		"source": product.Source, "external_id": product.ExternalID, "currency": product.Currency,
		"current_price_cents": product.CurrentPriceCents, "original_price_cents": product.OriginalPriceCents,
		"shipping_cost_cents": product.ShippingCostCents, "discount_percent": product.DiscountPercent,
		"coupon_code": product.CouponCode, "deal_score": product.DealScore, "rating": product.Rating,
		"review_count": product.ReviewCount, "is_available": product.IsAvailable, "is_featured": product.IsFeatured,
		"starts_at": product.StartsAt, "ends_at": product.EndsAt, "last_checked_at": product.LastCheckedAt,
		"country_id": product.CountryID, "category_id": product.CategoryID,
	}
}
