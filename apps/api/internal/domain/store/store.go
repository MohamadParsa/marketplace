package store

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

type Store struct {
	Version              int      `json:"version,omitempty"`
	Profile              *Profile `json:"profile,omitempty"`
	ContentPolicy        string   `json:"content_policy,omitempty"`
	ContentPolicyVersion int      `json:"content_policy_version"`
	ID                   string   `json:"id"`
	Name                 string   `json:"name"`
	Owner                string   `json:"owner"`
	City                 string   `json:"city"`
	Active               bool     `json:"active"`
	Model                string   `json:"model"`
	LeadDays             int      `json:"lead_days"`
}
