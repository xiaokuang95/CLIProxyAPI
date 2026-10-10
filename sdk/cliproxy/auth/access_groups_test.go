package auth

import "testing"

func TestAuthSelectionEligibilityAllowsProviders(t *testing.T) {
	restricted := authSelectionEligibility{allowedProviders: map[string]struct{}{"claude": {}}}
	if restricted.allows(&Auth{Provider: "gemini"}) {
		t.Fatal("gemini should be rejected")
	}
	if !restricted.allows(&Auth{Provider: "claude"}) {
		t.Fatal("claude should be allowed")
	}
	if restricted.allows(nil) {
		t.Fatal("nil auth should be rejected")
	}
	if !(authSelectionEligibility{}).allows(&Auth{Provider: "gemini"}) {
		t.Fatal("unrestricted eligibility should allow any provider")
	}
	compat := authSelectionEligibility{allowedProviders: map[string]struct{}{"openrouter": {}}}
	if !compat.allows(&Auth{Provider: "openai-compatibility", Attributes: map[string]string{"compat_name": "openrouter"}}) {
		t.Fatal("openai-compat name should match")
	}
}
