package routes

import (
	"leadstorefront/pkgs/middleware"
	"leadstorefront/pkgs/models"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	form_validator "github.com/joegasewicz/form-validator"
)

type Admin struct {
	API         *APIClient
	LoginFields []form_validator.Field
	Roles       []string
}

func (admin *Admin) Auth(roles ...string) gin.HandlerFunc {
	return middleware.RequireRoles(roles...)
}

func (admin *Admin) Get(c *gin.Context) {
	switch c.FullPath() {
	case "/admin/login":
		admin.Login(c)
	case "/admin/register":
		admin.RegisterForm(c)
	default:
		admin.Home(c)
	}
}

func (admin *Admin) Post(c *gin.Context) {
	switch c.FullPath() {
	case "/admin/login":
		admin.LoginPost(c)
	case "/admin/register":
		admin.RegisterPost(c)
	default:
		c.AbortWithStatus(http.StatusMethodNotAllowed)
	}
}

func (admin *Admin) Put(c *gin.Context) {
	c.AbortWithStatus(http.StatusMethodNotAllowed)
}

func (admin *Admin) Delete(c *gin.Context) {
	_ = middleware.ClearUserSession(c)
	c.Redirect(http.StatusFound, "/admin/login")
}

func (admin *Admin) Home(c *gin.Context) {
	user, _ := c.Get("currentUser")
	currentUser := user.(models.User)

	c.HTML(http.StatusOK, "admin_home", gin.H{
		"Title":        "Admin",
		"Email":        currentUser.Email,
		"IsAdmin":      true,
		"IsAdminRoute": true,
	})
}

func (admin *Admin) Login(c *gin.Context) {
	c.HTML(http.StatusOK, "admin_login", gin.H{
		"Title":        "Admin Login",
		"IsAdminRoute": true,
	})
}

func (admin *Admin) LoginPost(c *gin.Context) {
	config := form_validator.Config{Fields: admin.loginFields()}
	if ok := form_validator.ValidateForm(c.Request, &config); !ok {
		renderAdminLogin(c, "Email and password are required.")
		return
	}
	email, _ := form_validator.GetString("email", &config)
	password, _ := form_validator.GetString("password", &config)

	var response struct {
		User models.User `json:"user"`
	}
	err := admin.API.Post(c, "/admin/login", map[string]interface{}{
		"email":    strings.ToLower(strings.TrimSpace(email)),
		"password": password,
	}, &response)
	if err != nil {
		renderAdminLogin(c, "Invalid email or password.")
		return
	}

	if err := middleware.SetUserSession(c, response.User.ID); err != nil {
		renderAdminLogin(c, "Could not create a session.")
		return
	}
	c.Redirect(http.StatusFound, "/admin")
}

func (admin *Admin) RegisterForm(c *gin.Context) {
	c.HTML(http.StatusOK, "admin_register", gin.H{
		"Title":        "Admin Register",
		"Roles":        admin.roles(),
		"IsAdminRoute": true,
	})
}

func (admin *Admin) RegisterPost(c *gin.Context) {
	config := form_validator.Config{Fields: admin.registerFields()}
	if ok := form_validator.ValidateForm(c.Request, &config); !ok {
		renderAdminRegister(c, "Email and password are required.")
		return
	}
	email, _ := form_validator.GetString("email", &config)
	password, _ := form_validator.GetString("password", &config)
	roleName, _ := form_validator.GetString("role", &config)
	email = strings.ToLower(strings.TrimSpace(email))
	roleName = strings.ToLower(strings.TrimSpace(roleName))

	if roleName != "admin" && roleName != "editor" {
		renderAdminRegister(c, "Select a valid admin role.")
		return
	}

	var response struct {
		User models.User `json:"user"`
	}
	if err := admin.API.Post(c, "/admin/register", map[string]interface{}{
		"email":    email,
		"password": password,
		"role":     roleName,
	}, &response); err != nil {
		renderAdminRegister(c, "Could not create the account.")
		return
	}

	if err := middleware.SetUserSession(c, response.User.ID); err != nil {
		renderAdminRegister(c, "Account created, but session creation failed.")
		return
	}

	c.Redirect(http.StatusFound, "/admin")
}

func (admin *Admin) loginFields() []form_validator.Field {
	if admin.LoginFields != nil {
		return admin.LoginFields
	}
	return []form_validator.Field{
		{Name: "email", Validate: true, Type: "string"},
		{Name: "password", Validate: true, Type: "string"},
	}
}

func (admin *Admin) registerFields() []form_validator.Field {
	return []form_validator.Field{
		{Name: "email", Validate: true, Type: "string"},
		{Name: "password", Validate: true, Type: "string"},
		{Name: "role", Validate: true, Type: "string"},
	}
}

func (admin *Admin) roles() []string {
	if admin.Roles != nil {
		return admin.Roles
	}
	return []string{"admin", "editor"}
}

func renderAdminLogin(c *gin.Context, message string) {
	c.HTML(http.StatusUnauthorized, "admin_login", gin.H{
		"Title":        "Admin Login",
		"Error":        message,
		"IsAdminRoute": true,
	})
}

func renderAdminRegister(c *gin.Context, message string) {
	c.HTML(http.StatusBadRequest, "admin_register", gin.H{
		"Title":        "Admin Register",
		"Error":        message,
		"Roles":        []string{"admin", "editor"},
		"IsAdminRoute": true,
	})
}
