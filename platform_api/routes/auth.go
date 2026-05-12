package routes

import (
	"leadstorefront/pkgs/models"
	"leadstorefront/pkgs/utils"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func currentAPIUser(c *gin.Context, db *gorm.DB) (models.User, bool) {
	userID, ok := utils.UserIDFromAuthHeader(c.GetHeader("Authorization"))
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return models.User{}, false
	}

	var user models.User
	if err := db.Preload("Role").First(&user, userID).Error; err != nil {
		utils.WriteRecordError(c, err, "could not load user")
		return models.User{}, false
	}
	return user, true
}

func isSuper(user models.User) bool {
	return user.Role.Name == "super"
}
