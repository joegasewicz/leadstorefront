package routes

import (
	"gadgetscout/pkgs/models"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type ProductCategory struct {
	DB *gorm.DB
}

func (category *ProductCategory) Get(c *gin.Context) {
	var countries []models.Country
	var categories []models.ProductCategory
	_ = category.DB.Order("name asc").Find(&countries).Error
	_ = category.DB.Order("name asc").Find(&categories).Error
	c.JSON(http.StatusOK, gin.H{"countries": countries, "categories": categories})
}

func (category *ProductCategory) Post(c *gin.Context) {
	c.AbortWithStatus(http.StatusMethodNotAllowed)
}

func (category *ProductCategory) Put(c *gin.Context) {
	c.AbortWithStatus(http.StatusMethodNotAllowed)
}

func (category *ProductCategory) Delete(c *gin.Context) {
	c.AbortWithStatus(http.StatusMethodNotAllowed)
}
