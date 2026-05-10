package utils

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func TestGetVersion(t *testing.T) {
	assert.Equal(t, "/api/v1", GetVersion(""))
	assert.Equal(t, "/api/v1", GetVersion("/"))
	assert.Equal(t, "/api/v1/health", GetVersion("health"))
	assert.Equal(t, "/api/v1/health", GetVersion("/health"))
}

func TestSlugify(t *testing.T) {
	tests := map[string]string{
		"  Hello World  ":       "hello-world",
		"USB-C Cable + Charger": "usb-c-cable-charger",
		"Already---Slugged":     "already-slugged",
		"!!!":                   "",
	}

	for value, expected := range tests {
		assert.Equal(t, expected, Slugify(value))
	}
}

func TestPtrTime(t *testing.T) {
	now := time.Date(2026, 5, 10, 12, 0, 0, 0, time.UTC)

	assert.Equal(t, now, *PtrTime(now))
}

func TestCountryCodeFromRequest(t *testing.T) {
	router := gin.New()
	router.GET(APIVersion+"/:country/products", func(c *gin.Context) {
		c.String(http.StatusOK, CountryCodeFromRequest(c))
	})
	router.GET(APIVersion+"/admin", func(c *gin.Context) {
		c.String(http.StatusOK, CountryCodeFromRequest(c))
	})

	assert.Equal(t, "uk", performUtilsRequest(router, "/api/v1/uk/products").Body.String())
	assert.Equal(t, "admin", performUtilsRequest(router, "/api/v1/admin").Body.String())
}

func TestPaginationHelpers(t *testing.T) {
	router := gin.New()
	router.GET("/pagination", func(c *gin.Context) {
		page, limit, offset := GetPagination(c)
		pageQuery, limitQuery := GetPaginationQuery(c)
		c.JSON(http.StatusOK, gin.H{
			"page":        page,
			"limit":       limit,
			"offset":      offset,
			"page_query":  pageQuery,
			"limit_query": limitQuery,
		})
	})

	response := performUtilsRequest(router, "/pagination?page=2&limit=150")

	assert.JSONEq(t, `{"page":2,"limit":100,"offset":100,"page_query":"2","limit_query":"100"}`, response.Body.String())
}

func TestNewPagination(t *testing.T) {
	assert.Equal(t, Pagination{
		Page:     2,
		Limit:    10,
		Total:    35,
		NextPage: 3,
		PrevPage: 1,
		HasNext:  true,
		HasPrev:  true,
	}, NewPagination(2, 10, 35))

	assert.Equal(t, Pagination{
		Page:  1,
		Limit: 10,
		Total: 10,
	}, NewPagination(1, 10, 10))
}

func TestStaticHandlers(t *testing.T) {
	router := gin.New()
	router.GET("/api", APIRoot)
	router.GET("/admin", AdminHome)
	router.GET("/admin/login", AdminStatic)

	assert.JSONEq(t, `{"service":"platform_api","route":"/"}`, performUtilsRequest(router, "/api").Body.String())
	assert.JSONEq(t, `{"route":"/admin"}`, performUtilsRequest(router, "/admin").Body.String())
	assert.JSONEq(t, `{"route":"/admin/login"}`, performUtilsRequest(router, "/admin/login").Body.String())
}

func TestWriteRecordError(t *testing.T) {
	router := gin.New()
	router.GET("/missing", func(c *gin.Context) {
		WriteRecordError(c, gorm.ErrRecordNotFound, "could not load record")
	})
	router.GET("/failed", func(c *gin.Context) {
		WriteRecordError(c, errors.New("database failed"), "could not load record")
	})

	missing := performUtilsRequest(router, "/missing")
	assert.Equal(t, http.StatusNotFound, missing.Code)
	assert.JSONEq(t, `{"error":"not found"}`, missing.Body.String())

	failed := performUtilsRequest(router, "/failed")
	assert.Equal(t, http.StatusInternalServerError, failed.Code)
	assert.JSONEq(t, `{"error":"could not load record"}`, failed.Body.String())
}

func performUtilsRequest(router http.Handler, path string) *httptest.ResponseRecorder {
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, path, nil)
	router.ServeHTTP(response, request)
	return response
}
