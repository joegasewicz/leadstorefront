package routes

import (
	"leadstorefront/pkgs/middleware"
	"leadstorefront/pkgs/models"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

type Storefronts struct {
	API *APIClient
}

type publicLeadField struct {
	Label      string
	Name       string
	Type       string
	Options    []string
	IsRequired bool
}

func (storefronts *Storefronts) Get(c *gin.Context) {
	country := c.Param("country")
	if country == "" {
		country = middleware.DefaultCountryCode
	}
	if !middleware.IsSupportedCountryCode(country) {
		c.Redirect(http.StatusFound, "/"+middleware.DefaultCountryCode+"/storefronts/"+c.Param("id"))
		return
	}

	var response struct {
		Storefront models.Storefront `json:"storefront"`
	}
	if err := storefronts.API.Get(c, "/storefronts/"+c.Param("id"), &response); err != nil {
		c.String(http.StatusInternalServerError, "could not load storefront")
		return
	}

	storefront := response.Storefront
	normalizeStorefrontAssetURLs(&storefront)
	if redirectURL, ok := storefrontCustomDomainURL(c, storefront); ok {
		c.Redirect(http.StatusMovedPermanently, redirectURL)
		return
	}

	captureAttribution(c, storefront.ID)
	renderStorefront(c, storefronts.API, http.StatusOK, country, storefront)
}

func renderStorefront(c *gin.Context, api *APIClient, status int, country string, storefront models.Storefront) {
	storefrontPath := storefrontBasePath(c, country, storefront)
	design, _, designSections := storefrontDesignTemplateData(storefront)
	c.HTML(status, "storefront_show", gin.H{
		"Title":             storefront.Name + " | LeadStorefront",
		"Country":           country,
		"Storefront":        storefront,
		"StorefrontDesign":  design,
		"DesignSections":    designSections,
		"StorefrontPath":    storefrontPath,
		"UseStorefrontFont": true,
		"LeadFields":        publicLeadFields(c, api, storefront.ID),
		"LeadFormAction":    c.Request.URL.RequestURI(),
		"Flash":             middleware.PopFlash(c),
	})
}

func publicLeadFields(c *gin.Context, api *APIClient, storefrontID uint) []publicLeadField {
	var response struct {
		Fields []models.LeadFormField `json:"fields"`
	}
	if err := api.Get(c, "/storefronts/"+uintToString(storefrontID)+"/lead-form", &response); err != nil {
		return nil
	}
	fields := make([]publicLeadField, 0, len(response.Fields))
	for _, field := range response.Fields {
		fields = append(fields, publicLeadField{
			Label:      field.Label,
			Name:       field.Name,
			Type:       field.Type,
			Options:    leadFieldOptions(field.Options),
			IsRequired: field.IsRequired,
		})
	}
	return fields
}

func leadFieldOptions(raw string) []string {
	parts := strings.Split(raw, ",")
	options := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			options = append(options, part)
		}
	}
	return options
}

func storefrontCustomDomainURL(c *gin.Context, storefront models.Storefront) (string, bool) {
	if isLocalHost(requestHost(c.Request.Host)) {
		return "", false
	}
	domain := requestHost(storefront.Domain)
	if domain == "" || isPlatformHost(domain) || requestHost(c.Request.Host) == domain {
		return "", false
	}
	if !isPlatformHost(requestHost(c.Request.Host)) {
		return "", false
	}
	return "https://" + strings.TrimSuffix(domain, "/"), true
}

func isLocalHost(host string) bool {
	switch host {
	case "localhost", "127.0.0.1", "::1":
		return true
	default:
		return false
	}
}

func (storefronts *Storefronts) Post(c *gin.Context) {
	c.AbortWithStatus(http.StatusMethodNotAllowed)
}

func (storefronts *Storefronts) Put(c *gin.Context) {
	c.AbortWithStatus(http.StatusMethodNotAllowed)
}

func (storefronts *Storefronts) Delete(c *gin.Context) {
	c.AbortWithStatus(http.StatusMethodNotAllowed)
}
