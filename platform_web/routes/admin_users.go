package routes

import (
	"leadstorefront/pkgs/middleware"
	"leadstorefront/pkgs/models"
	"leadstorefront/pkgs/utils"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	form_validator "github.com/joegasewicz/form-validator"
)

type AdminUsers struct {
	API    *APIClient
	Fields []form_validator.Field
}

func (users *AdminUsers) Get(c *gin.Context) {
	if strings.HasSuffix(c.FullPath(), "/edit") {
		users.Edit(c)
		return
	}
	users.Index(c)
}

func (users *AdminUsers) Post(c *gin.Context) {
	if strings.HasSuffix(c.FullPath(), "/delete") {
		users.Delete(c)
		return
	}
	users.Update(c)
}

func (users *AdminUsers) Put(c *gin.Context) {
	c.AbortWithStatus(http.StatusMethodNotAllowed)
}

func (users *AdminUsers) Index(c *gin.Context) {
	var response struct {
		Users      []models.User    `json:"users"`
		Pagination utils.Pagination `json:"pagination"`
	}
	page, limit := utils.GetPaginationQuery(c)
	if err := users.API.Get(c, "/admin/users?page="+page+"&limit="+limit, &response); err != nil {
		c.String(http.StatusInternalServerError, "could not load users")
		return
	}

	c.HTML(http.StatusOK, "admin_users_index", gin.H{
		"Title":        "Users",
		"Users":        response.Users,
		"Pagination":   response.Pagination,
		"Limit":        limit,
		"Flash":        middleware.PopFlash(c),
		"IsAdmin":      true,
		"IsSuper":      true,
		"IsAdminRoute": true,
	})
}

func (users *AdminUsers) Edit(c *gin.Context) {
	id, ok := apiPathID(c.Param("id"))
	if !ok {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}
	user, ok := users.find(c, id)
	if !ok {
		return
	}
	users.renderForm(c, http.StatusOK, user, "")
}

func (users *AdminUsers) Update(c *gin.Context) {
	id, ok := apiPathID(c.Param("id"))
	if !ok {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}
	user, ok := users.find(c, id)
	if !ok {
		return
	}

	config := form_validator.Config{Fields: users.fields()}
	if ok := form_validator.ValidateForm(c.Request, &config); !ok {
		users.renderForm(c, http.StatusBadRequest, user, "Email is required.")
		return
	}
	email, _ := form_validator.GetString("email", &config)
	password := strings.TrimSpace(c.PostForm("password"))
	email = strings.ToLower(strings.TrimSpace(email))
	if email == "" {
		users.renderForm(c, http.StatusBadRequest, user, "Email is required.")
		return
	}

	if err := users.API.Put(c, "/admin/users/"+id, map[string]interface{}{"email": email, "password": password}, nil); err != nil {
		user.Email = email
		users.renderForm(c, http.StatusBadRequest, user, "Could not update the user.")
		return
	}

	_ = middleware.SetFlash(c, "User updated.")
	c.Redirect(http.StatusFound, "/admin/users")
}

func (users *AdminUsers) Delete(c *gin.Context) {
	id, ok := apiPathID(c.Param("id"))
	if !ok {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}
	if c.PostForm("confirm_delete") != "on" {
		_ = middleware.SetFlash(c, "Confirm that you want to delete the user.")
		c.Redirect(http.StatusFound, "/admin/users")
		return
	}

	if err := users.API.Delete(c, "/admin/users/"+id+"/delete", nil); err != nil {
		_ = middleware.SetFlash(c, "Could not delete the user.")
		c.Redirect(http.StatusFound, "/admin/users")
		return
	}

	_ = middleware.SetFlash(c, "User deleted.")
	c.Redirect(http.StatusFound, "/admin/users")
}

func (users *AdminUsers) find(c *gin.Context, id string) (models.User, bool) {
	var response struct {
		User models.User `json:"user"`
	}
	if err := users.API.Get(c, "/admin/users/"+id, &response); err != nil {
		c.String(http.StatusNotFound, "could not load user")
		return models.User{}, false
	}
	return response.User, true
}

func (users *AdminUsers) renderForm(c *gin.Context, status int, user models.User, message string) {
	c.HTML(status, "admin_user_form", gin.H{
		"Title":        "Edit user",
		"User":         user,
		"Error":        message,
		"IsAdmin":      true,
		"IsSuper":      true,
		"IsAdminRoute": true,
	})
}

func (users *AdminUsers) fields() []form_validator.Field {
	if users.Fields != nil {
		return users.Fields
	}
	return []form_validator.Field{{Name: "email", Validate: true, Type: "string"}}
}
