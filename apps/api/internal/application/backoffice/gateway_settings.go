package backoffice

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/cherriva/marketplace/apps/api/internal/domain/commerce"
	"github.com/cherriva/marketplace/apps/api/internal/domain/identity"
)

var ErrGatewayKeyUnavailable = errors.New("gateway_key_unavailable")

type GatewayFieldError struct{ Fields map[string]string }

func (e *GatewayFieldError) Error() string { return "invalid_gateway_settings" }

type GatewayOption struct {
	Value string `json:"value"`
	Label string `json:"label"`
}
type GatewayField struct {
	Key                   string          `json:"key"`
	Label                 string          `json:"label"`
	Kind                  string          `json:"kind"`
	Default               string          `json:"default"`
	RequiredForActivation bool            `json:"required_for_activation"`
	ReadOnly              bool            `json:"read_only"`
	Helper                string          `json:"helper"`
	Options               []GatewayOption `json:"options,omitempty"`
}
type GatewayProvider struct {
	ID             string         `json:"id"`
	Name           string         `json:"name"`
	ExecutionReady bool           `json:"execution_ready"`
	Fields         []GatewayField `json:"fields"`
}

func GatewayProviders() []GatewayProvider {
	return []GatewayProvider{{ID: "zarinpal", Name: "زرین‌پال", ExecutionReady: false, Fields: []GatewayField{
		{Key: "merchant_id", Label: "شناسه پذیرنده (Merchant ID)", Kind: "secret", RequiredForActivation: true, Helper: "شناسه ۳۶ نویسه‌ای پذیرنده؛ برای محیط آزمایشی UUID ساختگی. مقدار ذخیره‌شده دوباره نمایش داده نمی‌شود."},
		{Key: "environment", Label: "محیط درگاه", Kind: "select", Default: "sandbox", Options: []GatewayOption{{"sandbox", "آزمایشی (Sandbox)"}, {"production", "واقعی"}}},
		{Key: "base_url", Label: "آدرس پایه سرویس", Kind: "url", Default: "https://sandbox.zarinpal.com/pg/v4/payment", Helper: "آدرس رسمی متناسب با محیط؛ مقصد دلخواه پذیرفته نمی‌شود."},
		{Key: "currency", Label: "واحد مبلغ ارسالی", Kind: "text", Default: "IRR", ReadOnly: true, Helper: "ریال؛ نمایش و ورود مبلغ در سایت و پنل همچنان تومان است."},
		{Key: "description", Label: "شرح تراکنش", Kind: "text", Default: "شارژ کیف پول چریوا", Helper: "حداکثر ۵۰۰ نویسه."},
		{Key: "auto_verify", Label: "اعتبارسنجی نزد بانک", Kind: "select", Default: "false", Options: []GatewayOption{{"false", "با درخواست سرور چریوا"}, {"true", "خودکار توسط زرین‌پال"}}, Helper: "این گزینه جای تأیید نتیجه در سرور چریوا را نمی‌گیرد."},
	}}}
}

type GatewaySettingsInput struct {
	ID                  string            `json:"id"`
	Version             int               `json:"version"`
	Provider            string            `json:"provider"`
	Name                string            `json:"name"`
	Active              bool              `json:"active"`
	MinToman            string            `json:"min_toman"`
	MaxToman            string            `json:"max_toman"`
	AllCategories       bool              `json:"all_categories"`
	AllowedCategories   []string          `json:"allowed_categories"`
	ExcludedCategories  []string          `json:"excluded_categories"`
	Instructions        string            `json:"instructions"`
	MaxAutomaticRetries int               `json:"max_automatic_retries"`
	AllowWalletTopup    bool              `json:"allow_wallet_topup"`
	Values              map[string]string `json:"values"`
	Credentials         map[string]string `json:"credentials,omitempty"`
	ClearCredentials    []string          `json:"clear_credentials,omitempty"`
}
type GatewaySettingsView struct {
	GatewaySettingsInput
	ConfiguredCredentials []string `json:"configured_credentials"`
	ExecutionReady        bool     `json:"execution_ready"`
}
type GatewaySettingsDirectory struct {
	Providers           []GatewayProvider     `json:"providers"`
	Items               []GatewaySettingsView `json:"items"`
	CanWrite            bool                  `json:"can_write"`
	CredentialsWritable bool                  `json:"credentials_writable"`
}

