package routes

import (
	"leadstorefront/pkgs/middleware"
	"leadstorefront/pkgs/models"
	"net/http"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"
)

type Home struct {
	API *APIClient
}

func (home *Home) Redirect(c *gin.Context) {
	if redirectPath, ok := home.customDomainRedirect(c); ok {
		c.Redirect(http.StatusFound, redirectPath)
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

func (home *Home) customDomainRedirect(c *gin.Context) (string, bool) {
	host := requestHost(c.Request.Host)
	if isPlatformHost(host) {
		return "", false
	}

	var response struct {
		Storefront models.Storefront `json:"storefront"`
	}
	if err := home.API.Get(c, "/storefront-domains/"+url.PathEscape(host), &response); err != nil {
		return "", false
	}

	country := response.Storefront.PrimaryCountry.Code
	if !middleware.IsSupportedCountryCode(country) {
		country = middleware.DefaultCountryCode
	}
	return "/" + country + "/storefronts/" + response.Storefront.Slug, true
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
