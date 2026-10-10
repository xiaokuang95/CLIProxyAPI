package config

import "strings"

const openAICompatibleProviderPrefix = "openai-compatible-"

// AccessGroup binds client API keys to a set of upstream providers.
type AccessGroup struct {
	APIKeys   []string `yaml:"api-keys,omitempty" json:"api-keys,omitempty"`
	Providers []string `yaml:"providers,omitempty" json:"providers,omitempty"`
}

// SanitizeAccessGroups trims group names, keys, and providers, dropping empty entries.
func (cfg *SDKConfig) SanitizeAccessGroups() {
	if cfg == nil || len(cfg.AccessGroups) == 0 {
		cfg.AccessGroups = nil
		return
	}
	out := make(map[string]AccessGroup, len(cfg.AccessGroups))
	for rawName, group := range cfg.AccessGroups {
		name := strings.TrimSpace(rawName)
		if name == "" {
			continue
		}
		keys := uniqueTrimmedStrings(group.APIKeys)
		if len(keys) == 0 {
			continue
		}
		out[name] = AccessGroup{
			APIKeys:   keys,
			Providers: uniqueTrimmedLower(group.Providers),
		}
	}
	if len(out) == 0 {
		cfg.AccessGroups = nil
		return
	}
	cfg.AccessGroups = out
}

// ClientAPIKeys returns access.api-keys plus keys declared in access.groups.
func (cfg *SDKConfig) ClientAPIKeys() []string {
	if cfg == nil {
		return nil
	}
	n := len(cfg.APIKeys)
	for _, group := range cfg.AccessGroups {
		n += len(group.APIKeys)
	}
	if n == 0 {
		return nil
	}
	keys := make([]string, 0, n)
	keys = append(keys, cfg.APIKeys...)
	for _, group := range cfg.AccessGroups {
		keys = append(keys, group.APIKeys...)
	}
	return keys
}

// AllowedProvidersForAPIKey returns the union of providers assigned to key.
// An empty result means the key is unrestricted.
func (cfg *SDKConfig) AllowedProvidersForAPIKey(key string) []string {
	key = strings.TrimSpace(key)
	if cfg == nil || key == "" || len(cfg.AccessGroups) == 0 {
		return nil
	}
	seen := make(map[string]struct{})
	var out []string
	for _, group := range cfg.AccessGroups {
		if !stringSliceContains(group.APIKeys, key) {
			continue
		}
		for _, provider := range group.Providers {
			if _, ok := seen[provider]; ok {
				continue
			}
			seen[provider] = struct{}{}
			out = append(out, provider)
		}
	}
	return out
}

// AllowedProviderSetForAPIKey is AllowedProvidersForAPIKey as a set.
// A nil/empty set means unrestricted.
func (cfg *SDKConfig) AllowedProviderSetForAPIKey(key string) map[string]struct{} {
	providers := cfg.AllowedProvidersForAPIKey(key)
	if len(providers) == 0 {
		return nil
	}
	set := make(map[string]struct{}, len(providers))
	for _, provider := range providers {
		set[provider] = struct{}{}
	}
	return set
}

// ProviderAllowed reports whether provider is in allowed.
// Empty allowed means every provider is allowed.
func ProviderAllowed(provider string, allowed map[string]struct{}) bool {
	if len(allowed) == 0 {
		return true
	}
	name := strings.ToLower(strings.TrimSpace(provider))
	if name == "" {
		return false
	}
	if _, ok := allowed[name]; ok {
		return true
	}
	if strings.HasPrefix(name, openAICompatibleProviderPrefix) {
		if _, ok := allowed[strings.TrimPrefix(name, openAICompatibleProviderPrefix)]; ok {
			return true
		}
	} else if name != "openai-compatibility" {
		if _, ok := allowed[openAICompatibleProviderPrefix+name]; ok {
			return true
		}
	}
	return false
}

func uniqueTrimmedStrings(values []string) []string {
	if len(values) == 0 {
		return nil
	}
	out := make([]string, 0, len(values))
	seen := make(map[string]struct{}, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		out = append(out, value)
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func uniqueTrimmedLower(values []string) []string {
	if len(values) == 0 {
		return nil
	}
	out := make([]string, 0, len(values))
	seen := make(map[string]struct{}, len(values))
	for _, value := range values {
		value = strings.ToLower(strings.TrimSpace(value))
		if value == "" {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		out = append(out, value)
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func stringSliceContains(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}
