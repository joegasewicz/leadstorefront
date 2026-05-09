package middleware

import (
	"leadstorefront/pkgs"
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/memstore"
	"github.com/gin-gonic/gin"
)

const (
	SessionName = "leadstorefront_session"
	UserIDKey   = "user_id"
	FlashKey    = "flash"
)

func Sessions() gin.HandlerFunc {
	store := memstore.NewStore([]byte(pkgs.Config.Session.Secret))
	store.Options(sessions.Options{
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})

	return sessions.Sessions(SessionName, store)
}

func SetUserSession(c *gin.Context, userID uint) error {
	session := sessions.Default(c)
	session.Set(UserIDKey, userID)
	return session.Save()
}

func ClearUserSession(c *gin.Context) error {
	session := sessions.Default(c)
	session.Clear()
	return session.Save()
}

func SetFlash(c *gin.Context, message string) error {
	session := sessions.Default(c)
	session.Set(FlashKey, message)
	return session.Save()
}

func PopFlash(c *gin.Context) string {
	session := sessions.Default(c)
	value := session.Get(FlashKey)
	if value == nil {
		return ""
	}

	session.Delete(FlashKey)
	_ = session.Save()

	message, _ := value.(string)
	return message
}

func CurrentUserID(c *gin.Context) (uint, bool) {
	value := sessions.Default(c).Get(UserIDKey)
	switch userID := value.(type) {
	case uint:
		return userID, true
	case int:
		return uint(userID), true
	case int64:
		return uint(userID), true
	case float64:
		return uint(userID), true
	default:
		return 0, false
	}
}
