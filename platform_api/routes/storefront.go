package routes

import (
	"errors"
	"leadstorefront/pkgs"
	"leadstorefront/pkgs/models"
	"leadstorefront/pkgs/utils"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	entityfileuploader "codeberg.org/joegasewicz/entity-file-uploader"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Storefront struct {
	DB *gorm.DB
}

func (storefront *Storefront) Get(c *gin.Context) {
	if strings.HasSuffix(c.FullPath(), "/create") {
		storefront.getOptions(c)
		return
	}
	if c.Param("domain") != "" {
		storefront.getActiveByDomain(c)
		return
	}
	if c.Param("id") != "" {
		storefront.getByID(c)
		return
	}
	storefront.getAdminList(c)
}

func (storefront *Storefront) getOptions(c *gin.Context) {
	user, ok := currentAPIUser(c, storefront.DB)
	if !ok {
		return
	}

	var countries []models.Country
	var users []models.User
	_ = storefront.DB.Order("name asc").Find(&countries).Error
	if isSuper(user) {
		_ = storefront.DB.Preload("Role").Order("email asc").Find(&users).Error
	} else {
		users = []models.User{user}
	}
	c.JSON(http.StatusOK, gin.H{"countries": countries, "users": users})
}

func (storefront *Storefront) Post(c *gin.Context) {
	if strings.Contains(c.FullPath(), "/nav-logo") {
		storefront.postNavLogo(c)
		return
	}
	if strings.Contains(c.FullPath(), "/products") {
		storefront.postProduct(c)
		return
	}
	if strings.Contains(c.FullPath(), "/articles") {
		storefront.postArticle(c)
		return
	}
	record, ok := storefront.bindJSON(c)
	if !ok {
		return
	}
	user, ok := currentAPIUser(c, storefront.DB)
	if !ok {
		return
	}
	if !isSuper(user) {
		record.OwnerID = &user.ID
	}
	restored, handled := storefront.restoreDeleted(c, record, user)
	if handled {
		if restored.ID != 0 {
			c.JSON(http.StatusCreated, gin.H{"storefront": restored})
		}
		return
	}
	if err := storefront.DB.Create(&record).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not create storefront"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"storefront": record})
}

func (storefront *Storefront) restoreDeleted(c *gin.Context, record models.Storefront, user models.User) (models.Storefront, bool) {
	var existing models.Storefront
	query := storefront.DB.Unscoped().Where("domain = ? AND deleted_at IS NOT NULL", record.Domain)
	if !isSuper(user) {
		query = query.Where("owner_id = ?", user.ID)
	}
	if err := query.First(&existing).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return models.Storefront{}, false
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not load storefront"})
		return models.Storefront{}, true
	}

	updates := storefrontUpdateMap(record)
	updates["deleted_at"] = nil
	if err := storefront.DB.Unscoped().Model(&existing).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not restore storefront"})
		return models.Storefront{}, true
	}
	if err := storefront.DB.Preload("PrimaryCountry").Preload("Owner.Role").First(&existing, existing.ID).Error; err != nil {
		utils.WriteRecordError(c, err, "could not load storefront")
		return models.Storefront{}, true
	}
	return existing, true
}

func (storefront *Storefront) postNavLogo(c *gin.Context) {
	storefrontID := uintPathID(c.Param("id"))
	if storefrontID == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	if ok := storefront.authorizeStorefrontID(c, storefrontID); !ok {
		return
	}

	var record models.Storefront
	if err := storefront.DB.First(&record, storefrontID).Error; err != nil {
		utils.WriteRecordError(c, err, "could not load storefront")
		return
	}
	changed := false
	if logoWidth := strings.TrimSpace(c.PostForm("logo_width_px")); logoWidth != "" {
		parsed, err := strconv.Atoi(logoWidth)
		if err != nil || parsed <= 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "nav logo width must be a positive whole number"})
			return
		}
		record.LogoWidthPx = parsed
		changed = true
	}
	if uploaded, err := saveStorefrontNavLogo(c, &record); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	} else if uploaded {
		changed = true
	}
	if changed {
		_ = storefront.DB.Save(&record).Error
	}
	c.JSON(http.StatusOK, gin.H{"storefront": record})
}

func (storefront *Storefront) postProduct(c *gin.Context) {
	var request struct {
		ProductID uint `json:"product_id"`
	}
	if err := c.ShouldBindJSON(&request); err != nil || request.ProductID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid product storefront"})
		return
	}
	record := models.ProductStorefront{ProductID: request.ProductID, StorefrontID: uintPathID(c.Param("id"))}
	if record.StorefrontID == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	if ok := storefront.authorizeStorefrontID(c, record.StorefrontID); !ok {
		return
	}
	if err := storefront.DB.Where("product_id = ? AND storefront_id = ?", record.ProductID, record.StorefrontID).FirstOrCreate(&record).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not assign product"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"product_storefront": record})
}

