package routes

import (
	"leadstorefront/pkgs/middleware"
	"leadstorefront/pkgs/models"
	"net/http"
	"net/url"

	"github.com/gin-gonic/gin"
)

type Articles struct {
	API *APIClient
}

func (articles *Articles) Get(c *gin.Context) {
	if c.Param("slug") != "" {
		articles.Show(c)
		return
	}
	articles.Index(c)
}

func (articles *Articles) Index(c *gin.Context) {
	country := c.Param("country")
	storefront, hasStorefront := currentStorefront(c, articles.API)
	if country == "" && hasStorefront {
		country = storefront.PrimaryCountry.Code
	}
	if country == "" {
		country = middleware.DefaultCountryCode
	}
	if !middleware.IsSupportedCountryCode(country) {
		if storefrontIDFromPath(c) != "" {
			c.Redirect(http.StatusFound, "/"+middleware.DefaultCountryCode+"/storefronts/"+storefrontIDFromPath(c)+"/articles")
			return
		}
		c.Redirect(http.StatusFound, "/"+middleware.DefaultCountryCode+"/articles")
		return
	}

	var response struct {
		Articles []models.Article `json:"articles"`
	}
	path := "/" + country + "/articles"
	articlesPath := "/" + country + "/articles"
	if hasStorefront {
		path += "?storefront_id=" + url.QueryEscape(uintToString(storefront.ID))
		articlesPath = storefrontBasePath(c, country, storefront) + "/articles"
	}
	if err := articles.API.Get(c, path, &response); err != nil {
		c.String(http.StatusInternalServerError, "could not load articles")
		return
	}
	design, _, _ := storefrontDesignTemplateData(storefront)

	c.HTML(http.StatusOK, "articles_index", gin.H{
		"Title":             "Articles | LeadStorefront",
		"Country":           country,
		"Articles":          response.Articles,
		"ArticlesPath":      articlesPath,
		"Storefront":        storefront,
		"StorefrontDesign":  design,
		"UseStorefrontFont": hasStorefront,
	})
}

func (articles *Articles) Show(c *gin.Context) {
	country := c.Param("country")
	storefront, hasStorefront := currentStorefront(c, articles.API)
	if country == "" && hasStorefront {
		country = storefront.PrimaryCountry.Code
	}
	if country == "" {
		country = middleware.DefaultCountryCode
	}
	if !middleware.IsSupportedCountryCode(country) {
		if storefrontIDFromPath(c) != "" {
			c.Redirect(http.StatusFound, "/"+middleware.DefaultCountryCode+"/storefronts/"+storefrontIDFromPath(c)+"/articles/"+c.Param("slug"))
			return
		}
		c.Redirect(http.StatusFound, "/"+middleware.DefaultCountryCode+"/articles/"+c.Param("slug"))
		return
	}

	var response struct {
		Article models.Article `json:"article"`
	}
	path := "/" + country + "/articles/" + url.PathEscape(c.Param("slug"))
	articlesPath := "/" + country + "/articles"
	if hasStorefront {
		path += "?storefront_id=" + url.QueryEscape(uintToString(storefront.ID))
		articlesPath = storefrontBasePath(c, country, storefront) + "/articles"
	}
	if err := articles.API.Get(c, path, &response); err != nil {
		c.String(http.StatusInternalServerError, "could not load article")
		return
	}
	article := response.Article

	title := article.Title + " | LeadStorefront"
	if article.MetaTitle != "" {
		title = article.MetaTitle
	}
	design, _, _ := storefrontDesignTemplateData(storefront)

	c.HTML(http.StatusOK, "article_show", gin.H{
		"Title":             title,
		"Country":           country,
		"Article":           article,
		"ArticlesPath":      articlesPath,
		"Storefront":        storefront,
		"StorefrontDesign":  design,
		"UseStorefrontFont": hasStorefront,
	})
}

func (articles *Articles) Post(c *gin.Context) {
	c.AbortWithStatus(http.StatusMethodNotAllowed)
}

func (articles *Articles) Put(c *gin.Context) {
	c.AbortWithStatus(http.StatusMethodNotAllowed)
}

func (articles *Articles) Delete(c *gin.Context) {
	c.AbortWithStatus(http.StatusMethodNotAllowed)
}
