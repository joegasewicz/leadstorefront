package main

import (
	"gadgetscout/pkgs"
	"gadgetscout/pkgs/middleware"
	"gadgetscout/web/routes"
	"log"

	"github.com/gin-gonic/gin"
)

func main() {
	cfg := pkgs.Config.Web

	app := gin.Default()
	app.Use(middleware.Sessions())
	app.Static("/assets", "web/static/assets")
	app.Static("/uploads", "uploads")
	app.LoadHTMLFiles(
		"web/templates/base.gohtml",
		"web/templates/partials/meta-tags.gohtml",
		"web/templates/partials/navbar.gohtml",
		"web/templates/partials/footer.gohtml",
		"web/templates/partials/scripts.gohtml",
		"web/templates/routes/home.gohtml",
		"web/templates/routes/articles_index.gohtml",
		"web/templates/routes/article_show.gohtml",
		"web/templates/routes/products_index.gohtml",
		"web/templates/routes/product_show.gohtml",
		"web/templates/routes/admin_home.gohtml",
		"web/templates/routes/admin_login.gohtml",
		"web/templates/routes/admin_register.gohtml",
		"web/templates/routes/admin_articles_index.gohtml",
		"web/templates/routes/admin_article_form.gohtml",
		"web/templates/routes/admin_products_index.gohtml",
		"web/templates/routes/admin_product_form.gohtml",
	)
	routes.Register(app)

	log.Printf("web serving at http://%s%s", cfg.Domain, cfg.Addr)
	if err := app.Run(cfg.Addr); err != nil {
		log.Fatal(err)
	}
}
