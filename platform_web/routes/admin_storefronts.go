package routes

import (
	"leadstorefront/pkgs/middleware"
	"leadstorefront/pkgs/models"
	"leadstorefront/pkgs/utils"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	form_validator "github.com/joegasewicz/form-validator"
)

type AdminStorefronts struct {
	API        *APIClient
	FormFields []form_validator.Field
}

func (storefronts *AdminStorefronts) Get(c *gin.Context) {
	if strings.HasSuffix(c.FullPath(), "/create") {
		storefronts.Create(c)
		return
	}
	if strings.HasSuffix(c.FullPath(), "/delete") {
		storefronts.DeleteForm(c)
		return
	}
	if c.Param("id") != "" {
		storefronts.Show(c)
		return
	}
	storefronts.Index(c)
}

func (storefronts *AdminStorefronts) Post(c *gin.Context) {
	if strings.Contains(c.FullPath(), "/products") {
		storefronts.AssignProduct(c)
		return
	}
	if strings.Contains(c.FullPath(), "/articles") {
		storefronts.AssignArticle(c)
		return
	}
	storefronts.CreatePost(c)
}

func (storefronts *AdminStorefronts) Index(c *gin.Context) {
	var response struct {
		Storefronts []models.Storefront `json:"storefronts"`
		Pagination  utils.Pagination    `json:"pagination"`
	}
	page, limit := utils.GetPaginationQuery(c)
	if err := storefronts.API.Get(c, "/admin/storefronts?page="+page+"&limit="+limit, &response); err != nil {
		c.String(http.StatusInternalServerError, "could not load storefronts")
		return
	}

	c.HTML(http.StatusOK, "admin_storefronts_index", gin.H{
		"Title":        "Storefronts",
		"Storefronts":  response.Storefronts,
		"Pagination":   response.Pagination,
		"Limit":        limit,
		"Flash":        middleware.PopFlash(c),
		"IsAdmin":      true,
		"IsAdminRoute": true,
	})
}

func (storefronts *AdminStorefronts) Create(c *gin.Context) {
	storefronts.renderForm(c, http.StatusOK, "Create storefront", "/admin/storefronts/create", models.Storefront{IsActive: true}, "")
}

func (storefronts *AdminStorefronts) Show(c *gin.Context) {
	id, ok := apiPathID(c.Param("id"))
	if !ok {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}

	var storefrontResponse struct {
		Storefront models.Storefront `json:"storefront"`
	}
	if err := storefronts.API.Get(c, "/admin/storefronts/"+id, &storefrontResponse); err != nil {
		c.String(http.StatusNotFound, "could not load storefront")
		return
	}

	var productResponse struct {
		Products   []models.Product `json:"products"`
		Pagination utils.Pagination `json:"pagination"`
	}
	_ = storefronts.API.Get(c, "/admin/products?storefront_id="+id+"&limit=100", &productResponse)
	var availableProductResponse struct {
		Products []models.Product `json:"products"`
	}
	_ = storefronts.API.Get(c, "/admin/products?limit=100", &availableProductResponse)

	var articleResponse struct {
		Articles   []models.Article `json:"articles"`
		Pagination utils.Pagination `json:"pagination"`
	}
	_ = storefronts.API.Get(c, "/admin/articles?storefront_id="+id+"&limit=100", &articleResponse)
	var availableArticleResponse struct {
		Articles []models.Article `json:"articles"`
	}
	_ = storefronts.API.Get(c, "/admin/articles?limit=100", &availableArticleResponse)

	c.HTML(http.StatusOK, "admin_storefront_show", gin.H{
		"Title":             storefrontResponse.Storefront.Name,
		"Storefront":        storefrontResponse.Storefront,
		"Products":          productResponse.Products,
		"Articles":          articleResponse.Articles,
		"AvailableProducts": unassignedProducts(availableProductResponse.Products, productResponse.Products),
		"AvailableArticles": unassignedArticles(availableArticleResponse.Articles, articleResponse.Articles),
		"Flash":             middleware.PopFlash(c),
		"IsAdmin":           true,
		"IsAdminRoute":      true,
	})
}

