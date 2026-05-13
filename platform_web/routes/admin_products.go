package routes

import (
	"leadstorefront/pkgs/middleware"
	"leadstorefront/pkgs/models"
	"leadstorefront/pkgs/utils"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	form_validator "github.com/joegasewicz/form-validator"
)

type AdminProducts struct {
	API        *APIClient
	FormFields []form_validator.Field
}

func (products *AdminProducts) Get(c *gin.Context) {
	switch {
	case strings.HasSuffix(c.FullPath(), "/create"):
		products.Create(c)
	case strings.HasSuffix(c.FullPath(), "/edit"):
		products.Edit(c)
	default:
		products.Index(c)
	}
}

func (products *AdminProducts) Index(c *gin.Context) {
	var response struct {
		Products   []models.Product `json:"products"`
		Pagination utils.Pagination `json:"pagination"`
	}
	page, limit := utils.GetPaginationQuery(c)
	if err := products.API.Get(c, "/admin/products?page="+page+"&limit="+limit, &response); err != nil {
		c.String(http.StatusInternalServerError, "could not load products")
		return
	}

	c.HTML(http.StatusOK, "admin_products_index", gin.H{
		"Title":        "Products",
		"Products":     response.Products,
		"Pagination":   response.Pagination,
		"Limit":        limit,
		"Flash":        middleware.PopFlash(c),
		"IsAdmin":      true,
		"IsSuper":      isCurrentSuper(c),
		"IsAdminRoute": true,
	})
}

func (products *AdminProducts) Create(c *gin.Context) {
	products.renderForm(c, http.StatusOK, "Create product", "/admin/products/create", models.Product{IsAvailable: true}, "")
}

func (products *AdminProducts) Post(c *gin.Context) {
	product, err := products.productFromRequest(c)
	if err != nil {
		products.renderForm(c, http.StatusBadRequest, "Create product", "/admin/products/create", product, err.Error())
		return
	}

	var response struct {
		Product models.Product `json:"product"`
	}
	if err := products.API.Post(c, "/admin/products/create", productPayload(product), &response); err != nil {
		products.renderForm(c, http.StatusBadRequest, "Create product", "/admin/products/create", product, "Could not create the product.")
		return
	}
	if err := products.assignStorefronts(c, response.Product.ID); err != nil {
		products.renderForm(c, http.StatusBadRequest, "Create product", "/admin/products/create", product, "Could not assign the product to the selected storefronts.")
		return
	}

	_ = middleware.SetFlash(c, "Product created.")
	c.Redirect(http.StatusFound, "/admin/products")
}

func (products *AdminProducts) Edit(c *gin.Context) {
	product, ok := products.find(c)
	if !ok {
		return
	}

	products.renderForm(c, http.StatusOK, "Edit product", "/admin/products/"+c.Param("id")+"/edit", product, "")
}

func (products *AdminProducts) Put(c *gin.Context) {
	product, err := products.productFromRequest(c)
	if err != nil {
		products.renderForm(c, http.StatusBadRequest, "Edit product", "/admin/products/"+c.Param("id")+"/edit", product, err.Error())
		return
	}

	id, ok := apiPathID(c.Param("id"))
	if !ok {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}
	if err := products.API.Put(c, "/admin/products/"+id+"/edit", productPayload(product), nil); err != nil {
		products.renderForm(c, http.StatusBadRequest, "Edit product", "/admin/products/"+c.Param("id")+"/edit", product, "Could not update the product.")
		return
	}
	if err := products.assignStorefronts(c, uintIDFromPath(id)); err != nil {
		products.renderForm(c, http.StatusBadRequest, "Edit product", "/admin/products/"+c.Param("id")+"/edit", product, "Could not assign the product to the selected storefronts.")
		return
	}

	_ = middleware.SetFlash(c, "Product updated.")
	c.Redirect(http.StatusFound, "/admin/products")
}

func (products *AdminProducts) Delete(c *gin.Context) {
	id, ok := apiPathID(c.Param("id"))
	if !ok {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}

	if err := products.API.Delete(c, "/admin/products/"+id+"/delete", nil); err != nil {
		_ = middleware.SetFlash(c, "Could not delete the product.")
		c.Redirect(http.StatusFound, "/admin/products")
		return
	}

	_ = middleware.SetFlash(c, "Product deleted.")
	c.Redirect(http.StatusFound, "/admin/products")
}

