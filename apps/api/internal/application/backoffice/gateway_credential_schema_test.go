package backoffice

import (
	"strings"
	"testing"
)

func TestGatewayCredentialSchemaSupportsAPIKeyWithoutMerchantOrUUIDConstraint(t *testing.T) {
	provider := GatewayProvider{ID: "api-key-fixture", Fields: []GatewayField{{Key: "api_key", Kind: "secret", MaxLength: 4096}}}
	key := "Long-Case-Sensitive-" + strings.Repeat("AbC+9/=", 40)
	updates, fields := validateGatewayCredentials(provider, GatewaySettingsInput{Credentials: map[string]string{"api_key": key}})
	if len(fields) != 0 || updates["api_key"] != key {
		t.Fatal("generic API key was treated as a merchant UUID")
	}
	_, fields = validateGatewayCredentials(provider, GatewaySettingsInput{Credentials: map[string]string{"merchant_id": "01234567-89ab-4cde-8fab-0123456789ab"}})
	if len(fields) == 0 {
		t.Fatal("unrelated credential accepted")
	}
}
