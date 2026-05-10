package routes

import (
	"net/http"
	"net/url"

	"github.com/gin-gonic/gin"
)

type DomainCheck struct {
	API *APIClient
}

func (check *DomainCheck) Get(c *gin.Context) {
	domain := requestHost(c.Query("domain"))
	if domain == "" || isPlatformHost(domain) {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}

	var response struct {
		Storefront struct {
			ID uint `json:"id"`
		} `json:"storefront"`
	}
	if err := check.API.Get(c, "/storefront-domains/"+url.PathEscape(domain), &response); err != nil || response.Storefront.ID == 0 {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}

	c.Status(http.StatusOK)
}

func (check *DomainCheck) Post(c *gin.Context) {
	c.AbortWithStatus(http.StatusMethodNotAllowed)
}

func (check *DomainCheck) Put(c *gin.Context) {
	c.AbortWithStatus(http.StatusMethodNotAllowed)
}

func (check *DomainCheck) Delete(c *gin.Context) {
	c.AbortWithStatus(http.StatusMethodNotAllowed)
}
