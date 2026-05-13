package routes

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStaticAPIRoutes(t *testing.T) {
	router := testRouter()

	tests := []struct {
		name     string
		path     string
		expected map[string]interface{}
	}{
		{
			name: "health",
			path: "/api/v1/health",
			expected: map[string]interface{}{
				"service": "platform_api",
				"status":  "ok",
			},
		},
		{
			name: "root",
			path: "/api/v1",
			expected: map[string]interface{}{
				"service": "platform_api",
				"route":   "/",
			},
		},
		{
			name: "admin home",
			path: "/api/v1/admin",
			expected: map[string]interface{}{
				"route": "/admin",
			},
		},
		{
			name: "admin login page",
			path: "/api/v1/admin/login",
			expected: map[string]interface{}{
				"route": "/api/v1/admin/login",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response := performRequest(router, http.MethodGet, tt.path, "")

			require.Equal(t, http.StatusOK, response.Code)
			assertJSONIncludes(t, response.Body.String(), tt.expected)
		})
	}
}

func TestAPIRoutesRejectInvalidJSON(t *testing.T) {
	router := testRouter()

	tests := []struct {
		name  string
		path  string
		error string
	}{
		{name: "login", path: "/api/v1/admin/login", error: "invalid login"},
		{name: "register", path: "/api/v1/admin/register", error: "invalid registration"},
		{name: "storefront create", path: "/api/v1/admin/storefronts/create", error: "invalid storefront"},
		{name: "product storefront assign", path: "/api/v1/admin/storefronts/1/products", error: "invalid product storefront"},
		{name: "article storefront assign", path: "/api/v1/admin/storefronts/1/articles", error: "invalid article storefront"},
		{name: "product create", path: "/api/v1/admin/products/create", error: "invalid product"},
		{name: "article create", path: "/api/v1/admin/articles/create", error: "invalid article"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response := performRequest(router, http.MethodPost, tt.path, "{")

			require.Equal(t, http.StatusBadRequest, response.Code)
			assertJSONIncludes(t, response.Body.String(), map[string]interface{}{"error": tt.error})
		})
	}
}

func TestRegisterRejectsUnsupportedMethods(t *testing.T) {
	router := testRouter()

	tests := []struct {
		name   string
		method string
		path   string
	}{
		{name: "health post", method: http.MethodPost, path: "/api/v1/health"},
		{name: "product create get missing route", method: http.MethodGet, path: "/api/v1/admin/products/create/unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response := performRequest(router, tt.method, tt.path, "")

			assert.Equal(t, http.StatusNotFound, response.Code)
		})
	}
}

func TestUserManagementRequiresAuth(t *testing.T) {
	router := testRouter()

	for _, path := range []string{"/api/v1/admin/users", "/api/v1/admin/users/1"} {
		t.Run(path, func(t *testing.T) {
			response := performRequest(router, http.MethodGet, path, "")

			assert.Equal(t, http.StatusUnauthorized, response.Code)
			assertJSONIncludes(t, response.Body.String(), map[string]interface{}{"error": "unauthorized"})
		})
	}
}

func TestDomainLookupCandidates(t *testing.T) {
	assert.Equal(t, []string{"lankanote.com", "www.lankanote.com"}, domainLookupCandidates("www.lankanote.com."))
}

func testRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	Register(router, nil)
	return router
}

func performRequest(router http.Handler, method string, path string, body string) *httptest.ResponseRecorder {
	request := httptest.NewRequest(method, path, strings.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	return response
}

func assertJSONIncludes(t *testing.T, raw string, expected map[string]interface{}) {
	t.Helper()

	var actual map[string]interface{}
	require.NoError(t, json.Unmarshal([]byte(raw), &actual))

	for key, value := range expected {
		assert.Equal(t, value, actual[key])
	}
}
