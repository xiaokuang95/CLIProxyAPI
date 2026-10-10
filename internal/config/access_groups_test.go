package config

import (
	"testing"

	"gopkg.in/yaml.v3"
)

func TestParseAccessGroups(t *testing.T) {
	const raw = `
access:
  api-keys: ["open-key"]
  groups:
    claude-only:
      api-keys: ["sk-claude", " sk-claude "]
      providers: ["Claude", "gemini"]
    empty-keys:
      api-keys: [" "]
      providers: ["codex"]
`
	cfg, err := ParseConfigBytes([]byte(raw))
	if err != nil {
		t.Fatal(err)
	}
	cfg.SanitizeAccessGroups()
	if got := cfg.AllowedProvidersForAPIKey("open-key"); got != nil {
		t.Fatalf("unrestricted key got %v", got)
	}
	got := cfg.AllowedProvidersForAPIKey("sk-claude")
	if len(got) != 2 || got[0] != "claude" || got[1] != "gemini" {
		t.Fatalf("sk-claude providers = %v", got)
	}
	if _, ok := cfg.AccessGroups["empty-keys"]; ok {
		t.Fatal("group with no keys should be dropped")
	}
	keys := cfg.ClientAPIKeys()
	if !stringSliceContains(keys, "open-key") || !stringSliceContains(keys, "sk-claude") {
		t.Fatalf("ClientAPIKeys() = %v", keys)
	}
}

func TestProviderAllowedOpenAICompatAliases(t *testing.T) {
	allowed := map[string]struct{}{"openrouter": {}}
	if !ProviderAllowed("openai-compatible-openrouter", allowed) {
		t.Fatal("compat key should match bare name")
	}
	if !ProviderAllowed("openrouter", map[string]struct{}{"openai-compatible-openrouter": {}}) {
		t.Fatal("bare name should match compat key")
	}
	if ProviderAllowed("claude", allowed) {
		t.Fatal("unrelated provider must be rejected")
	}
	if !ProviderAllowed("claude", nil) {
		t.Fatal("empty allowlist is unrestricted")
	}
}

func TestAccessGroupsV8Validate(t *testing.T) {
	data := []byte(`config-version: 8
server:
  port: 8317
access:
  api-keys: ["open"]
  groups:
    claude-only:
      api-keys: ["sk-claude"]
      providers: ["claude"]
`)
	if err := ValidateV8Config(data); err != nil {
		t.Fatal(err)
	}
}

func TestAccessGroupsRoundTripYAML(t *testing.T) {
	cfg := &SDKConfig{
		APIKeys: []string{"open"},
		AccessGroups: map[string]AccessGroup{
			"team": {APIKeys: []string{"sk-1"}, Providers: []string{"claude"}},
		},
	}
	data, err := yaml.Marshal(cfg)
	if err != nil {
		t.Fatal(err)
	}
	var decoded SDKConfig
	if err = yaml.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded.AccessGroups["team"].APIKeys[0] != "sk-1" {
		t.Fatalf("round trip groups = %+v", decoded.AccessGroups)
	}
}
