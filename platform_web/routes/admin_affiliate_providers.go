package routes

import (
	"leadstorefront/pkgs/middleware"
	"leadstorefront/pkgs/models"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

type AdminAffiliateProviders struct {
	API *APIClient
}

func (route *AdminAffiliateProviders) Get(c *gin.Context) {
	if c.Param("id") != "" {
		route.StorefrontIndex(c)
		return
	}
	var response struct {
		Providers []models.AffiliateProvider `json:"providers"`
	}
	if err := route.API.Get(c, "/admin/affiliate-providers", &response); err != nil {
		c.String(http.StatusInternalServerError, "could not load affiliate providers")
		return
	}
	c.HTML(http.StatusOK, "admin_affiliate_providers_index", gin.H{
		"Title":        "Affiliate Providers",
		"Providers":    response.Providers,
		"IsAdmin":      true,
		"IsSuper":      isCurrentSuper(c),
		"IsAdminRoute": true,
	})
}

func (route *AdminAffiliateProviders) StorefrontIndex(c *gin.Context) {
	id, ok := apiPathID(c.Param("id"))
	if !ok {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}
	var response struct {
		Storefront  models.Storefront                    `json:"storefront"`
		Providers   []models.AffiliateProvider           `json:"providers"`
		Connections []models.StorefrontAffiliateProvider `json:"connections"`
	}
	if err := route.API.Get(c, "/admin/storefronts/"+id+"/affiliate-providers", &response); err != nil {
		c.String(http.StatusInternalServerError, "could not load affiliate provider connections")
		return
	}
	c.HTML(http.StatusOK, "admin_storefront_affiliate_providers", gin.H{
		"Title":                response.Storefront.Name + " Affiliate Providers",
		"Storefront":           response.Storefront,
		"Providers":            response.Providers,
		"Connections":          response.Connections,
		"UnconnectedProviders": unconnectedAffiliateProviders(response.Providers, response.Connections),
		"Flash":                middleware.PopFlash(c),
		"IsAdmin":              true,
		"IsSuper":              isCurrentSuper(c),
		"IsAdminRoute":         true,
	})
}

func (route *AdminAffiliateProviders) Post(c *gin.Context) {
	id, ok := apiPathID(c.Param("id"))
	if !ok {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}
	payload, err := affiliateProviderPayload(c)
	if err != nil {
		_ = middleware.SetFlash(c, err.Error())
		c.Redirect(http.StatusFound, "/admin/storefronts/"+id+"/affiliate-providers")
		return
	}
	if c.Param("connection_id") != "" {
		connectionID, ok := apiPathID(c.Param("connection_id"))
		if !ok {
			c.AbortWithStatus(http.StatusNotFound)
			return
		}
		if err := route.API.Put(c, "/admin/storefronts/"+id+"/affiliate-providers/"+connectionID, payload, nil); err != nil {
			_ = middleware.SetFlash(c, "Could not update affiliate provider.")
			c.Redirect(http.StatusFound, "/admin/storefronts/"+id+"/affiliate-providers")
			return
		}
		_ = middleware.SetFlash(c, "Affiliate provider updated.")
		c.Redirect(http.StatusFound, "/admin/storefronts/"+id+"/affiliate-providers")
		return
	}
	if err := route.API.Post(c, "/admin/storefronts/"+id+"/affiliate-providers", payload, nil); err != nil {
		_ = middleware.SetFlash(c, "Could not connect affiliate provider.")
		c.Redirect(http.StatusFound, "/admin/storefronts/"+id+"/affiliate-providers")
		return
	}
	_ = middleware.SetFlash(c, "Affiliate provider connected.")
	c.Redirect(http.StatusFound, "/admin/storefronts/"+id+"/affiliate-providers")
}

func (route *AdminAffiliateProviders) Put(c *gin.Context) {
	c.AbortWithStatus(http.StatusMethodNotAllowed)
}

func (route *AdminAffiliateProviders) Delete(c *gin.Context) {
	c.AbortWithStatus(http.StatusMethodNotAllowed)
}

func affiliateProviderPayload(c *gin.Context) (map[string]interface{}, error) {
	providerID, err := parseRequiredUint(c.PostForm("affiliate_provider_id"), "Select a provider.")
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"affiliate_provider_id": providerID,
		"affiliate_id":          strings.TrimSpace(c.PostForm("affiliate_id")),
		"partner_id":            strings.TrimSpace(c.PostForm("partner_id")),
		"aid":                   strings.TrimSpace(c.PostForm("aid")),
		"cid":                   strings.TrimSpace(c.PostForm("cid")),
		"sub_id_format":         strings.TrimSpace(c.PostForm("sub_id_format")),
		"click_ref_format":      strings.TrimSpace(c.PostForm("click_ref_format")),
		"tracking_domain":       strings.TrimSpace(c.PostForm("tracking_domain")),
		"deep_link_base_url":    strings.TrimSpace(c.PostForm("deep_link_base_url")),
		"api_key":               strings.TrimSpace(c.PostForm("api_key")),
		"default_market":        strings.TrimSpace(c.PostForm("default_market")),
		"commission_type":       strings.TrimSpace(c.PostForm("commission_type")),
	}, nil
}

func unconnectedAffiliateProviders(providers []models.AffiliateProvider, connections []models.StorefrontAffiliateProvider) []models.AffiliateProvider {
	connected := map[uint]struct{}{}
	for _, connection := range connections {
		connected[connection.AffiliateProviderID] = struct{}{}
	}
	result := make([]models.AffiliateProvider, 0, len(providers))
	for _, provider := range providers {
		if _, ok := connected[provider.ID]; !ok {
			result = append(result, provider)
		}
	}
	return result
}
