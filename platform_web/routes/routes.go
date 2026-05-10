package routes

import (
	"github.com/gin-gonic/gin"
)

func Register(app *gin.Engine) {
	api := NewAPIClient()
	health := &Health{}
	home := &Home{API: api}
	storefronts := &Storefronts{API: api}
	products := &Products{API: api}
	articles := &Articles{API: api}
	admin := &Admin{API: api}
	adminArticles := &AdminArticles{API: api}
	adminProducts := &AdminProducts{API: api}

	app.GET("/health", health.Get)
	app.POST("/health", health.Post)
	app.PUT("/health", health.Put)
	app.DELETE("/health", health.Delete)

	app.GET("/", home.Redirect)
	app.GET("/:country", home.Get)
	app.GET("/:country/storefronts/:slug", storefronts.Get)

	app.GET("/:country/articles", articles.Get)
	app.GET("/:country/articles/:slug", articles.Get)

	app.GET("/:country/products", products.Get)
	app.GET("/:country/products/:slug", products.Get)

	adminGroup := app.Group("/admin")
	adminGroup.GET("/login", admin.Get)
	adminGroup.POST("/login", admin.Post)
	adminGroup.GET("/register", admin.Get)
	adminGroup.POST("/register", admin.Post)
	adminGroup.POST("/logout", admin.Delete)

	protected := adminGroup.Group("")
	protected.Use(admin.Auth("admin", "editor"))
	protected.GET("", admin.Get)
	protected.PUT("", admin.Put)
	protected.GET("/articles", adminArticles.Get)
	protected.GET("/articles/create", adminArticles.Get)
	protected.POST("/articles/create", adminArticles.Post)
	protected.GET("/articles/:id/edit", adminArticles.Get)
	protected.POST("/articles/:id/edit", adminArticles.Put)
	protected.POST("/articles/:id/delete", adminArticles.Delete)
	protected.GET("/products", adminProducts.Get)
	protected.GET("/products/create", adminProducts.Get)
	protected.POST("/products/create", adminProducts.Post)
	protected.GET("/products/:id/edit", adminProducts.Get)
	protected.POST("/products/:id/edit", adminProducts.Put)
	protected.POST("/products/:id/delete", adminProducts.Delete)
}
