package store

import "github.com/cherriva/marketplace/apps/api/internal/domain/shipping"

type Profile struct {
	Description string `json:"description,omitempty"`
	Email       string `json:"email,omitempty"`
	Province    string `json:"province,omitempty"`
	Address     string `json:"address,omitempty"`
	PostalCode  string `json:"postal_code,omitempty"`
	Phone       string `json:"phone,omitempty"`
	Mobile      string `json:"mobile,omitempty"`
	ImageID     string `json:"image_id,omitempty"`
}

// Nil preserves the permissions of legacy stores until a policy is explicitly saved.
type SalePermission struct {
	MethodID    string `json:"method_id"`
	DisplayName string `json:"display_name,omitempty"`
}
type Eligibility struct {
	Version           int              `json:"version"`
	CategoryMode      string           `json:"category_mode"`
	AllowedCategories []string         `json:"allowed_categories"`
	DeniedCategories  []string         `json:"denied_categories"`
	Methods           []SalePermission `json:"methods"`
}

type Store struct {
	Preparation          *shipping.PreparationPolicy `json:"preparation,omitempty"`
	Eligibility          *Eligibility                `json:"eligibility,omitempty"`
	Version              int                         `json:"version,omitempty"`
	Profile              *Profile                    `json:"profile,omitempty"`
	ContentPolicy        string                      `json:"content_policy,omitempty"`
	ContentPolicyVersion int                         `json:"content_policy_version"`
	ID                   string                      `json:"id"`
	Name                 string                      `json:"name"`
	Owner                string                      `json:"owner"`
	City                 string                      `json:"city"`
	Active               bool                        `json:"active"`
	Model                string                      `json:"model"`
	LeadDays             int                         `json:"lead_days"`
}
