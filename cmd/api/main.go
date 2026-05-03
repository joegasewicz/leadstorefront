package main

import (
	"gadgetscout/api/routes"
	"gadgetscout/pkgs"
	"gadgetscout/pkgs/utils/database"
	"log"

	"github.com/gin-gonic/gin"
)

func main() {
	cfg := pkgs.Config.API

	db, err := database.NewPostgres()
	if err != nil {
		log.Fatal(err)
	}
	log.Println("postgres connection established")

	migrate := database.NewMigrate(db)
	log.Println("database migrations starting")
	if err := migrate.Run(); err != nil {
		log.Fatal(err)
	}

	app := gin.Default()
	routes.Register(app)

	log.Printf("api serving at http://%s%s", cfg.Domain, cfg.Addr)
	if err := app.Run(cfg.Addr); err != nil {
		log.Fatal(err)
	}
}
