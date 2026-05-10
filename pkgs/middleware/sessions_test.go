package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSessionUserAndFlashHelpers(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(Sessions())

	router.GET("/set", func(c *gin.Context) {
		require.NoError(t, SetUserSession(c, 42))
		require.NoError(t, SetFlash(c, "saved"))
		c.Status(http.StatusNoContent)
	})
	router.GET("/read", func(c *gin.Context) {
		userID, ok := CurrentUserID(c)
		c.JSON(http.StatusOK, gin.H{
			"user_id": userID,
			"ok":      ok,
			"flash":   PopFlash(c),
		})
	})
	router.GET("/clear", func(c *gin.Context) {
		require.NoError(t, ClearUserSession(c))
		c.Status(http.StatusNoContent)
	})

	cookie := latestCookie("", performSessionRequest(router, http.MethodGet, "/set", ""))

	read := performSessionRequest(router, http.MethodGet, "/read", cookie)
	assert.Equal(t, http.StatusOK, read.Code)
	assert.JSONEq(t, `{"user_id":42,"ok":true,"flash":"saved"}`, read.Body.String())
	cookie = latestCookie(cookie, read)

	readAgain := performSessionRequest(router, http.MethodGet, "/read", cookie)
	assert.Equal(t, http.StatusOK, readAgain.Code)
	assert.JSONEq(t, `{"user_id":42,"ok":true,"flash":""}`, readAgain.Body.String())
	cookie = latestCookie(cookie, readAgain)

	cleared := performSessionRequest(router, http.MethodGet, "/clear", cookie)
	assert.Equal(t, http.StatusNoContent, cleared.Code)
	cookie = latestCookie(cookie, cleared)

	afterClear := performSessionRequest(router, http.MethodGet, "/read", cookie)
	assert.JSONEq(t, `{"user_id":0,"ok":false,"flash":""}`, afterClear.Body.String())
}

func TestRequireRolesRedirectsWithoutSession(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(Sessions())
	router.GET("/admin", RequireRoles("admin"), func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})

	response := performSessionRequest(router, http.MethodGet, "/admin", "")

	assert.Equal(t, http.StatusFound, response.Code)
	assert.Equal(t, "/admin/login", response.Header().Get("Location"))
}

func performSessionRequest(router http.Handler, method string, path string, cookie string) *httptest.ResponseRecorder {
	request := httptest.NewRequest(method, path, nil)
	if cookie != "" {
		request.Header.Set("Cookie", cookie)
	}
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	return response
}

func latestCookie(current string, response *httptest.ResponseRecorder) string {
	cookies := response.Result().Cookies()
	if len(cookies) == 0 {
		return current
	}
	return cookies[0].String()
}
