package configaccess

import (
	"context"
	"net/http"
	"testing"

	sdkaccess "github.com/router-for-me/CLIProxyAPI/v8/sdk/access"
	sdkconfig "github.com/router-for-me/CLIProxyAPI/v8/sdk/config"
)

func TestRegisterIncludesGroupKeys(t *testing.T) {
	t.Cleanup(func() {
		sdkaccess.ClearExclusiveProvider()
		sdkaccess.UnregisterProvider(sdkaccess.AccessProviderTypeConfigAPIKey)
	})
	Register(&sdkconfig.SDKConfig{
		AccessGroups: map[string]sdkconfig.AccessGroup{
			"claude-only": {APIKeys: []string{"sk-group"}, Providers: []string{"claude"}},
		},
	})
	sdkaccess.SetExclusiveProvider(sdkaccess.AccessProviderTypeConfigAPIKey)
	providers := sdkaccess.RegisteredProviders()
	if len(providers) != 1 {
		t.Fatalf("providers = %d, want 1", len(providers))
	}
	result, err := providers[0].Authenticate(context.Background(), httptestRequest("sk-group"))
	if err != nil {
		t.Fatalf("group key rejected: %v", err)
	}
	if result == nil || result.Principal != "sk-group" {
		t.Fatalf("result = %+v", result)
	}
}

func httptestRequest(key string) *http.Request {
	req, _ := http.NewRequest(http.MethodGet, "http://example.invalid/v1/models", nil)
	req.Header.Set("Authorization", "Bearer "+key)
	return req
}
