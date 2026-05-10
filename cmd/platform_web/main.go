package main

import (
	"leadstorefront/pkgs"
	"leadstorefront/pkgs/middleware"
	"leadstorefront/platform_web/routes"
	"log"

	"github.com/gin-gonic/gin"
)

func main() {
	cfg := pkgs.Config.Web

	app := gin.Default()
	app.Use(middleware.Sessions())
	app.Static("/assets", "platform_web/static/assets")
	app.Static("/imgs", "platform_web/static/imgs")
	app.Static("/uploads", "uploads")
	app.LoadHTMLFiles(
		"platform_web/templates/base.gohtml",
		"platform_web/templates/partials/meta-tags.gohtml",
		"platform_web/templates/partials/navbar.gohtml",
		"platform_web/templates/partials/footer.gohtml",
		"platform_web/templates/partials/scripts.gohtml",
		"platform_web/templates/routes/home.gohtml",
		"platform_web/templates/routes/storefront_show.gohtml",
		"platform_web/templates/routes/articles_index.gohtml",
		"platform_web/templates/routes/article_show.gohtml",
		"platform_web/templates/routes/products_index.gohtml",
		"platform_web/templates/routes/product_show.gohtml",
		"platform_web/templates/routes/admin_home.gohtml",
		"platform_web/templates/routes/admin_login.gohtml",
		"platform_web/templates/routes/admin_register.gohtml",
		"platform_web/templates/routes/admin_storefronts_index.gohtml",
		"platform_web/templates/routes/admin_storefront_show.gohtml",
		"platform_web/templates/routes/admin_storefront_form.gohtml",
		"platform_web/templates/routes/admin_articles_index.gohtml",
		"platform_web/templates/routes/admin_article_form.gohtml",
		"platform_web/templates/routes/admin_products_index.gohtml",
		"platform_web/templates/routes/admin_product_form.gohtml",
	)
	routes.Register(app)

	log.Printf("platform_web serving at http://%s%s", cfg.Domain, cfg.Addr)
	if err := app.Run(cfg.Addr); err != nil {
		log.Fatal(err)
	}
}
