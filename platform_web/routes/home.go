package routes

import (
	"leadstorefront/pkgs/middleware"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Home struct {
	API *APIClient
}

func (home *Home) Redirect(c *gin.Context) {
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
