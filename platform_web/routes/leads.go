package routes

import (
	"encoding/json"
	"leadstorefront/pkgs/middleware"
	"leadstorefront/pkgs/models"
	"leadstorefront/pkgs/utils"
	"net/http"
	"net/url"
	"sort"
	"strings"

	"github.com/gin-gonic/gin"
)

type Leads struct {
	API *APIClient
}

type leadValue struct {
	Name  string
	Value string
}

func (leads *Leads) Get(c *gin.Context) {
	if c.Param("id") != "" {
		leads.Show(c)
		return
	}
	leads.Index(c)
}

func (leads *Leads) Post(c *gin.Context) {
	storefrontID := storefrontIDFromPath(c)
	if storefrontID == "" {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}
	var formResponse struct {
		Fields []models.LeadFormField `json:"fields"`
	}
	if err := leads.API.Get(c, "/storefronts/"+storefrontID+"/lead-form", &formResponse); err != nil || len(formResponse.Fields) == 0 {
		c.String(http.StatusBadRequest, "lead form is not available")
		return
	}

	values := map[string]string{}
	for _, field := range formResponse.Fields {
		values[field.Name] = strings.TrimSpace(c.PostForm("lead_" + field.Name))
	}
	source, tracking := leadTracking(c)
	if err := leads.API.Post(c, "/storefronts/"+storefrontID+"/leads", map[string]interface{}{
		"source":   source,
		"tracking": tracking,
		"values":   values,
	}, nil); err != nil {
		_ = middleware.SetFlash(c, "Could not submit the form.")
		c.Redirect(http.StatusFound, storefrontReturnPath(c, storefrontID)+"#lead-form")
		return
	}

	_ = middleware.SetFlash(c, "Thanks, your details have been sent.")
	c.Redirect(http.StatusFound, storefrontReturnPath(c, storefrontID)+"#lead-form")
}

func (leads *Leads) Put(c *gin.Context) {
	c.AbortWithStatus(http.StatusMethodNotAllowed)
}

func (leads *Leads) Delete(c *gin.Context) {
	c.AbortWithStatus(http.StatusMethodNotAllowed)
}

func (leads *Leads) Index(c *gin.Context) {
	var response struct {
		Leads      []models.Lead    `json:"leads"`
		Pagination utils.Pagination `json:"pagination"`
	}
	page, limit := utils.GetPaginationQuery(c)
	if err := leads.API.Get(c, "/admin/leads?page="+page+"&limit="+limit, &response); err != nil {
		c.String(http.StatusInternalServerError, "could not load leads")
		return
	}

	c.HTML(http.StatusOK, "admin_leads_index", gin.H{
		"Title":        "Leads",
		"Leads":        response.Leads,
		"Pagination":   response.Pagination,
		"Limit":        limit,
		"Flash":        middleware.PopFlash(c),
		"IsAdmin":      true,
		"IsSuper":      isCurrentSuper(c),
		"IsAdminRoute": true,
	})
}

func (leads *Leads) Show(c *gin.Context) {
	id, ok := apiPathID(c.Param("id"))
	if !ok {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}
	var response struct {
		Lead models.Lead `json:"lead"`
	}
	if err := leads.API.Get(c, "/admin/leads/"+id, &response); err != nil {
		c.String(http.StatusNotFound, "could not load lead")
		return
	}
	c.HTML(http.StatusOK, "admin_lead_show", gin.H{
		"Title":        "Lead",
		"Lead":         response.Lead,
		"Values":       leadValues(response.Lead.ValuesJSON),
		"IsAdmin":      true,
		"IsSuper":      isCurrentSuper(c),
		"IsAdminRoute": true,
	})
}

func leadTracking(c *gin.Context) (string, string) {
	query := c.Request.URL.Query()
	source := firstQueryValue(query, "utm_source", "source", "ref")
	tracking := firstQueryValue(query, "utm_campaign", "utm_medium", "utm_content", "tracking", "pixel")
	if tracking == "" && c.Request.Referer() != "" {
		tracking = c.Request.Referer()
	}
	return source, tracking
}

func firstQueryValue(query url.Values, names ...string) string {
	for _, name := range names {
		if value := strings.TrimSpace(query.Get(name)); value != "" {
			return value
		}
	}
	return ""
}

func storefrontReturnPath(c *gin.Context, storefrontID string) string {
	if country := c.Param("country"); country != "" {
		return "/" + country + "/storefronts/" + storefrontID
	}
	return "/storefronts/" + storefrontID
}

func leadValues(raw string) []leadValue {
	values := map[string]string{}
	_ = json.Unmarshal([]byte(raw), &values)
	result := make([]leadValue, 0, len(values))
	for name, value := range values {
		result = append(result, leadValue{Name: name, Value: value})
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].Name < result[j].Name
	})
	return result
}
