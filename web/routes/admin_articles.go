package routes

import (
	"gadgetscout/pkgs/middleware"
	"gadgetscout/pkgs/models"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	form_validator "github.com/joegasewicz/form-validator"
)

type AdminArticles struct {
	API        *APIClient
	FormFields []form_validator.Field
}

func (articles *AdminArticles) Get(c *gin.Context) {
	switch {
	case strings.HasSuffix(c.FullPath(), "/create"):
		articles.Create(c)
	case strings.HasSuffix(c.FullPath(), "/edit"):
		articles.Edit(c)
	default:
		articles.Index(c)
	}
}

func (articles *AdminArticles) Index(c *gin.Context) {
	var response struct {
		Articles []models.Article `json:"articles"`
	}
	if err := articles.API.Get(c, "/admin/articles", &response); err != nil {
		c.String(http.StatusInternalServerError, "could not load articles")
		return
	}

	c.HTML(http.StatusOK, "admin_articles_index", gin.H{
		"Title":          "Articles",
		"Articles":       response.Articles,
		"DefaultCountry": middleware.DefaultCountryCode,
		"Flash":          middleware.PopFlash(c),
		"IsAdmin":        true,
		"IsAdminRoute":   true,
	})
}

func (articles *AdminArticles) Create(c *gin.Context) {
	articles.renderForm(c, http.StatusOK, "Create article", "/admin/articles/create", models.Article{}, "")
}

func (articles *AdminArticles) Post(c *gin.Context) {
	article, err := articles.articleFromRequest(c)
	if err != nil {
		articles.renderForm(c, http.StatusBadRequest, "Create article", "/admin/articles/create", article, err.Error())
		return
	}

	var response struct {
		Article models.Article `json:"article"`
	}
	if err := articles.API.Post(c, "/admin/articles/create", articlePayload(article), &response); err != nil {
		articles.renderForm(c, http.StatusBadRequest, "Create article", "/admin/articles/create", article, "Could not create the article.")
		return
	}
	if err := articles.API.UploadArticleImage(c, response.Article.ID); err != nil {
		articles.renderForm(c, http.StatusBadRequest, "Create article", "/admin/articles/create", article, "Could not save the article image.")
		return
	}

	_ = middleware.SetFlash(c, "Article created.")
	c.Redirect(http.StatusFound, "/admin/articles")
}

func (articles *AdminArticles) Edit(c *gin.Context) {
	article, ok := articles.find(c)
	if !ok {
		return
	}

	articles.renderForm(c, http.StatusOK, "Edit article", "/admin/articles/"+c.Param("id")+"/edit", article, "")
}

func (articles *AdminArticles) Put(c *gin.Context) {
	article, err := articles.articleFromRequest(c)
	if err != nil {
		articles.renderForm(c, http.StatusBadRequest, "Edit article", "/admin/articles/"+c.Param("id")+"/edit", article, err.Error())
		return
	}

	id, ok := apiPathID(c.Param("id"))
	if !ok {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}
	if err := articles.API.Put(c, "/admin/articles/"+id+"/edit", articlePayload(article), nil); err != nil {
		articles.renderForm(c, http.StatusBadRequest, "Edit article", "/admin/articles/"+c.Param("id")+"/edit", article, "Could not update the article.")
		return
	}
	if parsedID, err := strconv.Atoi(id); err == nil {
		if err := articles.API.UploadArticleImage(c, uint(parsedID)); err != nil {
			articles.renderForm(c, http.StatusBadRequest, "Edit article", "/admin/articles/"+c.Param("id")+"/edit", article, "Could not save the article image.")
			return
		}
	}

	_ = middleware.SetFlash(c, "Article updated.")
	c.Redirect(http.StatusFound, "/admin/articles")
}

func (articles *AdminArticles) Delete(c *gin.Context) {
	id, ok := apiPathID(c.Param("id"))
	if !ok {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}

	if err := articles.API.Delete(c, "/admin/articles/"+id+"/delete", nil); err != nil {
		_ = middleware.SetFlash(c, "Could not delete the article.")
		c.Redirect(http.StatusFound, "/admin/articles")
		return
	}

	_ = middleware.SetFlash(c, "Article deleted.")
	c.Redirect(http.StatusFound, "/admin/articles")
}

func (articles *AdminArticles) renderForm(c *gin.Context, status int, title string, action string, article models.Article, message string) {
	var response struct {
		Categories []models.ArticleCategory `json:"categories"`
		Products   []models.Product         `json:"products"`
	}
	_ = articles.API.Get(c, "/admin/articles/options", &response)

	var productID uint
	if article.ProductID != nil {
		productID = *article.ProductID
	}

	c.HTML(status, "admin_article_form", gin.H{
		"Title":        title,
		"Action":       action,
		"Article":      article,
		"MainImageURL": article.ImageURL,
		"ProductID":    productID,
		"Categories":   response.Categories,
		"Products":     response.Products,
		"Error":        message,
		"IsAdmin":      true,
		"IsAdminRoute": true,
	})
}

