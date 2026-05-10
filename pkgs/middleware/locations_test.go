package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestDetectCountryCode(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name     string
		headers  map[string]string
		expected string
	}{
		{
			name:     "cdn country header",
			headers:  map[string]string{"CF-IPCountry": "US"},
			expected: "us",
		},
		{
			name:     "gb maps to uk",
			headers:  map[string]string{"X-Vercel-IP-Country": "GB"},
			expected: "uk",
		},
		{
			name:     "accept language country",
			headers:  map[string]string{"Accept-Language": "en-CA,en;q=0.9"},
			expected: "ca",
		},
		{
			name:     "unsupported country falls back",
			headers:  map[string]string{"CF-IPCountry": "FR", "Accept-Language": "fr-FR,fr;q=0.9"},
			expected: DefaultCountryCode,
		},
		{
			name:     "missing country falls back",
			headers:  map[string]string{},
			expected: DefaultCountryCode,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			context := testContext(http.MethodGet, "/", tt.headers)

			assert.Equal(t, tt.expected, DetectCountryCode(context))
		})
	}
}

func TestRedirectToLocalizedHome(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/", RedirectToLocalizedHome())

	request := httptest.NewRequest(http.MethodGet, "/", nil)
	request.Header.Set("X-Country-Code", "SG")
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	assert.Equal(t, http.StatusFound, response.Code)
	assert.Equal(t, "/sg", response.Header().Get("Location"))
}

func TestIsSupportedCountryCode(t *testing.T) {
	assert.True(t, IsSupportedCountryCode("uk"))
	assert.True(t, IsSupportedCountryCode("US"))
	assert.False(t, IsSupportedCountryCode("fr"))
	assert.False(t, IsSupportedCountryCode(""))
}

func testContext(method string, path string, headers map[string]string) *gin.Context {
	response := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(response)
	request := httptest.NewRequest(method, path, nil)
	for key, value := range headers {
		request.Header.Set(key, value)
	}
	context.Request = request
	return context
}
