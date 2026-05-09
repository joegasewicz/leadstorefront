package routes

import (
	"leadstorefront/pkgs/middleware"
	"leadstorefront/pkgs/models"
	"net/http"
	"net/url"

	"github.com/gin-gonic/gin"
)

type Products struct {
	API *APIClient
}

func (products *Products) Get(c *gin.Context) {
	if c.Param("slug") != "" {
		products.Show(c)
		return
	}
	products.Index(c)
}

func (products *Products) Index(c *gin.Context) {
	countryCode := c.Param("country")
	if !middleware.IsSupportedCountryCode(countryCode) {
		c.Redirect(http.StatusFound, "/"+middleware.DefaultCountryCode+"/products")
		return
	}

	var response struct {
		Products []models.Product `json:"products"`
	}
	if err := products.API.Get(c, "/"+countryCode+"/products", &response); err != nil {
		c.String(http.StatusInternalServerError, "could not load products")
		return
	}

	c.HTML(http.StatusOK, "products_index", gin.H{
		"Title":    "Products | LeadStorefront",
		"Country":  countryCode,
		"Products": response.Products,
	})
}

func (products *Products) Show(c *gin.Context) {
	countryCode := c.Param("country")
	if !middleware.IsSupportedCountryCode(countryCode) {
		c.Redirect(http.StatusFound, "/"+middleware.DefaultCountryCode+"/products/"+c.Param("slug"))
		return
	}

	var response struct {
		Product models.Product `json:"product"`
	}
	if err := products.API.Get(c, "/"+countryCode+"/products/"+url.PathEscape(c.Param("slug")), &response); err != nil {
		c.String(http.StatusInternalServerError, "could not load product")
		return
	}
	product := response.Product

	c.HTML(http.StatusOK, "product_show", gin.H{
		"Title":   product.Name + " | LeadStorefront",
		"Country": countryCode,
		"Product": product,
	})
}

func (products *Products) Post(c *gin.Context) {
	c.AbortWithStatus(http.StatusMethodNotAllowed)
}

func (products *Products) Put(c *gin.Context) {
	c.AbortWithStatus(http.StatusMethodNotAllowed)
}

func (products *Products) Delete(c *gin.Context) {
	c.AbortWithStatus(http.StatusMethodNotAllowed)
}
