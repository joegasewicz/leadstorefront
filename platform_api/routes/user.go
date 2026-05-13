package routes

import (
	"leadstorefront/pkgs/models"
	"leadstorefront/pkgs/utils"
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
	currentUser, ok := currentAPIUser(c, user.DB)
	if !ok {
		return
	}
	if c.FullPath() == utils.GetVersion("/admin/users") {
		user.index(c, currentUser)
		return
	}
	requestedUserID := uintPathID(c.Param("id"))
	if requestedUserID == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	if !isSuper(currentUser) && currentUser.ID != requestedUserID {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	}

	var record models.User
	err := user.DB.Preload("Role").First(&record, requestedUserID).Error
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
	currentUser, ok := currentAPIUser(c, user.DB)
	if !ok {
		return
	}
	if !isSuper(currentUser) {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	}

	requestedUserID := uintPathID(c.Param("id"))
	if requestedUserID == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}

	var request struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user"})
		return
	}

	var record models.User
	if err := user.DB.Preload("Role").First(&record, requestedUserID).Error; err != nil {
		utils.WriteRecordError(c, err, "could not load user")
		return
	}

	email := strings.ToLower(strings.TrimSpace(request.Email))
	if email == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "email is required"})
		return
	}
	record.Email = email

	if strings.TrimSpace(request.Password) != "" {
		passwordHash, err := bcrypt.GenerateFromPassword([]byte(request.Password), bcrypt.DefaultCost)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "could not hash password"})
			return
		}
		record.Password = string(passwordHash)
	}

	if err := user.DB.Save(&record).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not update user"})
		return
	}
	_ = user.DB.Preload("Role").First(&record, record.ID).Error
	c.JSON(http.StatusOK, gin.H{"user": record})
}

func (user *User) Delete(c *gin.Context) {
	currentUser, ok := currentAPIUser(c, user.DB)
	if !ok {
		return
	}
	if !isSuper(currentUser) {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	}

	requestedUserID := uintPathID(c.Param("id"))
	if requestedUserID == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	if requestedUserID == currentUser.ID {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cannot delete your own user"})
		return
	}

	var record models.User
	if err := user.DB.First(&record, requestedUserID).Error; err != nil {
		utils.WriteRecordError(c, err, "could not load user")
		return
	}
	err := user.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&models.Storefront{}).Where("owner_id = ?", requestedUserID).Update("is_active", false).Error; err != nil {
			return err
		}
		return tx.Delete(&record).Error
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not delete user"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"deleted": true})
}

func (user *User) index(c *gin.Context, currentUser models.User) {
	if !isSuper(currentUser) {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	}

	page, limit, offset := utils.GetPagination(c)
	var users []models.User
	var total int64
	query := user.DB.Model(&models.User{})
	if err := query.Count(&total).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not count users"})
		return
	}
	if err := query.Preload("Role").Order("created_at desc").Limit(limit).Offset(offset).Find(&users).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not load users"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"users": users, "pagination": utils.NewPagination(page, limit, total)})
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
	if record.Role.Name != "super" && record.Role.Name != "admin" && record.Role.Name != "editor" && record.Role.Name != "user" {
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
	if roleName != "admin" && roleName != "editor" && roleName != "user" {
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