func (products *AdminProducts) renderForm(c *gin.Context, status int, title string, action string, product models.Product, message string) {
	var response struct {
		Countries   []models.Country         `json:"countries"`
		Categories  []models.ProductCategory `json:"categories"`
		Storefronts []models.Storefront      `json:"storefronts"`
	}
	_ = products.API.Get(c, "/admin/products/options", &response)

	var shippingCostCents int64
	if product.ShippingCostCents != nil {
		shippingCostCents = *product.ShippingCostCents
	}

	c.HTML(status, "admin_product_form", gin.H{
		"Title":             title,
		"Action":            action,
		"Product":           product,
		"ShippingCostCents": shippingCostCents,
		"StartsAt":          formatDateTimeLocal(product.StartsAt),
		"EndsAt":            formatDateTimeLocal(product.EndsAt),
		"LastCheckedAt":     formatDateTimeLocal(product.LastCheckedAt),
		"Countries":         response.Countries,
		"Categories":        response.Categories,
		"Storefronts":       response.Storefronts,
		"Error":             message,
		"IsAdmin":           true,
		"IsSuper":           isCurrentSuper(c),
		"IsAdminRoute":      true,
	})
}

func (products *AdminProducts) productFromRequest(c *gin.Context) (models.Product, error) {
	config := form_validator.Config{Fields: products.formFields()}
	if ok := form_validator.ValidateForm(c.Request, &config); !ok {
		return models.Product{}, formError("Check the required product fields.")
	}

	countryID, err := parseRequiredUint(c.PostForm("country_id"), "Select a country.")
	if err != nil {
		return models.Product{}, err
	}

	categoryID, err := parseRequiredUint(c.PostForm("category_id"), "Select a product category.")
	if err != nil {
		return models.Product{}, err
	}

	currentPriceCents, err := parseRequiredInt64(c.PostForm("current_price_cents"), "Current price is required.")
	if err != nil {
		return models.Product{}, err
	}

	originalPriceCents, err := parseOptionalInt64Value(c.PostForm("original_price_cents"))
	if err != nil {
		return models.Product{}, err
	}

	shippingCostCents, err := parseOptionalInt64(c.PostForm("shipping_cost_cents"))
	if err != nil {
		return models.Product{}, err
	}

	discountPercent, err := parseOptionalInt(c.PostForm("discount_percent"))
	if err != nil {
		return models.Product{}, err
	}

	dealScore, err := parseOptionalInt(c.PostForm("deal_score"))
	if err != nil {
		return models.Product{}, err
	}

	rating, err := parseOptionalFloat32(c.PostForm("rating"))
	if err != nil {
		return models.Product{}, err
	}

	reviewCount, err := parseOptionalInt(c.PostForm("review_count"))
	if err != nil {
		return models.Product{}, err
	}

	startsAt, err := parseOptionalDateTimeLocal(c.PostForm("starts_at"))
	if err != nil {
		return models.Product{}, err
	}

	endsAt, err := parseOptionalDateTimeLocal(c.PostForm("ends_at"))
	if err != nil {
		return models.Product{}, err
	}

	lastCheckedAt, err := parseOptionalDateTimeLocal(c.PostForm("last_checked_at"))
	if err != nil {
		return models.Product{}, err
	}

	name := strings.TrimSpace(c.PostForm("name"))
	if name == "" {
		return models.Product{}, formError("Name is required.")
	}

	productURL := strings.TrimSpace(c.PostForm("product_url"))
	if productURL == "" {
		return models.Product{}, formError("Product URL is required.")
	}

	retailerName := strings.TrimSpace(c.PostForm("retailer_name"))
	if retailerName == "" {
		return models.Product{}, formError("Retailer name is required.")
	}

	currency := strings.ToUpper(strings.TrimSpace(c.PostForm("currency")))
	if currency == "" {
		return models.Product{}, formError("Currency is required.")
	}

	return models.Product{
		Name:               name,
		Slug:               slugify(name),
		Description:        strings.TrimSpace(c.PostForm("description")),
		Brand:              strings.TrimSpace(c.PostForm("brand")),
		ModelNumber:        strings.TrimSpace(c.PostForm("model_number")),
		ImageURL:           strings.TrimSpace(c.PostForm("image_url")),
		ProductURL:         productURL,
		AffiliateURL:       strings.TrimSpace(c.PostForm("affiliate_url")),
		RetailerName:       retailerName,
		RetailerURL:        strings.TrimSpace(c.PostForm("retailer_url")),
		Source:             strings.TrimSpace(c.PostForm("source")),
		ExternalID:         strings.TrimSpace(c.PostForm("external_id")),
		Currency:           currency,
		CurrentPriceCents:  currentPriceCents,
		OriginalPriceCents: originalPriceCents,
		ShippingCostCents:  shippingCostCents,
		DiscountPercent:    discountPercent,
		CouponCode:         strings.TrimSpace(c.PostForm("coupon_code")),
		DealScore:          dealScore,
		Rating:             rating,
		ReviewCount:        reviewCount,
		IsAvailable:        c.PostForm("is_available") == "on",
		IsFeatured:         c.PostForm("is_featured") == "on",
		StartsAt:           startsAt,
		EndsAt:             endsAt,
		LastCheckedAt:      lastCheckedAt,
		CountryID:          countryID,
		CategoryID:         categoryID,
	}, nil
}

