package commerce

// GatewaySettings is private configuration. It is deliberately separate from
// PaymentMethod, which is included in customer responses and order snapshots.
// These drafts cannot initiate a payment until a provider adapter is implemented.
type GatewaySettings struct {
	ID                   string            `json:"id"`
	Version              int               `json:"version"`
	Provider             string            `json:"provider"`
	Name                 string            `json:"name"`
	Min                  Amount            `json:"min"`
	Max                  Amount            `json:"max"`
	AllCategories        bool              `json:"all_categories"`
	AllowedCategories    []string          `json:"allowed_categories"`
	ExcludedCategories   []string          `json:"excluded_categories"`
	Instructions         string            `json:"instructions"`
	MaxAutomaticRetries  int               `json:"max_automatic_retries"`
	AllowWalletTopup     bool              `json:"allow_wallet_topup"`
	Values               map[string]string `json:"values"`
	EncryptedCredentials string            `json:"encrypted_credentials,omitempty"`
	CredentialNames      []string          `json:"credential_names,omitempty"`
}
