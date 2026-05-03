package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

const DefaultCountryCode = "uk"

var supportedCountryCodes = map[string]struct{}{
	"us": {},
	"uk": {},
	"au": {},
	"za": {},
	"nz": {},
	"ca": {},
	"ie": {},
	"sg": {},
}

func DetectCountryCode(c *gin.Context) string {
	for _, header := range []string{
		"CF-IPCountry",
		"X-Vercel-IP-Country",
		"X-Country-Code",
		"X-AppEngine-Country",
	} {
		if country := normalizeCountryCode(c.GetHeader(header)); country != "" {
			return country
		}
	}

	if country := countryFromAcceptLanguage(c.GetHeader("Accept-Language")); country != "" {
		return country
	}

	return DefaultCountryCode
}

func RedirectToLocalizedHome() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Redirect(http.StatusFound, "/"+DetectCountryCode(c))
	}
}

func IsSupportedCountryCode(value string) bool {
	_, ok := supportedCountryCodes[strings.ToLower(value)]
	return ok
}

func countryFromAcceptLanguage(value string) string {
	for _, language := range strings.Split(value, ",") {
		parts := strings.Split(strings.TrimSpace(language), ";")
		localeParts := strings.Split(parts[0], "-")
		if len(localeParts) < 2 {
			continue
		}

		if country := normalizeCountryCode(localeParts[len(localeParts)-1]); country != "" {
			return country
		}
	}

	return ""
}

func normalizeCountryCode(value string) string {
	country := strings.ToLower(strings.TrimSpace(value))
	if country == "gb" {
		country = "uk"
	}

	if IsSupportedCountryCode(country) {
		return country
	}

	return ""
}
