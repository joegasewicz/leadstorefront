package routes

import (
	"leadstorefront/pkgs/models"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type ArticleCategory struct {
	DB *gorm.DB
}

func (category *ArticleCategory) Get(c *gin.Context) {
	user, ok := currentAPIUser(c, category.DB)
	if !ok {
		return
	}
	var categories []models.ArticleCategory
	var products []models.Product
	var storefronts []models.Storefront
	_ = category.DB.Order("name asc").Find(&categories).Error
	_ = category.DB.Order("name asc").Find(&products).Error
	query := category.DB.Order("name asc")
	if !isSuper(user) {
		query = query.Where("owner_id = ?", user.ID)
	}
	_ = query.Find(&storefronts).Error
	c.JSON(http.StatusOK, gin.H{"categories": categories, "products": products, "storefronts": storefronts})
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