func (products *AdminProducts) find(c *gin.Context) (models.Product, bool) {
	id, ok := apiPathID(c.Param("id"))
	if !ok {
		c.AbortWithStatus(http.StatusNotFound)
		return models.Product{}, false
	}

	var response struct {
		Product models.Product `json:"product"`
	}
	if err := products.API.Get(c, "/admin/products/"+id, &response); err != nil {
		c.AbortWithStatus(http.StatusNotFound)
		return models.Product{}, false
	}

	return response.Product, true
}

func productPayload(product models.Product) map[string]interface{} {
	return map[string]interface{}{
		"name": product.Name, "slug": product.Slug, "description": product.Description,
		"brand": product.Brand, "model_number": product.ModelNumber, "image_url": product.ImageURL,
		"product_url": product.ProductURL, "affiliate_url": product.AffiliateURL,
		"retailer_name": product.RetailerName, "retailer_url": product.RetailerURL,
		"source": product.Source, "external_id": product.ExternalID, "currency": product.Currency,
		"current_price_cents": product.CurrentPriceCents, "original_price_cents": product.OriginalPriceCents,
		"shipping_cost_cents": int64PtrPayload(product.ShippingCostCents), "discount_percent": product.DiscountPercent,
		"coupon_code": product.CouponCode, "deal_score": product.DealScore, "rating": product.Rating,
		"review_count": product.ReviewCount, "is_available": product.IsAvailable, "is_featured": product.IsFeatured,
		"starts_at": product.StartsAt, "ends_at": product.EndsAt, "last_checked_at": product.LastCheckedAt,
		"country_id": product.CountryID, "category_id": product.CategoryID,
	}
}

func (products *AdminProducts) assignStorefronts(c *gin.Context, productID uint) error {
	if productID == 0 {
		return nil
	}
	for _, storefrontID := range c.PostFormArray("storefront_ids") {
		id, err := parseRequiredUint(storefrontID, "Select a valid storefront.")
		if err != nil {
			return err
		}
		if err := products.API.Post(c, "/admin/storefronts/"+uintToString(id)+"/products", map[string]interface{}{"product_id": productID}, nil); err != nil {
			return err
		}
	}
	return nil
}

func uintIDFromPath(id string) uint {
	parsed, err := strconv.Atoi(id)
	if err != nil || parsed <= 0 {
		return 0
	}
	return uint(parsed)
}

func (products *AdminProducts) formFields() []form_validator.Field {
	if products.FormFields != nil {
		return products.FormFields
	}
	return []form_validator.Field{
		{Name: "country_id", Validate: true, Type: "uint"},
		{Name: "category_id", Validate: true, Type: "uint"},
		{Name: "name", Validate: true, Type: "string"},
		{Name: "product_url", Validate: true, Type: "string"},
		{Name: "retailer_name", Validate: true, Type: "string"},
		{Name: "currency", Validate: true, Type: "string"},
		{Name: "current_price_cents", Validate: true, Type: "int64"},
		{Name: "original_price_cents", Validate: false, Type: "string"},
		{Name: "shipping_cost_cents", Validate: false, Type: "string"},
		{Name: "discount_percent", Validate: false, Type: "string"},
		{Name: "deal_score", Validate: false, Type: "string"},
		{Name: "rating", Validate: false, Type: "string"},
		{Name: "review_count", Validate: false, Type: "string"},
	}
}

func parseRequiredInt64(value string, message string) (int64, error) {
	parsed, err := parseOptionalInt64Value(value)
	if err != nil || parsed <= 0 {
		return 0, formError(message)
	}
	return parsed, nil
}

func parseOptionalInt64(value string) (*int64, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil, nil
	}

	parsed, err := parseOptionalInt64Value(value)
	if err != nil {
		return nil, err
	}

	return &parsed, nil
}

func parseOptionalInt64Value(value string) (int64, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return 0, nil
	}

	parsed, err := strconv.ParseInt(value, 10, 64)
	if err != nil || parsed < 0 {
		return 0, formError("Enter a valid non-negative whole number.")
	}

	return parsed, nil
}

func parseOptionalInt(value string) (int, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return 0, nil
	}

	parsed, err := strconv.Atoi(value)
	if err != nil || parsed < 0 {
		return 0, formError("Enter a valid non-negative whole number.")
	}

	return parsed, nil
}

func parseOptionalFloat32(value string) (float32, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return 0, nil
	}

	parsed, err := strconv.ParseFloat(value, 32)
	if err != nil || parsed < 0 {
		return 0, formError("Enter a valid non-negative decimal number.")
	}

	return float32(parsed), nil
}

func parseOptionalDateTimeLocal(value string) (*time.Time, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil, nil
	}

	parsed, err := time.ParseInLocation("2006-01-02T15:04", value, time.Local)
	if err != nil {
		return nil, formError("Enter a valid date and time.")
	}

	return &parsed, nil
}

func formatDateTimeLocal(value *time.Time) string {
	if value == nil {
		return ""
	}

	return value.Format("2006-01-02T15:04")
}
