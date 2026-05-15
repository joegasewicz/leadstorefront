package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"strings"
)

const (
	StorefrontSectionHero    = "hero"
	StorefrontSectionFooter  = "footer"
	StorefrontSectionContent = "content"
)

type StorefrontDesignConfig struct {
	Colors   StorefrontDesignColors    `json:"colors"`
	Sections []StorefrontDesignSection `json:"sections"`
}

type StorefrontDesignColors struct {
	Primary    string `json:"primary"`
	Accent     string `json:"accent"`
	Background string `json:"background"`
	Text       string `json:"text"`
	Surface    string `json:"surface"`
}

type StorefrontDesignSection struct {
	ID             string                         `json:"id"`
	Name           string                         `json:"name"`
	Type           string                         `json:"type"`
	Enabled        bool                           `json:"enabled"`
	ContainerStyle string                         `json:"container_style,omitempty"`
	TextAlignments map[string]string              `json:"text_alignments,omitempty"`
	Options        StorefrontDesignSectionOptions `json:"options"`
}

type StorefrontDesignSectionOptions struct {
	ContentKind string                          `json:"content_kind,omitempty"`
	Title       string                          `json:"title,omitempty"`
	Description string                          `json:"description,omitempty"`
	Columns     []StorefrontDesignContentColumn `json:"columns,omitempty"`
}

type StorefrontDesignContentColumn struct {
	Heading string `json:"heading"`
	Body    string `json:"body"`
}

type JSONB []byte

func (j JSONB) Value() (driver.Value, error) {
	if len(j) == 0 {
		return "{}", nil
	}
	return string(j), nil
}

func (j *JSONB) Scan(value interface{}) error {
	if value == nil {
		*j = JSONB([]byte("{}"))
		return nil
	}
	switch typed := value.(type) {
	case []byte:
		*j = append((*j)[0:0], typed...)
	case string:
		*j = append((*j)[0:0], typed...)
	default:
		return fmt.Errorf("unsupported JSONB value %T", value)
	}
	return nil
}

func (j JSONB) MarshalJSON() ([]byte, error) {
	if len(j) == 0 {
		return []byte("{}"), nil
	}
	return j, nil
}

func (j *JSONB) UnmarshalJSON(value []byte) error {
	if len(value) == 0 {
		*j = JSONB([]byte("{}"))
		return nil
	}
	*j = append((*j)[0:0], value...)
	return nil
}

func DefaultStorefrontDesignConfig() StorefrontDesignConfig {
	return StorefrontDesignConfig{
		Colors: StorefrontDesignColors{
			Primary:    "#67e8f9",
			Accent:     "#38bdf8",
			Background: "#020617",
			Text:       "#ffffff",
			Surface:    "#0f172a",
		},
		Sections: []StorefrontDesignSection{
			{ID: "hero", Name: "Hero", Type: StorefrontSectionHero, Enabled: true},
			{ID: "lead-form", Name: "Lead form", Type: StorefrontSectionContent, Enabled: true, Options: StorefrontDesignSectionOptions{ContentKind: "lead_form"}},
			{ID: "about", Name: "About", Type: StorefrontSectionContent, Enabled: true, Options: StorefrontDesignSectionOptions{ContentKind: "about"}},
			{ID: "products", Name: "Products", Type: StorefrontSectionContent, Enabled: true, Options: StorefrontDesignSectionOptions{ContentKind: "products"}},
			{ID: "articles", Name: "Articles", Type: StorefrontSectionContent, Enabled: true, Options: StorefrontDesignSectionOptions{ContentKind: "articles"}},
			{ID: "footer", Name: "Footer", Type: StorefrontSectionFooter, Enabled: true},
		},
	}
}

func StorefrontDesignFromJSON(raw JSONB) StorefrontDesignConfig {
	var config StorefrontDesignConfig
	if len(raw) > 0 {
		_ = json.Unmarshal(raw, &config)
	}
	return NormalizeStorefrontDesignConfig(config)
}

func StorefrontDesignJSON(raw JSONB) JSONB {
	return StorefrontDesignToJSON(StorefrontDesignFromJSON(raw))
}

func StorefrontDesignToJSON(config StorefrontDesignConfig) JSONB {
	normalized := NormalizeStorefrontDesignConfig(config)
	encoded, err := json.Marshal(normalized)
	if err != nil {
		return JSONB([]byte("{}"))
	}
	return JSONB(encoded)
}

func StorefrontDesignFromJSONString(raw string) JSONB {
	var config StorefrontDesignConfig
	if strings.TrimSpace(raw) != "" {
		_ = json.Unmarshal([]byte(raw), &config)
	}
	return StorefrontDesignToJSON(config)
}

func NormalizeStorefrontDesignConfig(config StorefrontDesignConfig) StorefrontDesignConfig {
	defaults := DefaultStorefrontDesignConfig()
	config.Colors.Primary = normalizedHexColor(config.Colors.Primary, defaults.Colors.Primary)
	config.Colors.Accent = normalizedHexColor(config.Colors.Accent, defaults.Colors.Accent)
	config.Colors.Background = normalizedHexColor(config.Colors.Background, defaults.Colors.Background)
	config.Colors.Text = normalizedHexColor(config.Colors.Text, defaults.Colors.Text)
	config.Colors.Surface = normalizedHexColor(config.Colors.Surface, defaults.Colors.Surface)

	sections := make([]StorefrontDesignSection, 0, len(config.Sections))
	seen := map[string]bool{}
	for _, section := range config.Sections {
		normalized, ok := normalizedStorefrontSection(section)
		if !ok || seen[normalized.ID] {
			continue
		}
		sections = append(sections, normalized)
		seen[normalized.ID] = true
	}
	if len(sections) == 0 {
		for _, section := range defaults.Sections {
			sections = append(sections, section)
		}
	}
	config.Sections = sections
	return config
}

