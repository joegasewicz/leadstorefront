package routes

import (
	"errors"
	"leadstorefront/pkgs/models"
	"leadstorefront/pkgs/utils"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	entityfileuploader "codeberg.org/joegasewicz/entity-file-uploader"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Article struct {
	DB *gorm.DB
}

func (article *Article) Get(c *gin.Context) {
	if c.Param("id") != "" {
		article.getByID(c)
		return
	}
	if c.Param("slug") != "" {
		article.getPublishedBySlug(c)
		return
	}
	if strings.HasPrefix(c.FullPath(), "/api/v1/admin/") {
		article.getAdminList(c)
		return
	}
	article.getPublishedList(c)
}

func (article *Article) Post(c *gin.Context) {
	record, ok := article.bindJSON(c)
	if !ok {
		return
	}
	if err := article.DB.Create(&record).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not create article"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"article": record})
}

func (article *Article) Put(c *gin.Context) {
	record, ok := article.bindJSON(c)
	if !ok {
		return
	}
	var existing models.Article
	if err := article.DB.First(&existing, c.Param("id")).Error; err != nil {
		utils.WriteRecordError(c, err, "could not load article")
		return
	}
	record.MainImage = existing.MainImage
	if err := article.DB.Model(&existing).Updates(articleUpdateMap(record)).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not update article"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"article": record})
}

func (article *Article) Delete(c *gin.Context) {
	var record models.Article
	if err := article.DB.First(&record, c.Param("id")).Error; err != nil {
		utils.WriteRecordError(c, err, "could not load article")
		return
	}
	if err := article.DB.Delete(&record).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not delete article"})
		return
	}
	_ = deleteArticleUploads(record)
	c.JSON(http.StatusOK, gin.H{"deleted": true})
}

func (article *Article) PostImage(c *gin.Context) {
	var record models.Article
	if err := article.DB.First(&record, c.Param("id")).Error; err != nil {
		utils.WriteRecordError(c, err, "could not load article")
		return
	}
	if uploaded, err := saveArticleMainImage(c, &record, record.MainImage); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	} else if uploaded {
		_ = article.DB.Save(&record).Error
	}
	c.JSON(http.StatusOK, gin.H{"article": record})
}

func (article *Article) getPublishedList(c *gin.Context) {
	var articles []models.Article
	query := publishedArticlesQuery(article.DB)
	if storefrontID := strings.TrimSpace(c.Query("storefront_id")); storefrontID != "" {
		query = query.Joins("JOIN article_storefronts ON article_storefronts.article_id = articles.id").Where("article_storefronts.storefront_id = ?", storefrontID)
	}
	if err := query.Find(&articles).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not load articles"})
		return
	}
	withArticleImageURLs(articles)
	c.JSON(http.StatusOK, gin.H{"articles": articles})
}

func (article *Article) getPublishedBySlug(c *gin.Context) {
	var record models.Article
	query := publishedArticlesQuery(article.DB).Where("slug = ?", c.Param("slug"))
	if storefrontID := strings.TrimSpace(c.Query("storefront_id")); storefrontID != "" {
		query = query.Joins("JOIN article_storefronts ON article_storefronts.article_id = articles.id").Where("article_storefronts.storefront_id = ?", storefrontID)
	}
	err := query.First(&record).Error
	if err != nil {
		utils.WriteRecordError(c, err, "could not load article")
		return
	}
	withArticleImageURL(&record)
	c.JSON(http.StatusOK, gin.H{"article": record})
}

func (article *Article) getAdminList(c *gin.Context) {
	var articles []models.Article
	var total int64
	page, limit, offset := utils.GetPagination(c)
	query := article.DB.Model(&models.Article{})
	if storefrontID := strings.TrimSpace(c.Query("storefront_id")); storefrontID != "" {
		query = query.Joins("JOIN article_storefronts ON article_storefronts.article_id = articles.id").Where("article_storefronts.storefront_id = ?", storefrontID)
	}
	if err := query.Count(&total).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not count articles"})
		return
	}
	listQuery := article.DB.Preload("ArticleCategory").Preload("Product").Order("created_at desc").Limit(limit).Offset(offset)
	if storefrontID := strings.TrimSpace(c.Query("storefront_id")); storefrontID != "" {
		listQuery = listQuery.Joins("JOIN article_storefronts ON article_storefronts.article_id = articles.id").Where("article_storefronts.storefront_id = ?", storefrontID)
	}
	if err := listQuery.Find(&articles).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not load articles"})
		return
	}
	withArticleImageURLs(articles)
	c.JSON(http.StatusOK, gin.H{"articles": articles, "pagination": utils.NewPagination(page, limit, total)})
}

