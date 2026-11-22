package store

type Store struct {
	ContentPolicy        string `json:"content_policy,omitempty"`
	ContentPolicyVersion int    `json:"content_policy_version"`
	ID                   string `json:"id"`
	Name                 string `json:"name"`
	Owner                string `json:"owner"`
	City                 string `json:"city"`
	Active               bool   `json:"active"`
	Model                string `json:"model"`
	LeadDays             int    `json:"lead_days"`
}