// ConfigureGatewaySettingsKey must run during construction, before serving requests.
// The same durable, external 32-byte key must be used by all API processes.
func (s *Service) ConfigureGatewaySettingsKey(encoded string) error {
	if encoded == "" {
		return nil
	}
	key, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil || len(key) != 32 {
		return ErrGatewayKeyUnavailable
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return ErrGatewayKeyUnavailable
	}
	s.gatewayCipher, err = cipher.NewGCM(block)
	return err
}
func gatewaySettingsView(g commerce.GatewaySettings) GatewaySettingsView {
	maxValue := ""
	if g.Max > 0 {
		maxValue = commerce.Toman(g.Max)
	}
	return GatewaySettingsView{GatewaySettingsInput: GatewaySettingsInput{ID: g.ID, Version: g.Version, Provider: g.Provider, Name: g.Name, MinToman: commerce.Toman(g.Min), MaxToman: maxValue, AllCategories: g.AllCategories, AllowedCategories: append([]string{}, g.AllowedCategories...), ExcludedCategories: append([]string{}, g.ExcludedCategories...), Instructions: g.Instructions, MaxAutomaticRetries: g.MaxAutomaticRetries, AllowWalletTopup: g.AllowWalletTopup, Values: g.Values}, ConfiguredCredentials: append([]string{}, g.CredentialNames...)}
}
func (s *Service) ReadGatewaySettings(ctx context.Context, a identity.User) (GatewaySettingsDirectory, error) {
	w, err := s.repository.Snapshot(ctx)
	if err != nil {
		return GatewaySettingsDirectory{}, err
	}
	if err = requireCommerceAdmin(w, a, CommerceSettingsRead); err != nil {
		return GatewaySettingsDirectory{}, err
	}
	canWrite := requireCommerceAdmin(w, a, CommercePaymentWrite) == nil
	out := GatewaySettingsDirectory{Providers: GatewayProviders(), Items: []GatewaySettingsView{}, CanWrite: canWrite, CredentialsWritable: canWrite && s.gatewayCipher != nil}
	for _, g := range commerceState(&w).GatewaySettings {
		out.Items = append(out.Items, gatewaySettingsView(g))
	}
	return out, nil
}
func gatewayAAD(id, provider string) []byte {
	return []byte("cherriva:gateway-settings:v1:" + id + ":" + provider)
}
func (s *Service) readGatewayCredentials(g commerce.GatewaySettings) (map[string]string, error) {
	values := map[string]string{}
	if g.EncryptedCredentials == "" {
		return values, nil
	}
	if s.gatewayCipher == nil {
		return nil, ErrGatewayKeyUnavailable
	}
	raw, err := base64.StdEncoding.DecodeString(g.EncryptedCredentials)
	if err != nil || len(raw) < s.gatewayCipher.NonceSize() {
		return nil, ErrGatewayKeyUnavailable
	}
	n := s.gatewayCipher.NonceSize()
	plain, err := s.gatewayCipher.Open(nil, raw[:n], raw[n:], gatewayAAD(g.ID, g.Provider))
	if err != nil || json.Unmarshal(plain, &values) != nil {
		return nil, ErrGatewayKeyUnavailable
	}
	return values, nil
}
func (s *Service) sealGatewayCredentials(g *commerce.GatewaySettings, values map[string]string) error {
	g.CredentialNames = []string{}
	for k := range values {
		g.CredentialNames = append(g.CredentialNames, k)
	}
	sort.Strings(g.CredentialNames)
	if len(values) == 0 {
		g.EncryptedCredentials = ""
		return nil
	}
	if s.gatewayCipher == nil {
		return ErrGatewayKeyUnavailable
	}
	nonce := make([]byte, s.gatewayCipher.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return err
	}
	plain, err := json.Marshal(values)
	if err != nil {
		return err
	}
	g.EncryptedCredentials = base64.StdEncoding.EncodeToString(s.gatewayCipher.Seal(nonce, nonce, plain, gatewayAAD(g.ID, g.Provider)))
	return nil
}

var merchantUUID = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)

func validateGatewayValues(in GatewaySettingsInput) (map[string]string, map[string]string) {
	fields := map[string]string{}
	values := map[string]string{}
	for _, f := range GatewayProviders()[0].Fields {
		if f.Kind != "secret" {
			values[f.Key] = f.Default
		}
	}
	for k, v := range in.Values {
		if _, ok := values[k]; !ok {
			fields["values."+k] = "این فیلد برای درگاه انتخاب‌شده تعریف نشده است."
			continue
		}
		values[k] = strings.TrimSpace(v)
	}
	base := "https://sandbox.zarinpal.com/pg/v4/payment"
	if values["environment"] == "production" {
		base = "https://payment.zarinpal.com/pg/v4/payment"
	} else if values["environment"] != "sandbox" {
		fields["values.environment"] = "محیط درگاه معتبر نیست."
	}
	if strings.TrimRight(values["base_url"], "/") != base {
		fields["values.base_url"] = "آدرس رسمی درگاه باید با محیط انتخاب‌شده یکسان باشد."
	}
	values["base_url"] = base
	if values["currency"] != "IRR" {
		fields["values.currency"] = "واحد ارسال در این نسخه ریال است."
	}
	if !textBound(values["description"], 500) || values["description"] == "" {
		fields["values.description"] = "شرح تراکنش را با حداکثر ۵۰۰ نویسه وارد کنید."
	}
	if values["auto_verify"] != "true" && values["auto_verify"] != "false" {
		fields["values.auto_verify"] = "روش اعتبارسنجی معتبر نیست."
	}
	return values, fields
}

