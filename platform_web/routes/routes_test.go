package routes

import (
	"html/template"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"leadstorefront/pkgs/middleware"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestRegisteredWebRoutes(t *testing.T) {
	router := testWebRouter()

	tests := []struct {
		name     string
		method   string
		path     string
		headers  map[string]string
		expected int
		location string
		body     string
	}{
		{
			name:     "health",
			method:   http.MethodGet,
			path:     "/health",
			expected: http.StatusOK,
			body:     `{"service":"platform_web","status":"ok"}`,
		},
		{
			name:     "root redirects to detected country",
			method:   http.MethodGet,
			path:     "/",
			headers:  map[string]string{"X-Country-Code": "US"},
			expected: http.StatusFound,
			location: "/us",
		},
		{
			name:     "unsupported country redirects to default",
			method:   http.MethodGet,
			path:     "/fr",
			expected: http.StatusFound,
			location: "/uk",
		},
		{
			name:     "unsupported product country redirects to default country products",
			method:   http.MethodGet,
			path:     "/fr/products",
			expected: http.StatusFound,
			location: "/uk/products",
		},
		{
			name:     "unsupported storefront country redirects to default country storefront",
			method:   http.MethodGet,
			path:     "/fr/storefronts/demo",
			expected: http.StatusFound,
			location: "/uk/storefronts/demo",
		},
		{
			name:     "unsupported article country redirects to default country articles",
			method:   http.MethodGet,
			path:     "/fr/articles",
			expected: http.StatusFound,
			location: "/uk/articles",
		},
		{
			name:     "admin requires login",
			method:   http.MethodGet,
			path:     "/admin",
			expected: http.StatusFound,
			location: "/admin/login",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response := performWebRequest(router, tt.method, tt.path, "", tt.headers)

			assert.Equal(t, tt.expected, response.Code)
			if tt.location != "" {
				assert.Equal(t, tt.location, response.Header().Get("Location"))
			}
			if tt.body != "" {
				assert.JSONEq(t, tt.body, response.Body.String())
			}
		})
	}
}

func TestHealthRejectsUnsupportedMethods(t *testing.T) {
	router := testWebRouter()

	for _, method := range []string{http.MethodPost, http.MethodPut, http.MethodDelete} {
		t.Run(method, func(t *testing.T) {
			response := performWebRequest(router, method, "/health", "", nil)

			assert.Equal(t, http.StatusMethodNotAllowed, response.Code)
		})
	}
}

func TestAdminFormsValidateBeforeCallingAPI(t *testing.T) {
	router := testWebRouter()

	login := performWebRequest(router, http.MethodPost, "/admin/login", url.Values{}.Encode(), map[string]string{
		"Content-Type": "application/x-www-form-urlencoded",
	})
	assert.Equal(t, http.StatusUnauthorized, login.Code)
	assert.Contains(t, login.Body.String(), "Email and password are required.")

	register := performWebRequest(router, http.MethodPost, "/admin/register", url.Values{
		"email":    {"editor@example.com"},
		"password": {"secret"},
	}.Encode(), map[string]string{"Content-Type": "application/x-www-form-urlencoded"})
	assert.Equal(t, http.StatusBadRequest, register.Code)
	assert.Contains(t, register.Body.String(), "Could not create the account.")
}

func TestAPIClientURL(t *testing.T) {
	client := &APIClient{BaseURL: "http://api:8001"}

	assert.Equal(t, "http://api:8001/api/v1/health", client.URL("/health"))
	assert.Equal(t, "http://api:8001/api/v1/uk/products", client.URL("uk/products"))
}

func testWebRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(middleware.Sessions())
	router.SetHTMLTemplate(template.Must(template.New("test").Parse(webTestTemplates)))
	Register(router)
	return router
}

func performWebRequest(router http.Handler, method string, path string, body string, headers map[string]string) *httptest.ResponseRecorder {
	request := httptest.NewRequest(method, path, strings.NewReader(body))
	request.Host = "leadstorefront.com"
	for key, value := range headers {
		if strings.EqualFold(key, "Host") {
			request.Host = value
		} else {
			request.Header.Set(key, value)
		}
	}
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	return response
}

const webTestTemplates = `
{{ define "home" }}home {{ .Country }}{{ end }}
{{ define "storefront_show" }}storefront {{ .Storefront.Name }}{{ end }}
{{ define "admin_login" }}admin login {{ .Error }}{{ end }}
{{ define "admin_register" }}admin register {{ .Error }}{{ end }}
{{ define "admin_home" }}admin home {{ .Email }}{{ end }}
{{ define "admin_storefronts_index" }}admin storefronts{{ end }}
{{ define "admin_storefront_form" }}admin storefront form {{ .Error }}{{ end }}
{{ define "products_index" }}products{{ end }}
{{ define "product_show" }}product{{ end }}
{{ define "articles_index" }}articles{{ end }}
{{ define "article_show" }}article{{ end }}
{{ define "admin_articles_index" }}admin articles{{ end }}
{{ define "admin_article_form" }}admin article form {{ .Error }}{{ end }}
{{ define "admin_products_index" }}admin products{{ end }}
{{ define "admin_product_form" }}admin product form {{ .Error }}{{ end }}
`