func normalizedStorefrontSection(section StorefrontDesignSection) (StorefrontDesignSection, bool) {
	sectionType := normalizedStorefrontSectionType(section.Type)
	contentKind := normalizedStorefrontContentKind(section.Options.ContentKind)
	if sectionType == "" {
		sectionType = legacySectionType(section.Type)
		contentKind = legacyContentKind(section.Type)
	}
	if sectionType == "" {
		return StorefrontDesignSection{}, false
	}
	if sectionType == StorefrontSectionContent && contentKind == "" {
		contentKind = "custom"
	}
	section.ID = normalizedSectionID(section.ID, section.Name, sectionType, contentKind)
	section.Name = normalizedSectionName(section.Name, sectionType, contentKind)
	section.Type = sectionType
	section.ContainerStyle = normalizedContainerStyle(section.ContainerStyle)
	section.TextAlignments = normalizedTextAlignments(section.TextAlignments)
	section.Options = normalizedStorefrontSectionOptions(section.Options, section.Name, sectionType, contentKind)
	return section, true
}

func normalizedContainerStyle(value string) string {
	value = strings.TrimSpace(value)
	if len(value) > 4000 {
		return value[:4000]
	}
	return value
}

func normalizedTextAlignments(alignments map[string]string) map[string]string {
	normalized := map[string]string{}
	for _, tag := range []string{"h1", "h2", "h3", "h4", "h5", "h6", "p"} {
		switch strings.ToLower(strings.TrimSpace(alignments[tag])) {
		case "left":
			normalized[tag] = "left"
		case "center":
			normalized[tag] = "center"
		case "right":
			normalized[tag] = "right"
		}
	}
	if len(normalized) == 0 {
		return nil
	}
	return normalized
}

func normalizedStorefrontSectionOptions(options StorefrontDesignSectionOptions, sectionName string, sectionType string, contentKind string) StorefrontDesignSectionOptions {
	if sectionType != StorefrontSectionContent {
		return StorefrontDesignSectionOptions{}
	}
	options.ContentKind = contentKind
	if contentKind != "custom" {
		options.Title = ""
		options.Description = ""
		options.Columns = nil
		return options
	}
	options.Title = strings.Join(strings.Fields(strings.TrimSpace(options.Title)), " ")
	if options.Title == "" {
		options.Title = sectionName
	}
	if len(options.Title) > 120 {
		options.Title = options.Title[:120]
	}
	options.Description = strings.TrimSpace(options.Description)
	if len(options.Description) > 1000 {
		options.Description = options.Description[:1000]
	}

	columns := make([]StorefrontDesignContentColumn, 0, len(options.Columns))
	for _, column := range options.Columns {
		column.Heading = strings.Join(strings.Fields(strings.TrimSpace(column.Heading)), " ")
		column.Body = strings.TrimSpace(column.Body)
		if column.Heading == "" && column.Body == "" {
			continue
		}
		if len(column.Heading) > 120 {
			column.Heading = column.Heading[:120]
		}
		if len(column.Body) > 2000 {
			column.Body = column.Body[:2000]
		}
		columns = append(columns, column)
		if len(columns) == 6 {
			break
		}
	}
	options.Columns = columns
	return options
}

func normalizedStorefrontSectionType(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case StorefrontSectionHero:
		return StorefrontSectionHero
	case StorefrontSectionFooter:
		return StorefrontSectionFooter
	case StorefrontSectionContent:
		return StorefrontSectionContent
	default:
		return ""
	}
}

func legacySectionType(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "lead_form", "about", "products", "articles":
		return StorefrontSectionContent
	default:
		return ""
	}
}

func legacyContentKind(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "lead_form":
		return "lead_form"
	case "about":
		return "about"
	case "products":
		return "products"
	case "articles":
		return "articles"
	default:
		return ""
	}
}

func normalizedStorefrontContentKind(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "lead_form":
		return "lead_form"
	case "about":
		return "about"
	case "products":
		return "products"
	case "articles":
		return "articles"
	case "custom":
		return "custom"
	default:
		return ""
	}
}

func normalizedSectionID(id string, name string, sectionType string, contentKind string) string {
	id = strings.ToLower(strings.TrimSpace(id))
	var builder strings.Builder
	for _, r := range id {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			builder.WriteRune(r)
		}
	}
	id = strings.Trim(builder.String(), "-")
	if id != "" {
		return id
	}
	if contentKind != "" {
		return strings.ReplaceAll(contentKind, "_", "-")
	}
	if sectionType != "" {
		return sectionType
	}
	return normalizedSectionName(name, sectionType, contentKind)
}

func normalizedSectionName(name string, sectionType string, contentKind string) string {
	name = strings.Join(strings.Fields(strings.TrimSpace(name)), " ")
	if name != "" {
		if len(name) > 80 {
			return name[:80]
		}
		return name
	}
	switch contentKind {
	case "lead_form":
		return "Lead form"
	case "about":
		return "About"
	case "products":
		return "Products"
	case "articles":
		return "Articles"
	case "custom":
		return "Content"
	}
	switch sectionType {
	case StorefrontSectionHero:
		return "Hero"
	case StorefrontSectionFooter:
		return "Footer"
	default:
		return "Section"
	}
}

func normalizedHexColor(value string, fallback string) string {
	value = strings.TrimSpace(value)
	if len(value) != 7 || value[0] != '#' {
		return fallback
	}
	for _, r := range value[1:] {
		if !((r >= '0' && r <= '9') || (r >= 'a' && r <= 'f') || (r >= 'A' && r <= 'F')) {
			return fallback
		}
	}
	return strings.ToLower(value)
}
