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
	if !middleware.IsSupportedCountryCode(country) {
		c.Redirect(http.StatusFound, "/"+middleware.DefaultCountryCode+"/articles")
		return
	}

	var response struct {
		Articles []models.Article `json:"articles"`
	}
	if err := articles.API.Get(c, "/"+country+"/articles", &response); err != nil {
		c.String(http.StatusInternalServerError, "could not load articles")
		return
	}

	c.HTML(http.StatusOK, "articles_index", gin.H{
		"Title":    "Articles | LeadStorefront",
		"Country":  country,
		"Articles": response.Articles,
	})
}

func (articles *Articles) Show(c *gin.Context) {
	country := c.Param("country")
	if !middleware.IsSupportedCountryCode(country) {
		c.Redirect(http.StatusFound, "/"+middleware.DefaultCountryCode+"/articles/"+c.Param("slug"))
		return
	}

	var response struct {
		Article models.Article `json:"article"`
	}
	if err := articles.API.Get(c, "/"+country+"/articles/"+url.PathEscape(c.Param("slug")), &response); err != nil {
		c.String(http.StatusInternalServerError, "could not load article")
		return
	}
	article := response.Article

	title := article.Title + " | LeadStorefront"
	if article.MetaTitle != "" {
		title = article.MetaTitle
	}

	c.HTML(http.StatusOK, "article_show", gin.H{
		"Title":   title,
		"Country": country,
		"Article": article,
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
