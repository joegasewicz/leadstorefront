package routes

import (
	"leadstorefront/pkgs/models"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"
)

func currentStorefront(c *gin.Context, api *APIClient) (models.Storefront, bool) {
	if id := storefrontIDFromPath(c); id != "" {
		var response struct {
			Storefront models.Storefront `json:"storefront"`
		}
		if err := api.Get(c, "/storefronts/"+url.PathEscape(id), &response); err != nil {
			return models.Storefront{}, false
		}
		if response.Storefront.ID == 0 {
			return models.Storefront{}, false
		}
		return response.Storefront, true
	}

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

func storefrontIDFromPath(c *gin.Context) string {
	if !strings.Contains(c.FullPath(), "/storefronts/:id") {
		return ""
	}
	return c.Param("id")
}

func storefrontBasePath(c *gin.Context, country string, storefront models.Storefront) string {
	id := uintToString(storefront.ID)
	if pathID := storefrontIDFromPath(c); pathID != "" {
		id = pathID
	}
	if c.Param("country") == "" {
		return "/storefronts/" + id
	}
	return "/" + country + "/storefronts/" + id
}
