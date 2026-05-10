package routes

import (
	"leadstorefront/pkgs/models"
	"leadstorefront/pkgs/utils"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Storefront struct {
	DB *gorm.DB
}

func (storefront *Storefront) Get(c *gin.Context) {
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

func (storefront *Storefront) Post(c *gin.Context) {
	record, ok := storefront.bindJSON(c)
	if !ok {
		return
	}
	if err := storefront.DB.Create(&record).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not create storefront"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"storefront": record})
}

func (storefront *Storefront) Put(c *gin.Context) {
	record, ok := storefront.bindJSON(c)
	if !ok {
		return
	}
	var existing models.Storefront
	if err := storefront.DB.First(&existing, c.Param("id")).Error; err != nil {
		utils.WriteRecordError(c, err, "could not load storefront")
		return
	}
	if err := storefront.DB.Model(&existing).Updates(storefrontUpdateMap(record)).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not update storefront"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"storefront": existing})
}

func (storefront *Storefront) Delete(c *gin.Context) {
	if err := storefront.DB.Delete(&models.Storefront{}, c.Param("id")).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not delete storefront"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"deleted": true})
}

func (storefront *Storefront) getAdminList(c *gin.Context) {
	var storefronts []models.Storefront
	var total int64
	page, limit, offset := utils.GetPagination(c)
	query := storefront.DB.Model(&models.Storefront{})
	if err := query.Count(&total).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not count storefronts"})
		return
	}
	if err := storefront.DB.Preload("PrimaryCountry").Preload("Owner.Role").Order("created_at desc").Limit(limit).Offset(offset).Find(&storefronts).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not load storefronts"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"storefronts": storefronts, "pagination": utils.NewPagination(page, limit, total)})
}

func (storefront *Storefront) getByID(c *gin.Context) {
	var record models.Storefront
	if err := storefront.DB.Preload("PrimaryCountry").Preload("Owner.Role").First(&record, c.Param("id")).Error; err != nil {
		utils.WriteRecordError(c, err, "could not load storefront")
		return
	}
	c.JSON(http.StatusOK, gin.H{"storefront": record})
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
