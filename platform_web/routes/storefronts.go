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

	renderStorefront(c, http.StatusOK, country, storefront)
}

func renderStorefront(c *gin.Context, status int, country string, storefront models.Storefront) {
	storefrontPath := storefrontBasePath(c, country, storefront)
	c.HTML(status, "storefront_show", gin.H{
		"Title":          storefront.Name + " | LeadStorefront",
		"Country":        country,
		"Storefront":     storefront,
		"StorefrontPath": storefrontPath,
	})
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
