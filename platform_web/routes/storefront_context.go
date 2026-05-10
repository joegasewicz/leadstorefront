package routes

import (
	"leadstorefront/pkgs/models"
	"net/url"

	"github.com/gin-gonic/gin"
)

func currentStorefront(c *gin.Context, api *APIClient) (models.Storefront, bool) {
	host := requestHost(c.Request.Host)
	if isPlatformHost(host) {
		return models.Storefront{}, false
	}

	var response struct {
		Storefront models.Storefront `json:"storefront"`
	}
	if err := api.Get(c, "/storefront-domains/"+url.PathEscape(host), &response); err != nil {
		return models.Storefront{}, false
	}
	if response.Storefront.ID == 0 {
		return models.Storefront{}, false
	}
	return response.Storefront, true
}
