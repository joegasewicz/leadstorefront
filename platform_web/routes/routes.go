package routes

import (
	"github.com/gin-gonic/gin"
)

func Register(app *gin.Engine) {
	api := NewAPIClient()
	health := &Health{}
	domainCheck := &DomainCheck{API: api}
	home := &Home{API: api}
	storefronts := &Storefronts{API: api}
	leads := &Leads{API: api}
	outboundClicks := &OutboundClicks{API: api}
	products := &Products{API: api}
	articles := &Articles{API: api}
	admin := &Admin{API: api}
	purchase := &Purchase{}
	adminStorefronts := &AdminStorefronts{API: api}
	adminArticles := &AdminArticles{API: api}
	adminProducts := &AdminProducts{API: api}
	adminUsers := &AdminUsers{API: api}
	adminLeadForms := &AdminLeadForms{API: api}
	adminAffiliateProviders := &AdminAffiliateProviders{API: api}

	app.GET("/health", health.Get)
	app.POST("/health", health.Post)
	app.PUT("/health", health.Put)
	app.DELETE("/health", health.Delete)

	app.GET("/domain-check", domainCheck.Get)
	app.POST("/domain-check", domainCheck.Post)
	app.PUT("/domain-check", domainCheck.Put)
	app.DELETE("/domain-check", domainCheck.Delete)

	app.GET("/", home.Redirect)
	app.GET("/storefronts/:id", storefronts.Get)
	app.GET("/storefronts/:id/out", outboundClicks.Get)
	app.POST("/storefronts/:id", leads.Post)
	app.GET("/storefronts/:id/articles", articles.Get)
	app.GET("/storefronts/:id/articles/:slug", articles.Get)
	app.GET("/storefronts/:id/products", products.Get)
	app.GET("/storefronts/:id/products/:slug", products.Get)
	app.GET("/register", admin.Get)
	app.POST("/register", admin.Post)
	app.GET("/purchase", admin.Auth("super", "admin", "editor", "user"), purchase.Get)
	app.POST("/purchase", admin.Auth("super", "admin", "editor", "user"), purchase.Post)
	app.GET("/payment", admin.Auth("super", "admin", "editor", "user"), purchase.Get)
	app.POST("/payment", admin.Auth("super", "admin", "editor", "user"), purchase.Post)
	app.GET("/:country/register", admin.Get)
	app.POST("/:country/register", admin.Post)
	app.GET("/:country", home.Get)
	app.GET("/:country/storefronts/:id", storefronts.Get)
	app.GET("/:country/storefronts/:id/out", outboundClicks.Get)
	app.POST("/:country/storefronts/:id", leads.Post)
	app.GET("/:country/storefronts/:id/articles", articles.Get)
	app.GET("/:country/storefronts/:id/articles/:slug", articles.Get)
	app.GET("/:country/storefronts/:id/products", products.Get)
	app.GET("/:country/storefronts/:id/products/:slug", products.Get)

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
	protected.Use(admin.Auth("super", "admin", "editor", "user"))
	protected.GET("", admin.Get)
	protected.PUT("", admin.Put)
	protected.GET("/storefronts", adminStorefronts.Get)
	protected.GET("/storefronts/create", adminStorefronts.Get)
	protected.POST("/storefronts/create", adminStorefronts.Post)
	protected.GET("/storefronts/:id/delete", adminStorefronts.Get)
	protected.POST("/storefronts/:id/delete", adminStorefronts.Delete)
	protected.GET("/storefronts/:id", adminStorefronts.Get)
	protected.GET("/storefronts/:id/affiliate-providers", adminAffiliateProviders.Get)
	protected.POST("/storefronts/:id/affiliate-providers", adminAffiliateProviders.Post)
	protected.POST("/storefronts/:id/affiliate-providers/:connection_id", adminAffiliateProviders.Post)
	protected.GET("/storefronts/:id/lead-form", adminLeadForms.Get)
	protected.POST("/storefronts/:id/lead-form", adminLeadForms.Post)
	protected.POST("/storefronts/:id/content", adminStorefronts.Post)
	protected.POST("/storefronts/:id/nav-logo", adminStorefronts.Post)
	protected.POST("/storefronts/:id/products", adminStorefronts.Post)
	protected.POST("/storefronts/:id/articles", adminStorefronts.Post)
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
	protected.GET("/leads", leads.Get)
	protected.GET("/leads/:id", leads.Get)
	protected.GET("/outbound-clicks", outboundClicks.Get)
	protected.GET("/affiliate-providers", adminAffiliateProviders.Get)

	superProtected := adminGroup.Group("")
	superProtected.Use(admin.Auth("super"))
	superProtected.GET("/users", adminUsers.Get)
	superProtected.GET("/users/:id/edit", adminUsers.Get)
	superProtected.POST("/users/:id/edit", adminUsers.Post)
	superProtected.POST("/users/:id/delete", adminUsers.Post)
}
