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
	storefront, hasStorefront := currentStorefront(c, products.API)
	if countryCode == "" && hasStorefront {
		countryCode = storefront.PrimaryCountry.Code
	}
	if countryCode == "" {
		countryCode = middleware.DefaultCountryCode
	}
	if !middleware.IsSupportedCountryCode(countryCode) {
		if storefrontIDFromPath(c) != "" {
			c.Redirect(http.StatusFound, "/"+middleware.DefaultCountryCode+"/storefronts/"+storefrontIDFromPath(c)+"/products")
			return
		}
		c.Redirect(http.StatusFound, "/"+middleware.DefaultCountryCode+"/products")
		return
	}

	var response struct {
		Products []models.Product `json:"products"`
	}
	path := "/" + countryCode + "/products"
	productsPath := "/" + countryCode + "/products"
	if hasStorefront {
		path += "?storefront_id=" + url.QueryEscape(uintToString(storefront.ID))
		productsPath = storefrontBasePath(c, countryCode, storefront) + "/products"
	}
	if err := products.API.Get(c, path, &response); err != nil {
		c.String(http.StatusInternalServerError, "could not load products")
		return
	}

	c.HTML(http.StatusOK, "products_index", gin.H{
		"Title":        "Products | LeadStorefront",
		"Country":      countryCode,
		"Products":     response.Products,
		"ProductsPath": productsPath,
		"Storefront":   storefront,
	})
}

func (products *Products) Show(c *gin.Context) {
	countryCode := c.Param("country")
	storefront, hasStorefront := currentStorefront(c, products.API)
	if countryCode == "" && hasStorefront {
		countryCode = storefront.PrimaryCountry.Code
	}
	if countryCode == "" {
		countryCode = middleware.DefaultCountryCode
	}
	if !middleware.IsSupportedCountryCode(countryCode) {
		if storefrontIDFromPath(c) != "" {
			c.Redirect(http.StatusFound, "/"+middleware.DefaultCountryCode+"/storefronts/"+storefrontIDFromPath(c)+"/products/"+c.Param("slug"))
			return
		}
		c.Redirect(http.StatusFound, "/"+middleware.DefaultCountryCode+"/products/"+c.Param("slug"))
		return
	}

	var response struct {
		Product models.Product `json:"product"`
	}
	path := "/" + countryCode + "/products/" + url.PathEscape(c.Param("slug"))
	productsPath := "/" + countryCode + "/products"
	if hasStorefront {
		path += "?storefront_id=" + url.QueryEscape(uintToString(storefront.ID))
		productsPath = storefrontBasePath(c, countryCode, storefront) + "/products"
	}
	if err := products.API.Get(c, path, &response); err != nil {
		c.String(http.StatusInternalServerError, "could not load product")
		return
	}
	product := response.Product

	c.HTML(http.StatusOK, "product_show", gin.H{
		"Title":        product.Name + " | LeadStorefront",
		"Country":      countryCode,
		"Product":      product,
		"ProductsPath": productsPath,
		"Storefront":   storefront,
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
