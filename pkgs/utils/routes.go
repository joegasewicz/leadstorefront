package utils

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

const APIVersion = "/api/v1"

type Pagination struct {
	Page     int   `json:"page"`
	Limit    int   `json:"limit"`
	Total    int64 `json:"total"`
	NextPage int   `json:"next_page"`
	PrevPage int   `json:"prev_page"`
	HasNext  bool  `json:"has_next"`
	HasPrev  bool  `json:"has_prev"`
}

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

func GetPagination(c *gin.Context) (int, int, int) {
	page := queryInt(c, "page", 1)
	limit := queryInt(c, "limit", 10)
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}
	return page, limit, (page - 1) * limit
}

func GetPaginationQuery(c *gin.Context) (string, string) {
	page := queryInt(c, "page", 1)
	limit := queryInt(c, "limit", 10)
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}
	return strconv.Itoa(page), strconv.Itoa(limit)
}

func NewPagination(page int, limit int, total int64) Pagination {
	pagination := Pagination{
		Page:  page,
		Limit: limit,
		Total: total,
	}
	if int64(page*limit) < total {
		pagination.NextPage = page + 1
		pagination.HasNext = true
	}
	if page > 1 {
		pagination.PrevPage = page - 1
		pagination.HasPrev = true
	}
	return pagination
}

func queryInt(c *gin.Context, key string, fallback int) int {
	value := strings.TrimSpace(c.Query(key))
	if value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return parsed
}
