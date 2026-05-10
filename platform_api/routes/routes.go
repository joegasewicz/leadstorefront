package routes

import (
	"leadstorefront/pkgs/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func Register(app *gin.Engine, db *gorm.DB) {
	country := &Country{DB: db}
	storefront := &Storefront{DB: db}
	product := &Product{DB: db}
	productCategory := &ProductCategory{DB: db}
	article := &Article{DB: db}
	articleCategory := &ArticleCategory{DB: db}
	user := &User{DB: db}

	app.GET(utils.GetVersion("/health"), health)
	app.GET(utils.GetVersion("/"), utils.APIRoot)
	app.GET(utils.GetVersion("/storefronts/:slug"), storefront.Get)
	app.GET(utils.GetVersion("/storefront-domains/:domain"), storefront.Get)
	app.GET(utils.GetVersion("/us"), country.Get)
	app.GET(utils.GetVersion("/uk"), country.Get)
	app.GET(utils.GetVersion("/au"), country.Get)
	app.GET(utils.GetVersion("/za"), country.Get)
	app.GET(utils.GetVersion("/nz"), country.Get)
	app.GET(utils.GetVersion("/ca"), country.Get)
	app.GET(utils.GetVersion("/ie"), country.Get)
	app.GET(utils.GetVersion("/sg"), country.Get)
	app.GET(utils.GetVersion("/us/products"), product.Get)
	app.GET(utils.GetVersion("/uk/products"), product.Get)
	app.GET(utils.GetVersion("/au/products"), product.Get)
	app.GET(utils.GetVersion("/za/products"), product.Get)
	app.GET(utils.GetVersion("/nz/products"), product.Get)
	app.GET(utils.GetVersion("/ca/products"), product.Get)
	app.GET(utils.GetVersion("/ie/products"), product.Get)
	app.GET(utils.GetVersion("/sg/products"), product.Get)
	app.GET(utils.GetVersion("/us/products/:slug"), product.Get)
	app.GET(utils.GetVersion("/uk/products/:slug"), product.Get)
	app.GET(utils.GetVersion("/au/products/:slug"), product.Get)
	app.GET(utils.GetVersion("/za/products/:slug"), product.Get)
	app.GET(utils.GetVersion("/nz/products/:slug"), product.Get)
	app.GET(utils.GetVersion("/ca/products/:slug"), product.Get)
	app.GET(utils.GetVersion("/ie/products/:slug"), product.Get)
	app.GET(utils.GetVersion("/sg/products/:slug"), product.Get)
	app.GET(utils.GetVersion("/us/articles"), article.Get)
	app.GET(utils.GetVersion("/uk/articles"), article.Get)
	app.GET(utils.GetVersion("/au/articles"), article.Get)
	app.GET(utils.GetVersion("/za/articles"), article.Get)
	app.GET(utils.GetVersion("/nz/articles"), article.Get)
	app.GET(utils.GetVersion("/ca/articles"), article.Get)
	app.GET(utils.GetVersion("/ie/articles"), article.Get)
	app.GET(utils.GetVersion("/sg/articles"), article.Get)
	app.GET(utils.GetVersion("/us/articles/:slug"), article.Get)
	app.GET(utils.GetVersion("/uk/articles/:slug"), article.Get)
	app.GET(utils.GetVersion("/au/articles/:slug"), article.Get)
	app.GET(utils.GetVersion("/za/articles/:slug"), article.Get)
	app.GET(utils.GetVersion("/nz/articles/:slug"), article.Get)
	app.GET(utils.GetVersion("/ca/articles/:slug"), article.Get)
	app.GET(utils.GetVersion("/ie/articles/:slug"), article.Get)
	app.GET(utils.GetVersion("/sg/articles/:slug"), article.Get)

	app.GET(utils.GetVersion("/admin"), utils.AdminHome)
	app.GET(utils.GetVersion("/admin/login"), utils.AdminStatic)
	app.POST(utils.GetVersion("/admin/login"), user.Post)
	app.GET(utils.GetVersion("/admin/register"), utils.AdminStatic)
	app.POST(utils.GetVersion("/admin/register"), user.Post)
	app.POST(utils.GetVersion("/admin/logout"), utils.AdminStatic)
	app.GET(utils.GetVersion("/admin/users/:id"), user.Get)

	app.GET(utils.GetVersion("/admin/storefronts"), storefront.Get)
	app.GET(utils.GetVersion("/admin/storefronts/create"), storefront.Get)
	app.GET(utils.GetVersion("/admin/storefronts/:id"), storefront.Get)
	app.POST(utils.GetVersion("/admin/storefronts/create"), storefront.Post)
	app.PUT(utils.GetVersion("/admin/storefronts/:id/edit"), storefront.Put)
	app.DELETE(utils.GetVersion("/admin/storefronts/:id/delete"), storefront.Delete)

	app.GET(utils.GetVersion("/admin/products"), product.Get)
	app.GET(utils.GetVersion("/admin/products/options"), productCategory.Get)
	app.GET(utils.GetVersion("/admin/products/create"), productCategory.Get)
	app.GET(utils.GetVersion("/admin/products/:id"), product.Get)
	app.POST(utils.GetVersion("/admin/products/create"), product.Post)
	app.GET(utils.GetVersion("/admin/products/:id/edit"), product.Get)
	app.PUT(utils.GetVersion("/admin/products/:id/edit"), product.Put)
	app.DELETE(utils.GetVersion("/admin/products/:id/delete"), product.Delete)

	app.GET(utils.GetVersion("/admin/articles"), article.Get)
	app.GET(utils.GetVersion("/admin/articles/options"), articleCategory.Get)
	app.GET(utils.GetVersion("/admin/articles/create"), articleCategory.Get)
	app.GET(utils.GetVersion("/admin/articles/:id"), article.Get)
	app.POST(utils.GetVersion("/admin/articles/create"), article.Post)
	app.GET(utils.GetVersion("/admin/articles/:id/edit"), article.Get)
	app.PUT(utils.GetVersion("/admin/articles/:id/edit"), article.Put)
	app.POST(utils.GetVersion("/admin/articles/:id/main-image"), article.PostImage)
	app.DELETE(utils.GetVersion("/admin/articles/:id/delete"), article.Delete)
}