func (articles *AdminArticles) articleFromRequest(c *gin.Context) (models.Article, error) {
	config := form_validator.Config{MaxMemory: 32 << 20, Fields: articles.formFields()}
	if ok := form_validator.ValidateMultiPartForm(c.Request, &config); !ok {
		return models.Article{}, formError("Check the required article fields.")
	}

	categoryID, err := parseRequiredUint(c.PostForm("article_category_id"), "Select an article category.")
	if err != nil {
		return models.Article{}, err
	}

	productID, err := parseOptionalUint(c.PostForm("product_id"))
	if err != nil {
		return models.Article{}, err
	}

	title := strings.TrimSpace(c.PostForm("title"))
	if title == "" {
		return models.Article{}, formError("Title is required.")
	}

	body := strings.TrimSpace(c.PostForm("body"))
	if body == "" {
		return models.Article{}, formError("Body is required.")
	}

	article := models.Article{
		Author:            strings.TrimSpace(c.PostForm("author")),
		Title:             title,
		Slug:              slugify(title),
		Subtitle:          strings.TrimSpace(c.PostForm("subtitle")),
		Body:              body,
		ImageURL:          strings.TrimSpace(c.PostForm("image_url")),
		MetaTitle:         strings.TrimSpace(c.PostForm("meta_title")),
		MetaDescription:   strings.TrimSpace(c.PostForm("meta_description")),
		MetaKeywords:      strings.TrimSpace(c.PostForm("meta_keywords")),
		CanonicalURL:      strings.TrimSpace(c.PostForm("canonical_url")),
		IsPublished:       c.PostForm("is_published") == "on",
		ArticleCategoryID: categoryID,
		ProductID:         productID,
	}

	if article.Author == "" {
		return article, formError("Author is required.")
	}

	if article.IsPublished {
		now := time.Now()
		article.PublishedAt = &now
	}

	return article, nil
}

func (articles *AdminArticles) find(c *gin.Context) (models.Article, bool) {
	id, ok := apiPathID(c.Param("id"))
	if !ok {
		c.AbortWithStatus(http.StatusNotFound)
		return models.Article{}, false
	}

	var response struct {
		Article models.Article `json:"article"`
	}
	if err := articles.API.Get(c, "/admin/articles/"+id, &response); err != nil {
		c.AbortWithStatus(http.StatusNotFound)
		return models.Article{}, false
	}

	return response.Article, true
}

func articlePayload(article models.Article) map[string]interface{} {
	return map[string]interface{}{
		"author": article.Author, "title": article.Title, "slug": article.Slug,
		"subtitle": article.Subtitle, "body": article.Body, "image_url": article.ImageURL,
		"meta_title": article.MetaTitle, "meta_description": article.MetaDescription,
		"meta_keywords": article.MetaKeywords, "canonical_url": article.CanonicalURL,
		"is_published": article.IsPublished, "published_at": article.PublishedAt,
		"article_category_id": article.ArticleCategoryID, "product_id": uintPtrPayload(article.ProductID),
	}
}

func (articles *AdminArticles) formFields() []form_validator.Field {
	if articles.FormFields != nil {
		return articles.FormFields
	}
	return []form_validator.Field{
		{Name: "article_category_id", Validate: true, Type: "uint"},
		{Name: "product_id", Validate: false, Type: "string"},
		{Name: "author", Validate: true, Type: "string"},
		{Name: "title", Validate: true, Type: "string"},
		{Name: "body", Validate: true, Type: "string"},
	}
}

func parseRequiredUint(value string, message string) (uint, error) {
	id, err := parseOptionalUint(value)
	if err != nil || id == nil {
		return 0, formError(message)
	}

	return *id, nil
}

func parseOptionalUint(value string) (*uint, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil, nil
	}

	parsed, err := strconv.Atoi(value)
	if err != nil || parsed <= 0 {
		return nil, formError("Invalid selected value.")
	}

	id := uint(parsed)
	return &id, nil
}

func slugify(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	var builder strings.Builder
	lastWasDash := false

	for _, char := range value {
		if (char >= 'a' && char <= 'z') || (char >= '0' && char <= '9') {
			builder.WriteRune(char)
			lastWasDash = false
			continue
		}

		if !lastWasDash {
			builder.WriteRune('-')
			lastWasDash = true
		}
	}

	return strings.Trim(builder.String(), "-")
}

type formError string

func (e formError) Error() string {
	return string(e)
}
