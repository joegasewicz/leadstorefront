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
	storefronts.Index(c)
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

func (storefronts *AdminStorefronts) Post(c *gin.Context) {
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

func (storefronts *AdminStorefronts) Put(c *gin.Context) {
	c.AbortWithStatus(http.StatusMethodNotAllowed)
}

func (storefronts *AdminStorefronts) Delete(c *gin.Context) {
	c.AbortWithStatus(http.StatusMethodNotAllowed)
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
