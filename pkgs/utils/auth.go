package utils

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"leadstorefront/pkgs"
	"strconv"
	"strings"
)

func SignedUserAuthToken(userID uint) string {
	message := strconv.FormatUint(uint64(userID), 10)
	return fmt.Sprintf("%s:%s", message, authSignature(message))
}

func UserIDFromAuthHeader(header string) (uint, bool) {
	token := strings.TrimSpace(header)
	token = strings.TrimSpace(strings.TrimPrefix(token, "Bearer "))
	parts := strings.Split(token, ":")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return 0, false
	}
	if !hmac.Equal([]byte(parts[1]), []byte(authSignature(parts[0]))) {
		return 0, false
	}
	parsed, err := strconv.ParseUint(parts[0], 10, 64)
	if err != nil || parsed == 0 {
		return 0, false
	}
	return uint(parsed), true
}

func authSignature(message string) string {
	mac := hmac.New(sha256.New, []byte(pkgs.Config.Session.Secret))
	mac.Write([]byte(message))
	return hex.EncodeToString(mac.Sum(nil))
}
