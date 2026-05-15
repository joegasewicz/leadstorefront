package routes

import (
	"encoding/json"
	"html/template"
	"leadstorefront/pkgs/models"
	"strings"
)

type storefrontDesignSectionView struct {
	ID             string
	Name           string
	Type           string
	Label          string
	Enabled        bool
	ContainerStyle string
	TextAlignments map[string]string
	InlineStyle    template.CSS
	ScopedCSS      template.CSS
	ContentKind    string
	CustomTitle    string
	Description    string
	Columns        []models.StorefrontDesignContentColumn
}

func storefrontDesignTemplateData(storefront models.Storefront) (models.StorefrontDesignConfig, string, []storefrontDesignSectionView) {
	design := models.StorefrontDesignFromJSON(storefront.DesignConfig)
	encoded, err := json.Marshal(design)
	if err != nil {
		encoded = []byte("{}")
	}
	sections := make([]storefrontDesignSectionView, 0, len(design.Sections))
	for _, section := range design.Sections {
		inlineStyle, scopedCSS := storefrontSectionStyles(section)
		sections = append(sections, storefrontDesignSectionView{
			ID:             section.ID,
			Name:           section.Name,
			Type:           section.Type,
			Label:          storefrontDesignSectionLabel(section),
			Enabled:        section.Enabled,
			ContainerStyle: section.ContainerStyle,
			TextAlignments: section.TextAlignments,
			InlineStyle:    inlineStyle,
			ScopedCSS:      scopedCSS,
			ContentKind:    section.Options.ContentKind,
			CustomTitle:    section.Options.Title,
			Description:    section.Options.Description,
			Columns:        section.Options.Columns,
		})
	}
	return design, string(encoded), sections
}

func storefrontSectionStyles(section models.StorefrontDesignSection) (template.CSS, template.CSS) {
	raw := strings.TrimSpace(section.ContainerStyle)
	var inlineStyle template.CSS
	var scopedCSS strings.Builder
	if raw != "" && !strings.Contains(raw, "{") {
		inlineStyle = template.CSS(raw)
	} else if raw != "" {
		scopedCSS.WriteString(scopeStorefrontSectionCSS(section.ID, raw))
	}
	scopedCSS.WriteString(scopeStorefrontAlignmentCSS(section.ID, section.TextAlignments))
	return inlineStyle, template.CSS(scopedCSS.String())
}

func scopeStorefrontSectionCSS(sectionID string, raw string) string {
	sectionID = strings.TrimSpace(sectionID)
	if sectionID == "" {
		return ""
	}
	scope := `[data-storefront-section="` + sectionID + `"]`
	var builder strings.Builder
	for _, block := range strings.Split(raw, "}") {
		block = strings.TrimSpace(block)
		if block == "" {
			continue
		}
		parts := strings.SplitN(block, "{", 2)
		if len(parts) != 2 {
			continue
		}
		selectors := make([]string, 0)
		for _, selector := range strings.Split(parts[0], ",") {
			selector = strings.TrimSpace(selector)
			if selector == "" || strings.HasPrefix(selector, "@") {
				continue
			}
			selectors = append(selectors, scope+" "+selector)
		}
		body := strings.TrimSpace(parts[1])
		if len(selectors) == 0 || body == "" {
			continue
		}
		builder.WriteString(strings.Join(selectors, ", "))
		builder.WriteString(" { ")
		builder.WriteString(body)
		builder.WriteString(" }\n")
	}
	return builder.String()
}

func scopeStorefrontAlignmentCSS(sectionID string, alignments map[string]string) string {
	sectionID = strings.TrimSpace(sectionID)
	if sectionID == "" || len(alignments) == 0 {
		return ""
	}
	scope := `[data-storefront-section="` + sectionID + `"]`
	var builder strings.Builder
	for _, tag := range []string{"h1", "h2", "h3", "h4", "h5", "h6", "p"} {
		align := alignments[tag]
		if align == "" {
			continue
		}
		builder.WriteString(scope)
		builder.WriteString(" ")
		builder.WriteString(tag)
		builder.WriteString(" { text-align: ")
		builder.WriteString(align)
		builder.WriteString("; }\n")
	}
	return builder.String()
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
