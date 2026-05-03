package routes

import (
	"gadgetscout/pkgs/models"
	"gadgetscout/pkgs/utils"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Country struct {
	DB *gorm.DB
}

func (country *Country) Get(c *gin.Context) {
	record, ok := country.byCode(c)
	if !ok {
		return
	}

	var products []models.Product
	if err := country.DB.Preload("Category").Preload("Country").Where("country_id = ?", record.ID).Order("updated_at desc").Limit(12).Find(&products).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not load latest deals"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"latest_deals": products})
}

func (country *Country) Post(c *gin.Context) {
	c.AbortWithStatus(http.StatusMethodNotAllowed)
}

func (country *Country) Put(c *gin.Context) {
	c.AbortWithStatus(http.StatusMethodNotAllowed)
}

func (country *Country) Delete(c *gin.Context) {
	c.AbortWithStatus(http.StatusMethodNotAllowed)
}

func (country *Country) byCode(c *gin.Context) (models.Country, bool) {
	var record models.Country
	if err := country.DB.Where("code = ?", strings.ToLower(utils.CountryCodeFromRequest(c))).First(&record).Error; err != nil {
		utils.WriteRecordError(c, err, "could not load country")
		return models.Country{}, false
	}
	return record, true
}