func (storefront *Storefront) postArticle(c *gin.Context) {
	var request struct {
		ArticleID uint `json:"article_id"`
	}
	if err := c.ShouldBindJSON(&request); err != nil || request.ArticleID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid article storefront"})
		return
	}
	record := models.ArticleStorefront{ArticleID: request.ArticleID, StorefrontID: uintPathID(c.Param("id"))}
	if record.StorefrontID == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	if ok := storefront.authorizeStorefrontID(c, record.StorefrontID); !ok {
		return
	}
	if err := storefront.DB.Where("article_id = ? AND storefront_id = ?", record.ArticleID, record.StorefrontID).FirstOrCreate(&record).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not assign article"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"article_storefront": record})
}

func (storefront *Storefront) Put(c *gin.Context) {
	record, ok := storefront.bindJSON(c)
	if !ok {
		return
	}
	user, ok := currentAPIUser(c, storefront.DB)
	if !ok {
		return
	}
	var existing models.Storefront
	query := storefront.DB
	if !isSuper(user) {
		query = query.Where("owner_id = ?", user.ID)
	}
	if err := query.First(&existing, c.Param("id")).Error; err != nil {
		utils.WriteRecordError(c, err, "could not load storefront")
		return
	}
	if !isSuper(user) {
		record.OwnerID = existing.OwnerID
	}
	if err := storefront.DB.Model(&existing).Updates(storefrontUpdateMap(record)).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not update storefront"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"storefront": existing})
}

func (storefront *Storefront) Delete(c *gin.Context) {
	user, ok := currentAPIUser(c, storefront.DB)
	if !ok {
		return
	}
	query := storefront.DB
	if !isSuper(user) {
		query = query.Where("owner_id = ?", user.ID)
	}
	result := query.Delete(&models.Storefront{}, c.Param("id"))
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not delete storefront"})
		return
	}
	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"deleted": true})
}

func (storefront *Storefront) getAdminList(c *gin.Context) {
	user, ok := currentAPIUser(c, storefront.DB)
	if !ok {
		return
	}
	var storefronts []models.Storefront
	var total int64
	page, limit, offset := utils.GetPagination(c)
	query := storefront.DB.Model(&models.Storefront{})
	if !isSuper(user) {
		query = query.Where("owner_id = ?", user.ID)
	}
	if err := query.Count(&total).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not count storefronts"})
		return
	}
	if err := query.Preload("PrimaryCountry").Preload("Owner.Role").Order("created_at desc").Limit(limit).Offset(offset).Find(&storefronts).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not load storefronts"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"storefronts": storefronts, "pagination": utils.NewPagination(page, limit, total)})
}

func (storefront *Storefront) getByID(c *gin.Context) {
	if !strings.HasPrefix(c.FullPath(), utils.APIVersion+"/admin/") {
		storefront.getActiveByID(c)
		return
	}
	user, ok := currentAPIUser(c, storefront.DB)
	if !ok {
		return
	}
	var record models.Storefront
	query := storefront.DB.Preload("PrimaryCountry").Preload("Owner.Role")
	if !isSuper(user) {
		query = query.Where("owner_id = ?", user.ID)
	}
	if err := query.First(&record, c.Param("id")).Error; err != nil {
		utils.WriteRecordError(c, err, "could not load storefront")
		return
	}
	c.JSON(http.StatusOK, gin.H{"storefront": record})
}

func (storefront *Storefront) getActiveByID(c *gin.Context) {
	var record models.Storefront
	err := storefront.DB.Preload("PrimaryCountry").Where("id = ? AND is_active = ?", c.Param("id"), true).First(&record).Error
	if err != nil {
		utils.WriteRecordError(c, err, "could not load storefront")
		return
	}
	c.JSON(http.StatusOK, gin.H{"storefront": record})
}

func (storefront *Storefront) authorizeStorefrontID(c *gin.Context, storefrontID uint) bool {
	user, ok := currentAPIUser(c, storefront.DB)
	if !ok {
		return false
	}
	if isSuper(user) {
		return true
	}
	var count int64
	if err := storefront.DB.Model(&models.Storefront{}).Where("id = ? AND owner_id = ?", storefrontID, user.ID).Count(&count).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not load storefront"})
		return false
	}
	if count == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return false
	}
	return true
}

