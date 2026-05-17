package routes

import (
	"encoding/json"
	"leadstorefront/pkgs/models"
	"leadstorefront/pkgs/utils"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type OutboundClick struct {
	DB *gorm.DB
}

func (route *OutboundClick) Get(c *gin.Context) {
	user, ok := currentAPIUser(c, route.DB)
	if !ok {
		return
	}
	page, limit, offset := utils.GetPagination(c)
	var clicks []models.OutboundClick
	var total int64
	query := route.DB.Model(&models.OutboundClick{}).Joins("JOIN storefronts ON storefronts.id = outbound_clicks.storefront_id").Preload("Storefront").Preload("Product")
	if !isSuper(user) {
		query = query.Where("storefronts.owner_id = ?", user.ID)
	}
	if err := query.Count(&total).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not count outbound clicks"})
		return
	}
	if err := query.Order("outbound_clicks.created_at desc").Limit(limit).Offset(offset).Find(&clicks).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not load outbound clicks"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"outbound_clicks": clicks, "pagination": utils.NewPagination(page, limit, total)})
}

func (route *OutboundClick) Post(c *gin.Context) {
	var request struct {
		DestinationURL string                   `json:"destination_url"`
		CountryCode    string                   `json:"country_code"`
		ProductID      *uint                    `json:"product_id"`
		VisitorID      string                   `json:"visitor_id"`
		Attribution    utils.AttributionPayload `json:"attribution"`
		LandingPath    string                   `json:"landing_path"`
		Referrer       string                   `json:"referrer"`
		UserAgent      string                   `json:"user_agent"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid outbound click"})
		return
	}
	storefrontID := leadStorefrontID(c)
	if storefrontID == 0 || strings.TrimSpace(request.DestinationURL) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid outbound click"})
		return
	}
	if !activeOutboundStorefrontExists(c, route.DB, storefrontID) {
		return
	}
	attributionJSON, err := json.Marshal(request.Attribution)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid outbound attribution"})
		return
	}
	record := models.OutboundClick{
		StorefrontID:    storefrontID,
		ProductID:       request.ProductID,
		VisitorID:       strings.TrimSpace(request.VisitorID),
		CountryCode:     strings.ToLower(strings.TrimSpace(request.CountryCode)),
		DestinationURL:  strings.TrimSpace(request.DestinationURL),
		AttributionJSON: models.JSONB(attributionJSON),
		CampaignSource:  request.Attribution.Source(),
		CampaignMedium:  request.Attribution.Medium(),
		CampaignName:    request.Attribution.Campaign(),
		AffiliateSource: request.Attribution.Partner(),
		MarketSource:    request.Attribution.Market(),
		ClickID:         request.Attribution.ClickID(),
		LandingPath:     strings.TrimSpace(request.LandingPath),
		Referrer:        strings.TrimSpace(request.Referrer),
		UserAgent:       strings.TrimSpace(request.UserAgent),
	}
	if err := route.DB.Create(&record).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not save outbound click"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"outbound_click": record})
}

func (route *OutboundClick) Put(c *gin.Context) {
	c.AbortWithStatus(http.StatusMethodNotAllowed)
}

func (route *OutboundClick) Delete(c *gin.Context) {
	c.AbortWithStatus(http.StatusMethodNotAllowed)
}

func activeOutboundStorefrontExists(c *gin.Context, db *gorm.DB, storefrontID uint) bool {
	var count int64
	if err := db.Model(&models.Storefront{}).Where("id = ? AND is_active = ?", storefrontID, true).Count(&count).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not load storefront"})
		return false
	}
	if count == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return false
	}
	return true
}
