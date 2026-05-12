package routes

import (
	"leadstorefront/pkgs"
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
		normalizeStorefrontAssetURLs(&response.Storefront)
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
	normalizeStorefrontAssetURLs(&response.Storefront)
	return response.Storefront, true
}

func normalizeStorefrontAssetURLs(storefront *models.Storefront) {
	if storefront == nil {
		return
	}
	if strings.HasPrefix(storefront.LogoURL, "/uploads/") {
		storefront.LogoURL = platformWebOrigin() + storefront.LogoURL
	}
}

func platformWebOrigin() string {
	domain := strings.TrimSpace(pkgs.Config.Web.Domain)
	if domain == "" {
		domain = "localhost"
	}
	if strings.HasPrefix(domain, "http://") || strings.HasPrefix(domain, "https://") {
		return strings.TrimRight(domain, "/")
	}
	if domain == "localhost" || domain == "127.0.0.1" {
		return "http://" + domain + pkgs.Config.Web.Addr
	}
	return "https://" + domain
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
