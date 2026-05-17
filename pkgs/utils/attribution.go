package utils

import (
	"net/url"
	"sort"
	"strings"
)

var (
	adClickParameterNames = []string{
		"fbclid",
		"gclid",
		"msclkid",
		"ttclid",
		"li_fat_id",
		"twclid",
		"epik",
		"rdt_cid",
	}
	utmParameterNames = []string{
		"utm_source",
		"utm_medium",
		"utm_campaign",
		"utm_content",
		"utm_term",
	}
	affiliateParameterNames = []string{
		"affid",
		"aid",
		"partner_id",
		"partner",
		"subid",
		"sub_id",
		"clickid",
		"click_ref",
		"campaign_id",
		"ref",
		"source",
		"cid",
	}
	travelParameterNames = []string{
		"market",
		"locale",
		"currency",
		"destination",
		"hotel_id",
		"offer_id",
		"checkin",
		"checkout",
		"adults",
		"children",
	}
)

type AttributionPayload struct {
	UTM       map[string]string `json:"utm"`
	AdClick   map[string]string `json:"ad_click"`
	Affiliate map[string]string `json:"affiliate"`
	Travel    map[string]string `json:"travel"`
	All       map[string]string `json:"all"`
}

func ParseAttribution(values url.Values) AttributionPayload {
	payload := AttributionPayload{
		UTM:       collectAttributionValues(values, utmParameterNames),
		AdClick:   collectAttributionValues(values, adClickParameterNames),
		Affiliate: collectAttributionValues(values, affiliateParameterNames),
		Travel:    collectAttributionValues(values, travelParameterNames),
		All:       map[string]string{},
	}
	for _, group := range []map[string]string{payload.UTM, payload.AdClick, payload.Affiliate, payload.Travel} {
		for key, value := range group {
			payload.All[key] = value
		}
	}
	return payload
}

func (payload AttributionPayload) HasData() bool {
	return len(payload.All) > 0
}

func (payload AttributionPayload) Source() string {
	return firstAttributionValue(payload.UTM, "utm_source", payload.Affiliate, "source", payload.Affiliate, "partner", payload.Affiliate, "partner_id", payload.Affiliate, "ref")
}

func (payload AttributionPayload) Medium() string {
	return firstAttributionValue(payload.UTM, "utm_medium")
}

func (payload AttributionPayload) Campaign() string {
	return firstAttributionValue(payload.UTM, "utm_campaign", payload.Affiliate, "campaign_id", payload.Affiliate, "cid")
}

func (payload AttributionPayload) ClickID() string {
	for _, name := range []string{"gclid", "fbclid", "msclkid", "ttclid", "li_fat_id", "twclid", "epik", "rdt_cid", "clickid", "click_ref", "subid", "sub_id"} {
		if value := payload.All[name]; value != "" {
			return value
		}
	}
	return ""
}

func (payload AttributionPayload) Partner() string {
	return firstAttributionValue(payload.Affiliate, "partner", payload.Affiliate, "partner_id", payload.Affiliate, "affid", payload.Affiliate, "aid")
}

func (payload AttributionPayload) Market() string {
	return firstAttributionValue(payload.Travel, "market", payload.Travel, "locale", payload.Travel, "currency")
}

func AttributionKeys(payload map[string]string) []string {
	keys := make([]string, 0, len(payload))
	for key := range payload {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func collectAttributionValues(values url.Values, names []string) map[string]string {
	collected := map[string]string{}
	for _, name := range names {
		if value := strings.TrimSpace(values.Get(name)); value != "" {
			collected[name] = value
		}
	}
	return collected
}

func firstAttributionValue(groupsAndKeys ...interface{}) string {
	for index := 0; index+1 < len(groupsAndKeys); index += 2 {
		group, ok := groupsAndKeys[index].(map[string]string)
		if !ok {
			continue
		}
		key, ok := groupsAndKeys[index+1].(string)
		if !ok {
			continue
		}
		if value := strings.TrimSpace(group[key]); value != "" {
			return value
		}
	}
	return ""
}
