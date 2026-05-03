package utils

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

const APIVersion = "/api/v1"

func GetVersion(path string) string {
	if path == "" || path == "/" {
		return APIVersion
	}
	if path[0] != '/' {
		return APIVersion + "/" + path
	}
	return APIVersion + path
}

func APIRoot(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"service": "api", "route": "/"})
}

func PtrTime(value time.Time) *time.Time {
	return &value
}

func CountryCodeFromRequest(c *gin.Context) string {
	if country := c.Param("country"); country != "" {
		return country
	}
	path := strings.TrimPrefix(c.Request.URL.Path, APIVersion+"/")
	parts := strings.Split(path, "/")
	if len(parts) == 0 {
		return ""
	}
	return parts[0]
}

func AdminHome(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"route": "/admin"})
}

func AdminStatic(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"route": c.FullPath()})
}

func WriteRecordError(c *gin.Context, err error, message string) {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	c.JSON(http.StatusInternalServerError, gin.H{"error": message})
}

func Slugify(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	var builder strings.Builder
	lastWasDash := false

	for _, char := range value {
		if (char >= 'a' && char <= 'z') || (char >= '0' && char <= '9') {
			builder.WriteRune(char)
			lastWasDash = false
			continue
		}

		if !lastWasDash {
			builder.WriteRune('-')
			lastWasDash = true
		}
	}

	return strings.Trim(builder.String(), "-")
}