func (s *Service) SaveGatewaySettings(ctx context.Context, a identity.User, in GatewaySettingsInput, key string) (json.RawMessage, error) {
	return s.mutateCommerce(ctx, "gateway-settings:"+a.ID, key, in, func(w Workspace) error { return requireCommerceAdmin(w, a, CommercePaymentWrite) }, func(w *Workspace, id string, now time.Time) (any, error) {
		c := commerceState(w)
		index := -1
		g := commerce.GatewaySettings{ID: id, Provider: in.Provider}
		if in.ID != "" {
			for i, saved := range c.GatewaySettings {
				if saved.ID == in.ID {
					index = i
					g = saved
					break
				}
			}
			if index < 0 {
				return nil, ErrNotFound
			}
			if g.Version != in.Version {
				return nil, ErrConflict
			}
			if g.Provider != in.Provider {
				return nil, ErrInvalid
			}
		} else if in.Version != 0 || len(c.GatewaySettings) >= 20 {
			return nil, ErrInvalid
		}
		if in.Provider != "zarinpal" {
			return nil, &GatewayFieldError{map[string]string{"provider": "ارائه‌دهنده درگاه پشتیبانی نمی‌شود."}}
		}
		values, fields := validateGatewayValues(in)
		if in.Active {
			fields["active"] = "اتصال پرداخت هنوز پیاده‌سازی نشده است؛ فقط ذخیره تنظیمات مجاز است."
		}
		if in.MaxAutomaticRetries < 0 || in.MaxAutomaticRetries > 10 {
			fields["max_automatic_retries"] = "تعداد تلاش مجدد باید از صفر تا ۱۰ باشد."
		}
		if strings.TrimSpace(in.Name) == "" || !textBound(in.Name, 80) {
			fields["name"] = "نام درگاه را با حداکثر ۸۰ نویسه وارد کنید."
		}
		if !textBound(in.Instructions, 1000) {
			fields["instructions"] = "راهنما باید حداکثر ۱۰۰۰ نویسه باشد."
		}
		if !categoryIDsValid(*w, in.AllowedCategories) || !categoryIDsValid(*w, in.ExcludedCategories) {
			fields["categories"] = "دسته‌های انتخاب‌شده معتبر نیستند."
		}
		minValue, err := commerce.ParseToman(in.MinToman)
		if err != nil {
			fields["min_toman"] = "حداقل مبلغ معتبر نیست."
		}
		var maxValue commerce.Amount
		if in.MaxToman != "" {
			maxValue, err = commerce.ParseToman(in.MaxToman)
			if err != nil || maxValue <= 0 || maxValue < minValue {
				fields["max_toman"] = "حداکثر باید مثبت و حداقل برابر کف مبلغ باشد."
			}
		}
		for k, v := range in.Credentials {
			if k != "merchant_id" {
				fields["credentials"] = "اطلاعات اتصال نامرتبط با درگاه ارسال شده است."
			} else if strings.TrimSpace(v) != "" && !merchantUUID.MatchString(strings.TrimSpace(v)) {
				fields["credentials.merchant_id"] = "شناسه پذیرنده باید UUID با ۳۶ نویسه باشد."
			}
		}
		for _, k := range in.ClearCredentials {
			if k != "merchant_id" || strings.TrimSpace(in.Credentials[k]) != "" {
				fields["credentials"] = "حذف و جایگزینی هم‌زمان اطلاعات اتصال معتبر نیست."
			}
		}
		if len(fields) > 0 {
			return nil, &GatewayFieldError{fields}
		}
		credentials, err := s.readGatewayCredentials(g)
		if err != nil {
			return nil, err
		}
		for _, k := range in.ClearCredentials {
			delete(credentials, k)
		}
		for k, v := range in.Credentials {
			if v = strings.TrimSpace(v); v != "" {
				credentials[k] = strings.ToLower(v)
			}
		}
		if err = s.sealGatewayCredentials(&g, credentials); err != nil {
			return nil, err
		}
		g.Version++
		g.Name = strings.TrimSpace(in.Name)
		g.Min = minValue
		g.Max = maxValue
		g.AllCategories = in.AllCategories
		g.AllowedCategories = append([]string{}, in.AllowedCategories...)
		g.ExcludedCategories = append([]string{}, in.ExcludedCategories...)
		g.Instructions = strings.TrimSpace(in.Instructions)
		g.MaxAutomaticRetries = in.MaxAutomaticRetries
		g.AllowWalletTopup = in.AllowWalletTopup
		g.Values = values
		if index < 0 {
			c.GatewaySettings = append(c.GatewaySettings, g)
		} else {
			c.GatewaySettings[index] = g
		}
		commerceAudit(w, id, a, "gateway.settings.save", g.ID, "", now)
		return gatewaySettingsView(g), nil
	})
}