func (storefront *Storefront) getActiveByDomain(c *gin.Context) {
	domain := strings.ToLower(strings.TrimSpace(c.Param("domain")))
	var record models.Storefront
	err := storefront.DB.Preload("PrimaryCountry").Where("domain IN ? AND is_active = ?", domainLookupCandidates(domain), true).First(&record).Error
	if err != nil {
		utils.WriteRecordError(c, err, "could not load storefront")
		return
	}
	c.JSON(http.StatusOK, gin.H{"storefront": record})
}

func domainLookupCandidates(domain string) []string {
	domain = strings.TrimPrefix(strings.TrimSuffix(domain, "."), "www.")
	if domain == "" {
		return []string{""}
	}
	return []string{domain, "www." + domain}
}

func (storefront *Storefront) bindJSON(c *gin.Context) (models.Storefront, bool) {
	var record models.Storefront
	if err := c.ShouldBindJSON(&record); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid storefront"})
		return models.Storefront{}, false
	}
	record.Name = strings.TrimSpace(record.Name)
	record.Domain = strings.ToLower(strings.TrimSpace(record.Domain))
	if record.Slug == "" {
		record.Slug = utils.Slugify(record.Name)
	} else {
		record.Slug = utils.Slugify(record.Slug)
	}
	return record, true
}

func storefrontUpdateMap(storefront models.Storefront) map[string]interface{} {
	return map[string]interface{}{
		"name":               storefront.Name,
		"slug":               storefront.Slug,
		"domain":             storefront.Domain,
		"description":        storefront.Description,
		"logo_url":           storefront.LogoURL,
		"logo_width_px":      storefront.LogoWidthPx,
		"hero_title":         storefront.HeroTitle,
		"hero_subtitle":      storefront.HeroSubtitle,
		"hero_image_url":     storefront.HeroImageURL,
		"about_title":        storefront.AboutTitle,
		"about_body":         storefront.AboutBody,
		"is_active":          storefront.IsActive,
		"primary_country_id": storefront.PrimaryCountryID,
		"owner_id":           storefront.OwnerID,
	}
}

func saveStorefrontNavLogo(c *gin.Context, storefront *models.Storefront) (bool, error) {
	file, header, err := c.Request.FormFile("nav_logo")
	if err != nil {
		if errors.Is(err, http.ErrMissingFile) {
			return false, nil
		}
		return false, err
	}
	_ = file.Close()
	if header.Filename == "" {
		return false, nil
	}
	if !isSafeStorefrontLogoName(header.Filename) {
		return false, errors.New("nav logo filename is invalid")
	}
	if !isAllowedStorefrontLogo(header.Filename) {
		return false, errors.New("nav logo must be a JPG, JPEG, PNG, SVG, or WebP file")
	}
	manager, err := storefrontLogoManager()
	if err != nil {
		return false, err
	}
	id := strconv.Itoa(int(storefront.ID))
	if _, err := manager.Upload(c.Writer, c.Request, id, "nav_logo"); err != nil {
		return false, err
	}
	storefront.LogoURL = manager.Get(header.Filename, id)
	return true, nil
}

func storefrontLogoManager() (*entityfileuploader.FileManager, error) {
	fileUpload := entityfileuploader.FileUpload{UploadDir: "uploads", MaxFileSize: 5, FileTypes: []string{"jpg", "jpeg", "png", "svg", "webp"}, URL: platformWebOrigin()}
	return fileUpload.Init("storefronts")
}

func platformWebOrigin() string {
	domain := strings.TrimSpace(pkgs.Config.Web.Domain)
	if domain == "" {
		domain = "localhost"
	}
	if strings.HasPrefix(domain, "http://") || strings.HasPrefix(domain, "https://") {
		return strings.TrimRight(domain, "/")
	}
	if domain == "localhost" || domain == "127.0.0.1" {
		return "http://" + domain + pkgs.Config.Web.Addr
	}
	return "https://" + domain
}

func isAllowedStorefrontLogo(fileName string) bool {
	extension := strings.TrimPrefix(strings.ToLower(filepath.Ext(fileName)), ".")
	switch extension {
	case "jpg", "jpeg", "png", "svg", "webp":
		return true
	default:
		return false
	}
}

func isSafeStorefrontLogoName(fileName string) bool {
	return fileName == filepath.Base(fileName) && !strings.ContainsAny(fileName, `/\`)
}

func uintPathID(value string) uint {
	parsed, err := strconv.Atoi(value)
	if err != nil || parsed <= 0 {
		return 0
	}
	return uint(parsed)
}
