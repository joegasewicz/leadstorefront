package routes

import (
	"leadstorefront/pkgs/middleware"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

type Home struct {
	API *APIClient
}

func (home *Home) Redirect(c *gin.Context) {
	if storefront, ok := currentStorefront(c, home.API); ok {
		country := storefront.PrimaryCountry.Code
		if !middleware.IsSupportedCountryCode(country) {
			country = middleware.DefaultCountryCode
		}
		renderStorefront(c, http.StatusOK, country, storefront)
		return
	}
	middleware.RedirectToLocalizedHome()(c)
}

func (home *Home) Get(c *gin.Context) {
	countryCode := c.Param("country")
	if !middleware.IsSupportedCountryCode(countryCode) {
		c.Redirect(http.StatusFound, "/"+middleware.DefaultCountryCode)
		return
	}
	if storefront, ok := currentStorefront(c, home.API); ok {
		renderStorefront(c, http.StatusOK, countryCode, storefront)
		return
	}

	c.HTML(http.StatusOK, "home", gin.H{
		"Title":   "LeadStorefront",
		"Country": countryCode,
	})
}

func (home *Home) Post(c *gin.Context) {
	c.AbortWithStatus(http.StatusMethodNotAllowed)
}

func (home *Home) Put(c *gin.Context) {
	c.AbortWithStatus(http.StatusMethodNotAllowed)
}

func (home *Home) Delete(c *gin.Context) {
	c.AbortWithStatus(http.StatusMethodNotAllowed)
}

func requestHost(host string) string {
	host = strings.ToLower(strings.TrimSpace(host))
	if colon := strings.Index(host, ":"); colon >= 0 {
		return host[:colon]
	}
	return host
}

func isPlatformHost(host string) bool {
	switch host {
	case "", "localhost", "127.0.0.1", "leadstorefront.com", "www.leadstorefront.com":
		return true
	default:
		return false
	}
}
