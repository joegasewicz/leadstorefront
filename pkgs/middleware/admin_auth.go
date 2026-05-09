package middleware

import (
	"encoding/json"
	"fmt"
	"leadstorefront/pkgs"
	"leadstorefront/pkgs/models"
	"leadstorefront/pkgs/utils"
	"net/http"

	"github.com/gin-gonic/gin"
	identity_client "github.com/joegasewicz/identity-client"
)

func RequireRoles(allowedRoles ...string) gin.HandlerFunc {
	allowed := make(map[string]struct{}, len(allowedRoles))
	for _, role := range allowedRoles {
		allowed[role] = struct{}{}
	}

	return func(c *gin.Context) {
		userID, ok := CurrentUserID(c)
		if !ok {
			c.Redirect(http.StatusFound, "/admin/login")
			c.Abort()
			return
		}

		var response struct {
			User models.User `json:"user"`
		}
		path := utils.GetVersion(fmt.Sprintf("/admin/users/%d", userID))
		client := identity_client.Identity{URL: fmt.Sprintf("http://%s%s%s", pkgs.Config.API.Domain, pkgs.Config.API.Addr, path)}
		data, err := client.Get(c.Request)
		if err != nil {
			c.Redirect(http.StatusFound, "/admin/login")
			c.Abort()
			return
		}
		body, _ := json.Marshal(data)
		if err := json.Unmarshal(body, &response); err != nil {
			c.Redirect(http.StatusFound, "/admin/login")
			c.Abort()
			return
		}
		user := response.User

		if _, ok := allowed[user.Role.Name]; !ok {
			c.AbortWithStatus(http.StatusForbidden)
			return
		}

		c.Set("currentUser", user)
		c.Next()
	}
}
