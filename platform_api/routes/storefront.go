package routes

import (
	"leadstorefront/pkgs/models"
	"leadstorefront/pkgs/utils"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Storefront struct {
	DB *gorm.DB
}

func (storefront *Storefront) Get(c *gin.Context) {
	if strings.HasSuffix(c.FullPath(), "/create") {
		storefront.getOptions(c)
		return
	}
	if c.Param("domain") != "" {
		storefront.getActiveByDomain(c)
		return
	}
	if c.Param("id") != "" {
		storefront.getByID(c)
		return
	}
	if c.Param("slug") != "" {
		storefront.getActiveBySlug(c)
		return
	}
	storefront.getAdminList(c)
}

func (storefront *Storefront) getOptions(c *gin.Context) {
	user, ok := currentAPIUser(c, storefront.DB)
	if !ok {
		return
	}

	var countries []models.Country
	var users []models.User
	_ = storefront.DB.Order("name asc").Find(&countries).Error
	if isSuper(user) {
		_ = storefront.DB.Preload("Role").Order("email asc").Find(&users).Error
	} else {
		users = []models.User{user}
	}
	c.JSON(http.StatusOK, gin.H{"countries": countries, "users": users})
}

func (storefront *Storefront) Post(c *gin.Context) {
	if strings.Contains(c.FullPath(), "/products") {
		storefront.postProduct(c)
		return
	}
	if strings.Contains(c.FullPath(), "/articles") {
		storefront.postArticle(c)
		return
	}
	record, ok := storefront.bindJSON(c)
	if !ok {
		return
	}
	user, ok := currentAPIUser(c, storefront.DB)
	if !ok {
		return
	}
	if !isSuper(user) {
		record.OwnerID = &user.ID
	}
	if err := storefront.DB.Create(&record).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not create storefront"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"storefront": record})
}

func (storefront *Storefront) postProduct(c *gin.Context) {
	var request struct {
		ProductID uint `json:"product_id"`
	}
	if err := c.ShouldBindJSON(&request); err != nil || request.ProductID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid product storefront"})
		return
	}
	record := models.ProductStorefront{ProductID: request.ProductID, StorefrontID: uintPathID(c.Param("id"))}
	if record.StorefrontID == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	if ok := storefront.authorizeStorefrontID(c, record.StorefrontID); !ok {
		return
	}
	if err := storefront.DB.Where("product_id = ? AND storefront_id = ?", record.ProductID, record.StorefrontID).FirstOrCreate(&record).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not assign product"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"product_storefront": record})
}

func (storefront *Storefront) postArticle(c *gin.Context) {
	var request struct {
		ArticleID uint `json:"article_id"`
	}
	if err := c.ShouldBindJSON(&request); err != nil || request.ArticleID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid article storefront"})
		return
	}
	record := models.ArticleStorefront{ArticleID: request.ArticleID, StorefrontID: uintPathID(c.Param("id"))}
	if record.StorefrontID == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	if ok := storefront.authorizeStorefrontID(c, record.StorefrontID); !ok {
		return
	}
	if err := storefront.DB.Where("article_id = ? AND storefront_id = ?", record.ArticleID, record.StorefrontID).FirstOrCreate(&record).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not assign article"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"article_storefront": record})
}

func (storefront *Storefront) Put(c *gin.Context) {
	record, ok := storefront.bindJSON(c)
	if !ok {
		return
	}
	user, ok := currentAPIUser(c, storefront.DB)
	if !ok {
		return
	}
	var existing models.Storefront
	query := storefront.DB
	if !isSuper(user) {
		query = query.Where("owner_id = ?", user.ID)
	}
	if err := query.First(&existing, c.Param("id")).Error; err != nil {
		utils.WriteRecordError(c, err, "could not load storefront")
		return
	}
	if !isSuper(user) {
		record.OwnerID = existing.OwnerID
	}
	if err := storefront.DB.Model(&existing).Updates(storefrontUpdateMap(record)).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not update storefront"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"storefront": existing})
}

func (storefront *Storefront) Delete(c *gin.Context) {
	user, ok := currentAPIUser(c, storefront.DB)
	if !ok {
		return
	}
	query := storefront.DB
	if !isSuper(user) {
		query = query.Where("owner_id = ?", user.ID)
	}
	result := query.Delete(&models.Storefront{}, c.Param("id"))
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not delete storefront"})
		return
	}
	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"deleted": true})
}

