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
		&models.ArticleCategory{},
		&models.Article{},
	); err != nil {
		return err
	}

	return Seed(m.DB)
}
