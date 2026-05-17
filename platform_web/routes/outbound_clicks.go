package routes

import (
	"leadstorefront/pkgs/middleware"
	"leadstorefront/pkgs/models"
	"leadstorefront/pkgs/utils"
	"net/http"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"
)

type OutboundClicks struct {
	API *APIClient
}

func (route *OutboundClicks) Get(c *gin.Context) {
	if strings.HasPrefix(c.FullPath(), "/admin/") {
		route.Index(c)
		return
	}
	storefrontID := storefrontIDFromPath(c)
	if storefrontID == "" {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}
	destination := strings.TrimSpace(c.Query("url"))
	if !validOutboundDestination(destination) {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	payload := map[string]interface{}{
		"destination_url": destination,
		"country_code":    c.Param("country"),
		"referrer":        c.Request.Referer(),
		"user_agent":      c.Request.UserAgent(),
	}
	if payload["country_code"] == "" {
		payload["country_code"] = middleware.DefaultCountryCode
	}
	if productID := uintFromString(c.Query("product_id")); productID != 0 {
		payload["product_id"] = productID
	}
	payload["visitor_id"] = attributionVisitorID(c)
	if attribution, landingPath, ok := currentOutboundAttribution(c, uintFromString(storefrontID)); ok {
		payload["attribution"] = attribution
		payload["landing_path"] = landingPath
	}
	_ = route.API.Post(c, "/storefronts/"+storefrontID+"/outbound-clicks", payload, nil)
	c.Redirect(http.StatusFound, destination)
}

func (route *OutboundClicks) Index(c *gin.Context) {
	var response struct {
		OutboundClicks []models.OutboundClick `json:"outbound_clicks"`
		Pagination     utils.Pagination       `json:"pagination"`
	}
	page, limit := utils.GetPaginationQuery(c)
	if err := route.API.Get(c, "/admin/outbound-clicks?page="+page+"&limit="+limit, &response); err != nil {
		c.String(http.StatusInternalServerError, "could not load outbound clicks")
		return
	}
	c.HTML(http.StatusOK, "admin_outbound_clicks_index", gin.H{
		"Title":          "Affiliate Clicks",
		"OutboundClicks": response.OutboundClicks,
		"Pagination":     response.Pagination,
		"Limit":          limit,
		"IsAdmin":        true,
		"IsSuper":        isCurrentSuper(c),
		"IsAdminRoute":   true,
	})
}

func (route *OutboundClicks) Post(c *gin.Context) {
	c.AbortWithStatus(http.StatusMethodNotAllowed)
}

func (route *OutboundClicks) Put(c *gin.Context) {
	c.AbortWithStatus(http.StatusMethodNotAllowed)
}

func (route *OutboundClicks) Delete(c *gin.Context) {
	c.AbortWithStatus(http.StatusMethodNotAllowed)
}

func outboundDealURL(c *gin.Context, storefront models.Storefront, country string, product models.Product) string {
	destination := strings.TrimSpace(product.AffiliateURL)
	if destination == "" {
		destination = strings.TrimSpace(product.ProductURL)
	}
	if destination == "" {
		return ""
	}
	if storefront.ID == 0 {
		return destination
	}
	path := storefrontBasePath(c, country, storefront) + "/out"
	query := url.Values{}
	query.Set("url", destination)
	query.Set("product_id", uintToString(product.ID))
	return path + "?" + query.Encode()
}

func validOutboundDestination(destination string) bool {
	parsed, err := url.Parse(destination)
	if err != nil {
		return false
	}
	return parsed.Scheme == "https" || parsed.Scheme == "http"
}
