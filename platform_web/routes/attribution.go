package routes

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"leadstorefront/pkgs/utils"
	"net/http"
	"strings"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

const (
	attributionVisitorCookie = "lsf_visitor_id"
	attributionPayloadKey    = "outbound_attribution_payload"
	attributionLandingKey    = "outbound_attribution_landing_path"
	attributionStorefrontKey = "outbound_attribution_storefront_id"
)

func captureAttribution(c *gin.Context, storefrontID uint) {
	payload := utils.ParseAttribution(c.Request.URL.Query())
	if !payload.HasData() || storefrontID == 0 {
		return
	}
	_ = attributionVisitorID(c)
	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return
	}
	session := sessions.Default(c)
	session.Set(attributionPayloadKey, string(payloadJSON))
	session.Set(attributionLandingKey, c.Request.URL.RequestURI())
	session.Set(attributionStorefrontKey, storefrontID)
	_ = session.Save()
}

func currentOutboundAttribution(c *gin.Context, storefrontID uint) (utils.AttributionPayload, string, bool) {
	session := sessions.Default(c)
	if sessionUint(session.Get(attributionStorefrontKey)) != storefrontID {
		return utils.AttributionPayload{}, "", false
	}
	raw, _ := session.Get(attributionPayloadKey).(string)
	if strings.TrimSpace(raw) == "" {
		return utils.AttributionPayload{}, "", false
	}
	var payload utils.AttributionPayload
	if err := json.Unmarshal([]byte(raw), &payload); err != nil || !payload.HasData() {
		return utils.AttributionPayload{}, "", false
	}
	landingPath, _ := session.Get(attributionLandingKey).(string)
	return payload, landingPath, true
}

func attributionVisitorID(c *gin.Context) string {
	if cookie, err := c.Cookie(attributionVisitorCookie); err == nil {
		if visitorID := strings.TrimSpace(cookie); visitorID != "" {
			return visitorID
		}
	}
	visitorID := randomVisitorID()
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     attributionVisitorCookie,
		Value:    visitorID,
		Path:     "/",
		MaxAge:   int((90 * 24 * time.Hour).Seconds()),
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   c.Request.TLS != nil,
	})
	return visitorID
}

func randomVisitorID() string {
	var buffer [16]byte
	if _, err := rand.Read(buffer[:]); err != nil {
		return hex.EncodeToString([]byte(time.Now().UTC().Format(time.RFC3339Nano)))
	}
	return hex.EncodeToString(buffer[:])
}

func sessionUint(value interface{}) uint {
	switch typed := value.(type) {
	case uint:
		return typed
	case int:
		return uint(typed)
	case int64:
		return uint(typed)
	case float64:
		return uint(typed)
	default:
		return 0
	}
}