func (storefronts *AdminStorefronts) CreatePost(c *gin.Context) {
	storefront, err := storefronts.storefrontFromRequest(c)
	if err != nil {
		storefronts.renderForm(c, http.StatusBadRequest, "Create storefront", "/admin/storefronts/create", storefront, err.Error())
		return
	}

	var response struct {
		Storefront models.Storefront `json:"storefront"`
	}
	if err := storefronts.API.Post(c, "/admin/storefronts/create", storefrontPayload(storefront), &response); err != nil {
		storefronts.renderForm(c, http.StatusBadRequest, "Create storefront", "/admin/storefronts/create", storefront, "Could not create the storefront.")
		return
	}

	_ = middleware.SetFlash(c, "Storefront created.")
	c.Redirect(http.StatusFound, "/admin/storefronts")
}

func (storefronts *AdminStorefronts) AssignProduct(c *gin.Context) {
	id, ok := apiPathID(c.Param("id"))
	if !ok {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}
	productID, err := parseRequiredUint(c.PostForm("product_id"), "Select a product.")
	if err != nil {
		_ = middleware.SetFlash(c, err.Error())
		c.Redirect(http.StatusFound, "/admin/storefronts/"+id)
		return
	}
	if err := storefronts.API.Post(c, "/admin/storefronts/"+id+"/products", map[string]interface{}{"product_id": productID}, nil); err != nil {
		_ = middleware.SetFlash(c, "Could not assign the product.")
		c.Redirect(http.StatusFound, "/admin/storefronts/"+id)
		return
	}
	_ = middleware.SetFlash(c, "Product assigned.")
	c.Redirect(http.StatusFound, "/admin/storefronts/"+id)
}

func (storefronts *AdminStorefronts) AssignArticle(c *gin.Context) {
	id, ok := apiPathID(c.Param("id"))
	if !ok {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}
	articleID, err := parseRequiredUint(c.PostForm("article_id"), "Select an article.")
	if err != nil {
		_ = middleware.SetFlash(c, err.Error())
		c.Redirect(http.StatusFound, "/admin/storefronts/"+id)
		return
	}
	if err := storefronts.API.Post(c, "/admin/storefronts/"+id+"/articles", map[string]interface{}{"article_id": articleID}, nil); err != nil {
		_ = middleware.SetFlash(c, "Could not assign the article.")
		c.Redirect(http.StatusFound, "/admin/storefronts/"+id)
		return
	}
	_ = middleware.SetFlash(c, "Article assigned.")
	c.Redirect(http.StatusFound, "/admin/storefronts/"+id)
}

func (storefronts *AdminStorefronts) Put(c *gin.Context) {
	c.AbortWithStatus(http.StatusMethodNotAllowed)
}

func (storefronts *AdminStorefronts) Delete(c *gin.Context) {
	id, ok := apiPathID(c.Param("id"))
	if !ok {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}

	storefront, ok := storefronts.find(c, id)
	if !ok {
		return
	}

	if c.PostForm("confirm_delete") != "on" {
		storefronts.renderDelete(c, http.StatusBadRequest, storefront, "Confirm that you want to delete this storefront.")
		return
	}
	if strings.TrimSpace(c.PostForm("domain")) != storefront.Domain {
		storefronts.renderDelete(c, http.StatusBadRequest, storefront, "Type the storefront domain exactly as shown.")
		return
	}

	if err := storefronts.API.Delete(c, "/admin/storefronts/"+id+"/delete", nil); err != nil {
		storefronts.renderDelete(c, http.StatusBadRequest, storefront, "Could not delete the storefront.")
		return
	}

	_ = middleware.SetFlash(c, "Storefront deleted.")
	c.Redirect(http.StatusFound, "/admin/storefronts")
}

func (storefronts *AdminStorefronts) DeleteForm(c *gin.Context) {
	id, ok := apiPathID(c.Param("id"))
	if !ok {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}

	storefront, ok := storefronts.find(c, id)
	if !ok {
		return
	}

	storefronts.renderDelete(c, http.StatusOK, storefront, "")
}

func (storefronts *AdminStorefronts) find(c *gin.Context, id string) (models.Storefront, bool) {
	var response struct {
		Storefront models.Storefront `json:"storefront"`
	}
	if err := storefronts.API.Get(c, "/admin/storefronts/"+id, &response); err != nil {
		c.String(http.StatusNotFound, "could not load storefront")
		return models.Storefront{}, false
	}
	return response.Storefront, true
}

