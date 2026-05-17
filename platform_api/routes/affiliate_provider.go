package routes

import (
	"leadstorefront/pkgs/models"
	"leadstorefront/pkgs/utils"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type AffiliateProvider struct {
	DB *gorm.DB
}

func (route *AffiliateProvider) Get(c *gin.Context) {
	if c.Param("id") != "" {
		route.getStorefrontConnections(c)
		return
	}
	var providers []models.AffiliateProvider
	if err := route.DB.Order("name asc").Find(&providers).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not load affiliate providers"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"providers": providers})
}

func (route *AffiliateProvider) Post(c *gin.Context) {
	user, ok := currentAPIUser(c, route.DB)
	if !ok {
		return
	}
	storefrontID := uintPathID(c.Param("id"))
	if storefrontID == 0 || !route.authorizeStorefront(c, user, storefrontID) {
		return
	}
	connection, ok := route.bindConnection(c, storefrontID)
	if !ok {
		return
	}
	if err := route.DB.Where("storefront_id = ? AND affiliate_provider_id = ?", storefrontID, connection.AffiliateProviderID).FirstOrCreate(&connection).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not connect affiliate provider"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"connection": connection})
}

func (route *AffiliateProvider) Put(c *gin.Context) {
	user, ok := currentAPIUser(c, route.DB)
	if !ok {
		return
	}
	storefrontID := uintPathID(c.Param("id"))
	if storefrontID == 0 || !route.authorizeStorefront(c, user, storefrontID) {
		return
	}
	connectionID := uintPathID(c.Param("connection_id"))
	if connectionID == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	record, ok := route.bindConnection(c, storefrontID)
	if !ok {
		return
	}
	var existing models.StorefrontAffiliateProvider
	if err := route.DB.Where("id = ? AND storefront_id = ?", connectionID, storefrontID).First(&existing).Error; err != nil {
		utils.WriteRecordError(c, err, "could not load affiliate provider connection")
		return
	}
	if err := route.DB.Model(&existing).Updates(connectionUpdateMap(record)).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not update affiliate provider connection"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"connection": existing})
}

func (route *AffiliateProvider) Delete(c *gin.Context) {
	user, ok := currentAPIUser(c, route.DB)
	if !ok {
		return
	}
	storefrontID := uintPathID(c.Param("id"))
	if storefrontID == 0 || !route.authorizeStorefront(c, user, storefrontID) {
		return
	}
	connectionID := uintPathID(c.Param("connection_id"))
	if connectionID == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	if err := route.DB.Where("storefront_id = ?", storefrontID).Delete(&models.StorefrontAffiliateProvider{}, connectionID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not disconnect affiliate provider"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"deleted": true})
}

func (route *AffiliateProvider) getStorefrontConnections(c *gin.Context) {
	user, ok := currentAPIUser(c, route.DB)
	if !ok {
		return
	}
	storefrontID := uintPathID(c.Param("id"))
	if storefrontID == 0 || !route.authorizeStorefront(c, user, storefrontID) {
		return
	}
	var storefront models.Storefront
	if err := route.DB.Preload("PrimaryCountry").First(&storefront, storefrontID).Error; err != nil {
		utils.WriteRecordError(c, err, "could not load storefront")
		return
	}
	var providers []models.AffiliateProvider
	if err := route.DB.Order("name asc").Find(&providers).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not load affiliate providers"})
		return
	}
	var connections []models.StorefrontAffiliateProvider
	if err := route.DB.Preload("AffiliateProvider").Where("storefront_id = ?", storefrontID).Order("updated_at desc").Find(&connections).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not load provider connections"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"storefront": storefront, "providers": providers, "connections": connections})
}

func (route *AffiliateProvider) authorizeStorefront(c *gin.Context, user models.User, storefrontID uint) bool {
	if isSuper(user) {
		return true
	}
	var count int64
	if err := route.DB.Model(&models.Storefront{}).Where("id = ? AND owner_id = ?", storefrontID, user.ID).Count(&count).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not load storefront"})
		return false
	}
	if count == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return false
	}
	return true
}

func (route *AffiliateProvider) bindConnection(c *gin.Context, storefrontID uint) (models.StorefrontAffiliateProvider, bool) {
	var request models.StorefrontAffiliateProvider
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid affiliate provider connection"})
		return models.StorefrontAffiliateProvider{}, false
	}
	if request.AffiliateProviderID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "affiliate provider is required"})
		return models.StorefrontAffiliateProvider{}, false
	}
	var provider models.AffiliateProvider
	if err := route.DB.First(&provider, request.AffiliateProviderID).Error; err != nil {
		utils.WriteRecordError(c, err, "could not load affiliate provider")
		return models.StorefrontAffiliateProvider{}, false
	}
	request.StorefrontID = storefrontID
	request.ApprovalStatus = "registration_required"
	request.IsActive = false
	request.DefaultMarket = strings.ToLower(strings.TrimSpace(request.DefaultMarket))
	return request, true
}

func connectionUpdateMap(connection models.StorefrontAffiliateProvider) map[string]interface{} {
	return map[string]interface{}{
		"affiliate_id": connection.AffiliateID, "partner_id": connection.PartnerID,
		"aid": connection.AID, "cid": connection.CID, "sub_id_format": connection.SubIDFormat,
		"click_ref_format": connection.ClickRefFormat, "tracking_domain": connection.TrackingDomain,
		"deep_link_base_url": connection.DeepLinkBaseURL, "api_key": connection.APIKey,
		"default_market": connection.DefaultMarket, "commission_type": connection.CommissionType,
	}
}