func (storefront *Storefront) getAdminList(c *gin.Context) {
	user, ok := currentAPIUser(c, storefront.DB)
	if !ok {
		return
	}
	var storefronts []models.Storefront
	var total int64
	page, limit, offset := utils.GetPagination(c)
	query := storefront.DB.Model(&models.Storefront{})
	if !isSuper(user) {
		query = query.Where("owner_id = ?", user.ID)
	}
	if err := query.Count(&total).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not count storefronts"})
		return
	}
	if err := query.Preload("PrimaryCountry").Preload("Owner.Role").Order("created_at desc").Limit(limit).Offset(offset).Find(&storefronts).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not load storefronts"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"storefronts": storefronts, "pagination": utils.NewPagination(page, limit, total)})
}

func (storefront *Storefront) getByID(c *gin.Context) {
	user, ok := currentAPIUser(c, storefront.DB)
	if !ok {
		return
	}
	var record models.Storefront
	query := storefront.DB.Preload("PrimaryCountry").Preload("Owner.Role")
	if !isSuper(user) {
		query = query.Where("owner_id = ?", user.ID)
	}
	if err := query.First(&record, c.Param("id")).Error; err != nil {
		utils.WriteRecordError(c, err, "could not load storefront")
		return
	}
	c.JSON(http.StatusOK, gin.H{"storefront": record})
}

func (storefront *Storefront) authorizeStorefrontID(c *gin.Context, storefrontID uint) bool {
	user, ok := currentAPIUser(c, storefront.DB)
	if !ok {
		return false
	}
	if isSuper(user) {
		return true
	}
	var count int64
	if err := storefront.DB.Model(&models.Storefront{}).Where("id = ? AND owner_id = ?", storefrontID, user.ID).Count(&count).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not load storefront"})
		return false
	}
	if count == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return false
	}
	return true
}

func (storefront *Storefront) getActiveBySlug(c *gin.Context) {
	var record models.Storefront
	err := storefront.DB.Preload("PrimaryCountry").Where("slug = ? AND is_active = ?", c.Param("slug"), true).First(&record).Error
	if err != nil {
		utils.WriteRecordError(c, err, "could not load storefront")
		return
	}
	c.JSON(http.StatusOK, gin.H{"storefront": record})
}

func (storefront *Storefront) getActiveByDomain(c *gin.Context) {
	domain := strings.ToLower(strings.TrimSpace(c.Param("domain")))
	var record models.Storefront
	err := storefront.DB.Preload("PrimaryCountry").Where("domain IN ? AND is_active = ?", domainLookupCandidates(domain), true).First(&record).Error
	if err != nil {
		utils.WriteRecordError(c, err, "could not load storefront")
		return
	}
	c.JSON(http.StatusOK, gin.H{"storefront": record})
}

func domainLookupCandidates(domain string) []string {
	domain = strings.TrimPrefix(strings.TrimSuffix(domain, "."), "www.")
	if domain == "" {
		return []string{""}
	}
	return []string{domain, "www." + domain}
}

func (storefront *Storefront) bindJSON(c *gin.Context) (models.Storefront, bool) {
	var record models.Storefront
	if err := c.ShouldBindJSON(&record); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid storefront"})
		return models.Storefront{}, false
	}
	record.Name = strings.TrimSpace(record.Name)
	record.Domain = strings.ToLower(strings.TrimSpace(record.Domain))
	if record.Slug == "" {
		record.Slug = utils.Slugify(record.Name)
	} else {
		record.Slug = utils.Slugify(record.Slug)
	}
	return record, true
}

func storefrontUpdateMap(storefront models.Storefront) map[string]interface{} {
	return map[string]interface{}{
		"name":               storefront.Name,
		"slug":               storefront.Slug,
		"domain":             storefront.Domain,
		"description":        storefront.Description,
		"logo_url":           storefront.LogoURL,
		"is_active":          storefront.IsActive,
		"primary_country_id": storefront.PrimaryCountryID,
		"owner_id":           storefront.OwnerID,
	}
}

func uintPathID(value string) uint {
	parsed, err := strconv.Atoi(value)
	if err != nil || parsed <= 0 {
		return 0
	}
	return uint(parsed)
}
