package routes

import (
	"gadgetscout/pkgs/models"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type ArticleCategory struct {
	DB *gorm.DB
}

func (category *ArticleCategory) Get(c *gin.Context) {
	var categories []models.ArticleCategory
	var products []models.Product
	_ = category.DB.Order("name asc").Find(&categories).Error
	_ = category.DB.Order("name asc").Find(&products).Error
	c.JSON(http.StatusOK, gin.H{"categories": categories, "products": products})
}

func (category *ArticleCategory) Post(c *gin.Context) {
	c.AbortWithStatus(http.StatusMethodNotAllowed)
}

func (category *ArticleCategory) Put(c *gin.Context) {
	c.AbortWithStatus(http.StatusMethodNotAllowed)
}

func (category *ArticleCategory) Delete(c *gin.Context) {
	c.AbortWithStatus(http.StatusMethodNotAllowed)
}