func (article *Article) getByID(c *gin.Context) {
	var record models.Article
	err := article.DB.First(&record, c.Param("id")).Error
	if err != nil {
		utils.WriteRecordError(c, err, "could not load article")
		return
	}
	withArticleImageURL(&record)
	c.JSON(http.StatusOK, gin.H{"article": record})
}

func (article *Article) bindJSON(c *gin.Context) (models.Article, bool) {
	var record models.Article
	if err := c.ShouldBindJSON(&record); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid article"})
		return models.Article{}, false
	}
	record.Slug = utils.Slugify(record.Title)
	if record.IsPublished && record.PublishedAt == nil {
		record.PublishedAt = utils.PtrTime(time.Now())
	}
	return record, true
}

func publishedArticlesQuery(db *gorm.DB) *gorm.DB {
	now := time.Now()
	return db.Preload("ArticleCategory").Preload("Product").Where("is_published = ? AND (published_at IS NULL OR published_at <= ?)", true, now).Order("published_at desc, created_at desc")
}

func articleUpdateMap(article models.Article) map[string]interface{} {
	return map[string]interface{}{
		"author": article.Author, "title": article.Title, "slug": article.Slug,
		"subtitle": article.Subtitle, "body": article.Body, "main_image": article.MainImage,
		"image_url": article.ImageURL, "meta_title": article.MetaTitle,
		"meta_description": article.MetaDescription, "meta_keywords": article.MetaKeywords,
		"canonical_url": article.CanonicalURL, "is_published": article.IsPublished,
		"published_at": article.PublishedAt, "article_category_id": article.ArticleCategoryID,
		"product_id": article.ProductID,
	}
}

func saveArticleMainImage(c *gin.Context, article *models.Article, existingFileName string) (bool, error) {
	file, header, err := c.Request.FormFile("main_image")
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
	if !isSafeArticleImageName(header.Filename) {
		return false, errors.New("main image filename is invalid")
	}
	if !isAllowedArticleImage(header.Filename) {
		return false, errors.New("main image must be a JPG, JPEG, PNG, or WebP file")
	}
	manager, err := articleImageManager()
	if err != nil {
		return false, err
	}
	id := strconv.Itoa(int(article.ID))
	if _, err := manager.Upload(c.Writer, c.Request, id, "main_image"); err != nil {
		return false, err
	}
	if existingFileName != "" && existingFileName != header.Filename {
		_ = manager.Delete(existingFileName, id)
	}
	article.MainImage = header.Filename
	return true, nil
}

func articleImageManager() (*entityfileuploader.FileManager, error) {
	fileUpload := entityfileuploader.FileUpload{UploadDir: "uploads", MaxFileSize: 10, FileTypes: []string{"jpg", "jpeg", "png", "webp"}, URL: ""}
	return fileUpload.Init("articles")
}

func articleMainImageURL(article models.Article) string {
	if article.MainImage == "" || article.ID == 0 {
		return ""
	}
	manager, err := articleImageManager()
	if err != nil {
		return ""
	}
	return manager.Get(article.MainImage, strconv.Itoa(int(article.ID)))
}

func withArticleImageURLs(articles []models.Article) {
	for index := range articles {
		withArticleImageURL(&articles[index])
	}
}

func withArticleImageURL(article *models.Article) {
	if imageURL := articleMainImageURL(*article); imageURL != "" {
		article.ImageURL = imageURL
	}
}

func deleteArticleUploads(article models.Article) error {
	if article.ID == 0 {
		return nil
	}
	manager, err := articleImageManager()
	if err != nil {
		return err
	}
	return manager.DeleteEntityByID(strconv.Itoa(int(article.ID)))
}

func isAllowedArticleImage(fileName string) bool {
	extension := strings.TrimPrefix(strings.ToLower(filepath.Ext(fileName)), ".")
	switch extension {
	case "jpg", "jpeg", "png", "webp":
		return true
	default:
		return false
	}
}

func isSafeArticleImageName(fileName string) bool {
	return fileName == filepath.Base(fileName) && !strings.ContainsAny(fileName, `/\`)
}
