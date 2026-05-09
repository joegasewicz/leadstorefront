package routes

import (
	"gadgetscout/pkgs/models"
	"gadgetscout/pkgs/utils"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type User struct {
	DB *gorm.DB
}

func (user *User) Get(c *gin.Context) {
	var record models.User
	err := user.DB.Preload("Role").First(&record, c.Param("id")).Error
	if err != nil {
		utils.WriteRecordError(c, err, "could not load user")
		return
	}
	c.JSON(http.StatusOK, gin.H{"user": record})
}

func (user *User) Post(c *gin.Context) {
	if strings.HasSuffix(c.FullPath(), "/login") {
		user.login(c)
		return
	}
	user.register(c)
}

func (user *User) Put(c *gin.Context) {
	c.AbortWithStatus(http.StatusMethodNotAllowed)
}

func (user *User) Delete(c *gin.Context) {
	c.AbortWithStatus(http.StatusMethodNotAllowed)
}

func (user *User) login(c *gin.Context) {
	var request struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid login"})
		return
	}
	var record models.User
	if err := user.DB.Preload("Role").Where("email = ?", strings.ToLower(strings.TrimSpace(request.Email))).First(&record).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid login"})
		return
	}
	if err := bcrypt.CompareHashAndPassword([]byte(record.Password), []byte(request.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid login"})
		return
	}
	if record.Role.Name != "admin" && record.Role.Name != "editor" {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"user": record})
}

func (user *User) register(c *gin.Context) {
	var request struct {
		Email    string `json:"email"`
		Password string `json:"password"`
		Role     string `json:"role"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid registration"})
		return
	}
	roleName := strings.ToLower(strings.TrimSpace(request.Role))
	if roleName != "admin" && roleName != "editor" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid role"})
		return
	}
	var role models.Role
	if err := user.DB.Where("name = ?", roleName).First(&role).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "role has not been seeded"})
		return
	}
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(request.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not hash password"})
		return
	}
	record := models.User{Email: strings.ToLower(strings.TrimSpace(request.Email)), Password: string(passwordHash), RoleID: role.ID}
	if err := user.DB.Create(&record).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not create user"})
		return
	}
	_ = user.DB.Preload("Role").First(&record, record.ID).Error
	c.JSON(http.StatusCreated, gin.H{"user": record})
}
