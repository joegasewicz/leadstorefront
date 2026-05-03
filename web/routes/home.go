package routes

import (
	"gadgetscout/pkgs/middleware"
	"gadgetscout/pkgs/models"
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

	var response struct {
		LatestDeals    []models.Product `json:"latest_deals"`
		LatestArticles []models.Article `json:"latest_articles"`
	}
	if err := home.API.Get(c, "/"+countryCode, &response); err != nil {
		c.String(http.StatusInternalServerError, "could not load latest deals")
		return
	}

	c.HTML(http.StatusOK, "home", gin.H{
		"Title":          "Gadget Scout",
		"Country":        countryCode,
		"LatestDeals":    response.LatestDeals,
		"LatestArticles": response.LatestArticles,
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
