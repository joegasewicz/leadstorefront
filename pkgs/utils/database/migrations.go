package database

import (
	"leadstorefront/pkgs/models"

	"gorm.io/gorm"
)

type Migrate struct {
	DB *gorm.DB
}

func NewMigrate(db *gorm.DB) *Migrate {
	return &Migrate{
		DB: db,
	}
}

func (m *Migrate) Run() error {
	if err := m.DB.AutoMigrate(
		&models.Country{},
		&models.Role{},
		&models.User{},
		&models.Storefront{},
		&models.ProductCategory{},
		&models.Product{},
		&models.ProductStorefront{},
		&models.ArticleCategory{},
		&models.Article{},
		&models.ArticleStorefront{},
	); err != nil {
		return err
	}
	if err := m.dropReplacedStorefrontColumns(); err != nil {
		return err
	}

	return Seed(m.DB)
}

func (m *Migrate) dropReplacedStorefrontColumns() error {
	if err := m.DB.Exec("ALTER TABLE products DROP COLUMN IF EXISTS storefront_id").Error; err != nil {
		return err
	}
	if err := m.DB.Exec("ALTER TABLE articles DROP COLUMN IF EXISTS storefront_id").Error; err != nil {
		return err
	}
	return nil
}