func (storefronts *AdminStorefronts) renderDelete(c *gin.Context, status int, storefront models.Storefront, message string) {
	c.HTML(status, "admin_storefront_delete", gin.H{
		"Title":        "Delete storefront",
		"Storefront":   storefront,
		"Error":        message,
		"IsAdmin":      true,
		"IsAdminRoute": true,
	})
}

func (storefronts *AdminStorefronts) renderForm(c *gin.Context, status int, title string, action string, storefront models.Storefront, message string) {
	var response struct {
		Countries []models.Country `json:"countries"`
		Users     []models.User    `json:"users"`
	}
	_ = storefronts.API.Get(c, "/admin/storefronts/create", &response)

	var ownerID uint
	if storefront.OwnerID != nil {
		ownerID = *storefront.OwnerID
	}

	c.HTML(status, "admin_storefront_form", gin.H{
		"Title":        title,
		"Action":       action,
		"Storefront":   storefront,
		"Countries":    response.Countries,
		"Users":        response.Users,
		"OwnerID":      ownerID,
		"Error":        message,
		"IsAdmin":      true,
		"IsAdminRoute": true,
	})
}

func (storefronts *AdminStorefronts) storefrontFromRequest(c *gin.Context) (models.Storefront, error) {
	config := form_validator.Config{Fields: storefronts.formFields()}
	if ok := form_validator.ValidateForm(c.Request, &config); !ok {
		return models.Storefront{}, formError("Check the required storefront fields.")
	}

	countryID, err := parseRequiredUint(c.PostForm("primary_country_id"), "Select a primary country.")
	if err != nil {
		return models.Storefront{}, err
	}

	ownerID, err := parseOptionalUint(c.PostForm("owner_id"))
	if err != nil {
		return models.Storefront{}, err
	}

	name := strings.TrimSpace(c.PostForm("name"))
	if name == "" {
		return models.Storefront{}, formError("Name is required.")
	}

	domain := strings.ToLower(strings.TrimSpace(c.PostForm("domain")))
	if domain == "" {
		return models.Storefront{}, formError("Domain is required.")
	}

	return models.Storefront{
		Name:             name,
		Slug:             slugify(name),
		Domain:           domain,
		Description:      strings.TrimSpace(c.PostForm("description")),
		LogoURL:          strings.TrimSpace(c.PostForm("logo_url")),
		IsActive:         c.PostForm("is_active") == "on",
		PrimaryCountryID: countryID,
		OwnerID:          ownerID,
	}, nil
}

func storefrontPayload(storefront models.Storefront) map[string]interface{} {
	return map[string]interface{}{
		"name":               storefront.Name,
		"slug":               storefront.Slug,
		"domain":             storefront.Domain,
		"description":        storefront.Description,
		"logo_url":           storefront.LogoURL,
		"is_active":          storefront.IsActive,
		"primary_country_id": storefront.PrimaryCountryID,
		"owner_id":           uintPtrPayload(storefront.OwnerID),
	}
}

func unassignedProducts(all []models.Product, assigned []models.Product) []models.Product {
	assignedIDs := make(map[uint]bool, len(assigned))
	for _, product := range assigned {
		assignedIDs[product.ID] = true
	}

	var available []models.Product
	for _, product := range all {
		if !assignedIDs[product.ID] {
			available = append(available, product)
		}
	}
	return available
}

func unassignedArticles(all []models.Article, assigned []models.Article) []models.Article {
	assignedIDs := make(map[uint]bool, len(assigned))
	for _, article := range assigned {
		assignedIDs[article.ID] = true
	}

	var available []models.Article
	for _, article := range all {
		if !assignedIDs[article.ID] {
			available = append(available, article)
		}
	}
	return available
}

func (storefronts *AdminStorefronts) formFields() []form_validator.Field {
	if storefronts.FormFields != nil {
		return storefronts.FormFields
	}
	return []form_validator.Field{
		{Name: "name", Validate: true, Type: "string"},
		{Name: "domain", Validate: true, Type: "string"},
		{Name: "primary_country_id", Validate: true, Type: "uint"},
		{Name: "owner_id", Validate: false, Type: "string"},
	}
}
