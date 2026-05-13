package routes

import (
	"leadstorefront/pkgs/middleware"
	"leadstorefront/pkgs/models"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

type AdminLeadForms struct {
	API *APIClient
}

func (forms *AdminLeadForms) Get(c *gin.Context) {
	storefront, fields, ok := forms.load(c)
	if !ok {
		return
	}
	forms.render(c, http.StatusOK, storefront, ensureLeadFormRows(fields), "")
}

func (forms *AdminLeadForms) Post(c *gin.Context) {
	storefront, _, ok := forms.load(c)
	if !ok {
		return
	}
	fields := leadFormFieldsFromRequest(c, storefront.ID)
	if err := forms.API.Put(c, "/admin/storefronts/"+uintToString(storefront.ID)+"/lead-form", map[string]interface{}{"fields": fields}, nil); err != nil {
		forms.render(c, http.StatusBadRequest, storefront, ensureLeadFormRows(fields), "Could not save the lead form.")
		return
	}
	_ = middleware.SetFlash(c, "Lead form saved.")
	c.Redirect(http.StatusFound, "/admin/storefronts/"+uintToString(storefront.ID))
}

func (forms *AdminLeadForms) Put(c *gin.Context) {
	c.AbortWithStatus(http.StatusMethodNotAllowed)
}

func (forms *AdminLeadForms) Delete(c *gin.Context) {
	c.AbortWithStatus(http.StatusMethodNotAllowed)
}

func (forms *AdminLeadForms) load(c *gin.Context) (models.Storefront, []models.LeadFormField, bool) {
	id, ok := apiPathID(c.Param("id"))
	if !ok {
		c.AbortWithStatus(http.StatusNotFound)
		return models.Storefront{}, nil, false
	}
	var storefrontResponse struct {
		Storefront models.Storefront `json:"storefront"`
	}
	if err := forms.API.Get(c, "/admin/storefronts/"+id, &storefrontResponse); err != nil {
		c.String(http.StatusNotFound, "could not load storefront")
		return models.Storefront{}, nil, false
	}
	var formResponse struct {
		Fields []models.LeadFormField `json:"fields"`
	}
	_ = forms.API.Get(c, "/admin/storefronts/"+id+"/lead-form", &formResponse)
	return storefrontResponse.Storefront, formResponse.Fields, true
}

func (forms *AdminLeadForms) render(c *gin.Context, status int, storefront models.Storefront, fields []models.LeadFormField, message string) {
	c.HTML(status, "admin_lead_form", gin.H{
		"Title":        "Lead form",
		"Storefront":   storefront,
		"Fields":       fields,
		"Error":        message,
		"IsAdmin":      true,
		"IsSuper":      isCurrentSuper(c),
		"IsAdminRoute": true,
	})
}

func leadFormFieldsFromRequest(c *gin.Context, storefrontID uint) []models.LeadFormField {
	labels := c.PostFormArray("field_label")
	names := c.PostFormArray("field_name")
	types := c.PostFormArray("field_type")
	options := c.PostFormArray("field_options")
	required := c.PostFormArray("field_required")
	requiredSet := map[string]struct{}{}
	for _, index := range required {
		requiredSet[index] = struct{}{}
	}

	fields := make([]models.LeadFormField, 0, len(labels))
	for index, label := range labels {
		label = strings.TrimSpace(label)
		if label == "" {
			continue
		}
		field := models.LeadFormField{
			StorefrontID: storefrontID,
			Label:        label,
			Name:         formArrayValue(names, index),
			Type:         formArrayValue(types, index),
			Options:      formArrayValue(options, index),
			SortOrder:    index + 1,
		}
		if _, ok := requiredSet[uintToString(uint(index))]; ok {
			field.IsRequired = true
		}
		fields = append(fields, field)
	}
	return fields
}

func formArrayValue(values []string, index int) string {
	if index >= len(values) {
		return ""
	}
	return strings.TrimSpace(values[index])
}

func ensureLeadFormRows(fields []models.LeadFormField) []models.LeadFormField {
	for len(fields) < 5 {
		fields = append(fields, models.LeadFormField{Type: "text"})
	}
	return fields
}
