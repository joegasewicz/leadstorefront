package routes

import (
	"encoding/json"
	"leadstorefront/pkgs/models"
	"leadstorefront/pkgs/utils"
	"net/http"
	"sort"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Lead struct {
	DB *gorm.DB
}

func (lead *Lead) Get(c *gin.Context) {
	switch {
	case strings.Contains(c.FullPath(), "/lead-form"):
		lead.getForm(c)
	case c.Param("id") != "":
		lead.getLead(c)
	default:
		lead.getAdminList(c)
	}
}

func (lead *Lead) Post(c *gin.Context) {
	var request struct {
		Source   string            `json:"source"`
		Tracking string            `json:"tracking"`
		Values   map[string]string `json:"values"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid lead"})
		return
	}
	storefrontID := leadStorefrontID(c)
	if storefrontID == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	if !lead.activeStorefrontExists(c, storefrontID) {
		return
	}

	var fields []models.LeadFormField
	if err := lead.DB.Where("storefront_id = ?", storefrontID).Order("sort_order asc, id asc").Find(&fields).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not load lead form"})
		return
	}
	if len(fields) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "lead form is not enabled"})
		return
	}
	if missing := missingRequiredLeadFields(fields, request.Values); len(missing) > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing required lead fields", "fields": missing})
		return
	}

	valuesJSON, err := json.Marshal(request.Values)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid lead values"})
		return
	}
	record := models.Lead{
		StorefrontID: storefrontID,
		Source:       strings.TrimSpace(request.Source),
		Tracking:     strings.TrimSpace(request.Tracking),
		ValuesJSON:   string(valuesJSON),
	}
	if err := lead.DB.Create(&record).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not create lead"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"lead": record})
}

func (lead *Lead) Put(c *gin.Context) {
	user, ok := currentAPIUser(c, lead.DB)
	if !ok {
		return
	}
	storefrontID := leadStorefrontID(c)
	if storefrontID == 0 || !lead.authorizeStorefront(c, user, storefrontID) {
		return
	}

	var request struct {
		Fields []models.LeadFormField `json:"fields"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid lead form"})
		return
	}
	fields, ok := normalizeLeadFormFields(c, storefrontID, request.Fields)
	if !ok {
		return
	}
	err := lead.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("storefront_id = ?", storefrontID).Delete(&models.LeadFormField{}).Error; err != nil {
			return err
		}
		if len(fields) == 0 {
			return nil
		}
		return tx.Create(&fields).Error
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not save lead form"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"fields": fields})
}

func (lead *Lead) Delete(c *gin.Context) {
	c.AbortWithStatus(http.StatusMethodNotAllowed)
}

func (lead *Lead) getForm(c *gin.Context) {
	storefrontID := leadStorefrontID(c)
	if storefrontID == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	if strings.HasPrefix(c.FullPath(), utils.APIVersion+"/admin/") {
		user, ok := currentAPIUser(c, lead.DB)
		if !ok || !lead.authorizeStorefront(c, user, storefrontID) {
			return
		}
	} else if !lead.activeStorefrontExists(c, storefrontID) {
		return
	}

	var fields []models.LeadFormField
	if err := lead.DB.Where("storefront_id = ?", storefrontID).Order("sort_order asc, id asc").Find(&fields).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not load lead form"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"fields": fields})
}

func (lead *Lead) getAdminList(c *gin.Context) {
	user, ok := currentAPIUser(c, lead.DB)
	if !ok {
		return
	}
	page, limit, offset := utils.GetPagination(c)
	var leads []models.Lead
	var total int64
	query := lead.DB.Model(&models.Lead{}).Joins("JOIN storefronts ON storefronts.id = leads.storefront_id").Preload("Storefront")
	if !isSuper(user) {
		query = query.Where("storefronts.owner_id = ?", user.ID)
	}
	if err := query.Count(&total).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not count leads"})
		return
	}
	if err := query.Order("leads.created_at desc").Limit(limit).Offset(offset).Find(&leads).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not load leads"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"leads": leads, "pagination": utils.NewPagination(page, limit, total)})
}

func (lead *Lead) getLead(c *gin.Context) {
	user, ok := currentAPIUser(c, lead.DB)
	if !ok {
		return
	}
	leadID := uintPathID(c.Param("id"))
	if leadID == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	var record models.Lead
	query := lead.DB.Preload("Storefront").Joins("JOIN storefronts ON storefronts.id = leads.storefront_id")
	if !isSuper(user) {
		query = query.Where("storefronts.owner_id = ?", user.ID)
	}
	if err := query.First(&record, leadID).Error; err != nil {
		utils.WriteRecordError(c, err, "could not load lead")
		return
	}
	c.JSON(http.StatusOK, gin.H{"lead": record})
}

func (lead *Lead) authorizeStorefront(c *gin.Context, user models.User, storefrontID uint) bool {
	if isSuper(user) {
		return true
	}
	var count int64
	if err := lead.DB.Model(&models.Storefront{}).Where("id = ? AND owner_id = ?", storefrontID, user.ID).Count(&count).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not load storefront"})
		return false
	}
	if count == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return false
	}
	return true
}

func (lead *Lead) activeStorefrontExists(c *gin.Context, storefrontID uint) bool {
	var count int64
	if err := lead.DB.Model(&models.Storefront{}).Where("id = ? AND is_active = ?", storefrontID, true).Count(&count).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not load storefront"})
		return false
	}
	if count == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return false
	}
	return true
}

func normalizeLeadFormFields(c *gin.Context, storefrontID uint, fields []models.LeadFormField) ([]models.LeadFormField, bool) {
	normalized := make([]models.LeadFormField, 0, len(fields))
	seen := map[string]struct{}{}
	for index, field := range fields {
		label := strings.TrimSpace(field.Label)
		fieldType := strings.ToLower(strings.TrimSpace(field.Type))
		if label == "" {
			continue
		}
		if !validLeadFieldType(fieldType) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid lead form field type"})
			return nil, false
		}
		name := utils.Slugify(field.Name)
		if name == "" {
			name = utils.Slugify(label)
		}
		if name == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid lead form field name"})
			return nil, false
		}
		if _, ok := seen[name]; ok {
			c.JSON(http.StatusBadRequest, gin.H{"error": "lead form field names must be unique"})
			return nil, false
		}
		seen[name] = struct{}{}
		normalized = append(normalized, models.LeadFormField{
			StorefrontID: storefrontID,
			Label:        label,
			Name:         name,
			Type:         fieldType,
			Options:      strings.TrimSpace(field.Options),
			IsRequired:   field.IsRequired,
			SortOrder:    index + 1,
		})
	}
	return normalized, true
}

func validLeadFieldType(fieldType string) bool {
	switch fieldType {
	case "text", "email", "tel", "textarea", "option":
		return true
	default:
		return false
	}
}

func missingRequiredLeadFields(fields []models.LeadFormField, values map[string]string) []string {
	missing := []string{}
	for _, field := range fields {
		if field.IsRequired && strings.TrimSpace(values[field.Name]) == "" {
			missing = append(missing, field.Name)
		}
	}
	sort.Strings(missing)
	return missing
}

func leadStorefrontID(c *gin.Context) uint {
	if id := uintPathID(c.Param("storefront_id")); id != 0 {
		return id
	}
	return uintPathID(c.Param("id"))
}
