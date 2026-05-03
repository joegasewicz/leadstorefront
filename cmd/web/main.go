package main

import (
	"gadgetscout/pkgs"
	"gadgetscout/web/routes"
	"log"

	"github.com/gin-gonic/gin"
)

func main() {
	cfg := pkgs.Config.Web

	app := gin.Default()
	app.LoadHTMLFiles(
		"web/templates/base.gohtml",
		"web/templates/partials/meta-tags.gohtml",
		"web/templates/partials/navbar.gohtml",
		"web/templates/partials/footer.gohtml",
		"web/templates/partials/scripts.gohtml",
		"web/templates/routes/home.gohtml",
	)
	routes.Register(app)

	log.Printf("web serving at http://%s%s", cfg.Domain, cfg.Addr)
	if err := app.Run(cfg.Addr); err != nil {
		log.Fatal(err)
	}
}
