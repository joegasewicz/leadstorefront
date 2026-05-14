package routes

import (
	"encoding/json"
	"leadstorefront/pkgs/models"
)

type storefrontDesignSectionView struct {
	ID          string
	Name        string
	Type        string
	Label       string
	Enabled     bool
	ContentKind string
	CustomTitle string
	Description string
	Columns     []models.StorefrontDesignContentColumn
}

func storefrontDesignTemplateData(storefront models.Storefront) (models.StorefrontDesignConfig, string, []storefrontDesignSectionView) {
	design := models.StorefrontDesignFromJSON(storefront.DesignConfig)
	encoded, err := json.Marshal(design)
	if err != nil {
		encoded = []byte("{}")
	}
	sections := make([]storefrontDesignSectionView, 0, len(design.Sections))
	for _, section := range design.Sections {
		sections = append(sections, storefrontDesignSectionView{
			ID:          section.ID,
			Name:        section.Name,
			Type:        section.Type,
			Label:       storefrontDesignSectionLabel(section),
			Enabled:     section.Enabled,
			ContentKind: section.Options.ContentKind,
			CustomTitle: section.Options.Title,
			Description: section.Options.Description,
			Columns:     section.Options.Columns,
		})
	}
	return design, string(encoded), sections
}

func storefrontDesignSectionLabel(section models.StorefrontDesignSection) string {
	if section.Name != "" {
		return section.Name
	}
	switch section.Type {
	case models.StorefrontSectionHero:
		return "Hero"
	case models.StorefrontSectionFooter:
		return "Footer"
	case models.StorefrontSectionContent:
		switch section.Options.ContentKind {
		case "lead_form":
			return "Lead form"
		case "about":
			return "About"
		case "products":
			return "Products"
		case "articles":
			return "Articles"
		default:
			return "Content"
		}
	default:
		return section.Type
	}
}
