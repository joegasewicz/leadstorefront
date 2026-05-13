package routes

import (
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	form_validator "github.com/joegasewicz/form-validator"
)

const (
	purchaseTierKey       = "purchase_tier"
	purchaseConfirmedKey  = "purchase_confirmed"
	defaultPurchaseTier   = "growth"
	storefrontCreateRoute = "/admin/storefronts/create"
)

type Purchase struct {
	Fields []form_validator.Field
}

type purchaseTier struct {
	ID          string
	Name        string
	Price       string
	Description string
	Featured    bool
	Features    []string
}

func (purchase *Purchase) Get(c *gin.Context) {
	switch c.FullPath() {
	case "/purchase":
		purchase.Choose(c)
	case "/payment":
		purchase.Payment(c)
	default:
		c.AbortWithStatus(http.StatusNotFound)
	}
}

func (purchase *Purchase) Post(c *gin.Context) {
	switch c.FullPath() {
	case "/purchase":
		purchase.ChoosePost(c)
	case "/payment":
		purchase.PaymentPost(c)
	default:
		c.AbortWithStatus(http.StatusMethodNotAllowed)
	}
}

func (purchase *Purchase) Put(c *gin.Context) {
	c.AbortWithStatus(http.StatusMethodNotAllowed)
}

func (purchase *Purchase) Delete(c *gin.Context) {
	c.AbortWithStatus(http.StatusMethodNotAllowed)
}

func (purchase *Purchase) Choose(c *gin.Context) {
	c.HTML(http.StatusOK, "purchase", gin.H{
		"Title":        "Choose a Tier",
		"Tiers":        purchaseTiers(),
		"SelectedTier": selectedTier(c),
		"IsAdmin":      true,
		"IsAdminRoute": true,
	})
}

func (purchase *Purchase) ChoosePost(c *gin.Context) {
	config := form_validator.Config{Fields: purchase.fields()}
	if ok := form_validator.ValidateForm(c.Request, &config); !ok {
		purchase.renderChoose(c, http.StatusBadRequest, "Choose a tier to continue.")
		return
	}
	tier, _ := form_validator.GetString("tier", &config)
	if !validPurchaseTier(tier) {
		purchase.renderChoose(c, http.StatusBadRequest, "Choose a valid tier to continue.")
		return
	}

	session := sessions.Default(c)
	session.Set(purchaseTierKey, tier)
	session.Delete(purchaseConfirmedKey)
	if err := session.Save(); err != nil {
		purchase.renderChoose(c, http.StatusInternalServerError, "Could not save your tier choice.")
		return
	}

	c.Redirect(http.StatusFound, "/payment")
}

func (purchase *Purchase) Payment(c *gin.Context) {
	tier, ok := sessionTier(c)
	if !ok {
		c.Redirect(http.StatusFound, "/purchase")
		return
	}

	c.HTML(http.StatusOK, "payment", gin.H{
		"Title":        "Payment",
		"Tier":         tierByID(tier),
		"IsAdmin":      true,
		"IsAdminRoute": true,
	})
}

func (purchase *Purchase) PaymentPost(c *gin.Context) {
	tier, ok := sessionTier(c)
	if !ok {
		c.Redirect(http.StatusFound, "/purchase")
		return
	}
	if c.PostForm("confirm_payment") != "on" {
		c.HTML(http.StatusBadRequest, "payment", gin.H{
			"Title":        "Payment",
			"Tier":         tierByID(tier),
			"Error":        "Confirm that you want to continue with this tier.",
			"IsAdmin":      true,
			"IsAdminRoute": true,
		})
		return
	}

	session := sessions.Default(c)
	session.Set(purchaseConfirmedKey, true)
	_ = session.Save()
	c.Redirect(http.StatusFound, storefrontCreateRoute)
}

func (purchase *Purchase) renderChoose(c *gin.Context, status int, message string) {
	c.HTML(status, "purchase", gin.H{
		"Title":        "Choose a Tier",
		"Tiers":        purchaseTiers(),
		"SelectedTier": selectedTier(c),
		"Error":        message,
		"IsAdmin":      true,
		"IsAdminRoute": true,
	})
}

func (purchase *Purchase) fields() []form_validator.Field {
	if purchase.Fields != nil {
		return purchase.Fields
	}
	return []form_validator.Field{{Name: "tier", Validate: true, Type: "string"}}
}

func selectedTier(c *gin.Context) string {
	if tier, ok := sessionTier(c); ok {
		return tier
	}
	return defaultPurchaseTier
}

func sessionTier(c *gin.Context) (string, bool) {
	value, _ := sessions.Default(c).Get(purchaseTierKey).(string)
	if validPurchaseTier(value) {
		return value, true
	}
	return "", false
}

func validPurchaseTier(tier string) bool {
	for _, candidate := range purchaseTiers() {
		if candidate.ID == tier {
			return true
		}
	}
	return false
}

func tierByID(id string) purchaseTier {
	for _, tier := range purchaseTiers() {
		if tier.ID == id {
			return tier
		}
	}
	return tierByID(defaultPurchaseTier)
}

func purchaseTiers() []purchaseTier {
	return []purchaseTier{
		{
			ID:          "launch",
			Name:        "Launch",
			Price:       "GBP 79/mo",
			Description: "For one focused storefront, affiliate destination, or lead-capture campaign.",
			Features: []string{
				"One hosted storefront",
				"One UK market experience",
				"Product, article, and offer pages",
				"Affiliate link management",
			},
		},
		{
			ID:          "growth",
			Name:        "Growth",
			Price:       "GBP 249/mo",
			Description: "For teams running active storefront campaigns with more content and lead workflows.",
			Featured:    true,
			Features: []string{
				"Up to five hosted storefronts",
				"UK plus two additional market paths",
				"Advanced lead-capture flows",
				"Priority support",
			},
		},
		{
			ID:          "scale",
			Name:        "Scale",
			Price:       "From GBP 799/mo",
			Description: "For storefront networks, publishers, agencies, and advanced operating requirements.",
			Features: []string{
				"Unlimited storefront planning",
				"Multi-market localization strategy",
				"Custom domain and routing support",
				"Implementation support",
			},
		},
	}
}
