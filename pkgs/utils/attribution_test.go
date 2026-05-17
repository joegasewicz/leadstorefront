package utils

import (
	"net/url"
	"testing"
)

func TestParseAttributionGroupsSupportedParameters(t *testing.T) {
	values := url.Values{}
	values.Set("utm_source", "google")
	values.Set("utm_medium", "cpc")
	values.Set("utm_campaign", "spring")
	values.Set("gclid", "g-123")
	values.Set("partner", "booking")
	values.Set("clickid", "click-456")
	values.Set("market", "uk")
	values.Set("destination", "london")
	values.Set("ignored", "value")

	payload := ParseAttribution(values)
	if !payload.HasData() {
		t.Fatal("expected attribution payload data")
	}
	if payload.Source() != "google" {
		t.Fatalf("expected source google, got %q", payload.Source())
	}
	if payload.Medium() != "cpc" {
		t.Fatalf("expected medium cpc, got %q", payload.Medium())
	}
	if payload.Campaign() != "spring" {
		t.Fatalf("expected campaign spring, got %q", payload.Campaign())
	}
	if payload.ClickID() != "g-123" {
		t.Fatalf("expected ad click ID to take priority, got %q", payload.ClickID())
	}
	if payload.Partner() != "booking" {
		t.Fatalf("expected partner booking, got %q", payload.Partner())
	}
	if payload.Market() != "uk" {
		t.Fatalf("expected market uk, got %q", payload.Market())
	}
	if _, ok := payload.All["ignored"]; ok {
		t.Fatal("did not expect unsupported parameter in payload")
	}
}
