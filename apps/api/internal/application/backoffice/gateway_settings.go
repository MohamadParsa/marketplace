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
	Section               string          `json:"section"`
	Format                string          `json:"format,omitempty"`
	MaxLength             int             `json:"max_length"`
	Default               string          `json:"default"`
	RequiredForActivation bool            `json:"required_for_activation"`
	ReadOnly              bool            `json:"read_only"`
	Helper                string          `json:"helper"`
	Options               []GatewayOption `json:"options,omitempty"`
}
type GatewaySection struct {
	ID          string `json:"id"`
	Label       string `json:"label"`
	Description string `json:"description"`
	Expanded    bool   `json:"expanded"`
}
type GatewayProvider struct {
	ID             string           `json:"id"`
	Name           string           `json:"name"`
	ExecutionReady bool             `json:"execution_ready"`
	Fields         []GatewayField   `json:"fields"`
	Sections       []GatewaySection `json:"sections"`
}

func GatewayProviders() []GatewayProvider {
	return []GatewayProvider{{ID: "zarinpal", Name: "زرین‌پال", ExecutionReady: false,
		Sections: []GatewaySection{
			{ID: "api", Label: "احراز هویت API زرین‌پال", Description: "راه‌اندازی حساب پذیرنده پیش از فعال‌سازی و توسط ادمین انجام می‌شود، نه در اولین پرداخت مشتری. در مرحله اتصال، دریافت کد و تأیید حساب در همین بخش اضافه می‌شود؛ فعلاً فقط مشخصات ذخیره می‌شوند. کد یک‌بارمصرف در تنظیمات ذخیره نمی‌شود."},
			{ID: "payment", Label: "تنظیمات پرداخت", Description: "تنظیمات پرداخت V4؛ شناسه پذیرنده در این فرم اختیاری است و به‌عنوان credential مشترک همه درگاه‌ها الزام نمی‌شود."},
			{ID: "advanced", Label: "تنظیمات پیشرفته تأیید پرداخت", Description: "پس از بازگشت مشتری به callback، سرور نتیجه را با verify بررسی می‌کند؛ صرف بازگشت به سایت یا Status=OK پرداخت قطعی نیست. احراز هویت API مربوط به دسترسی حساب پذیرنده است."},
		}, Fields: []GatewayField{
			{Key: "client_id", Label: "شناسه کلاینت (Client ID)", Kind: "credential", Section: "api", Format: "positive-integer", MaxLength: 20, Helper: "شناسه عددی کلاینت دریافتی از زرین‌پال؛ برای پیش‌نویس می‌تواند خالی باشد."},
			{Key: "client_secret", Label: "کلید محرمانه کلاینت (Client Secret)", Kind: "secret", Section: "api", MaxLength: 4096, Helper: "کلید دریافتی از زرین‌پال؛ حساس به حروف است و رمزنگاری‌شده ذخیره می‌شود."},
			{Key: "username", Label: "حساب پذیرنده زرین‌پال", Kind: "credential", Section: "api", MaxLength: 254, Helper: "شماره همراه یا ایمیل حساب پذیرنده برای احراز هویت API؛ در پیش‌نویس اختیاری است."},
			{Key: "channel", Label: "روش دریافت کد حساب پذیرنده", Kind: "select", Section: "api", Default: "ussd", Options: []GatewayOption{{"ussd", "USSD"}, {"sms", "پیامک"}}, Helper: "کد متعلق به حساب پذیرنده و مرحله راه‌اندازی ادمین است؛ مشتری کدی برای ورود به حساب پذیرنده وارد نمی‌کند."},
			{Key: "api_base_url", Label: "آدرس پایه احراز هویت و API حساب", Kind: "url", Section: "api", Default: "https://next.zarinpal.com/api", MaxLength: 2048, Helper: "برای oauth/initialize و oauth/token؛ این آدرس سرویس درخواست و تأیید پرداخت نیست."},
			{Key: "merchant_id", Label: "شناسه پذیرنده (Merchant ID) — اختیاری", Kind: "credential", Section: "payment", Format: "uuid", MaxLength: 36, Helper: "اختیاری در تنظیمات؛ فقط برای عملیات پرداخت V4 کاربرد دارد. با Client ID یکسان نیست."},
			{Key: "environment", Label: "محیط پرداخت", Kind: "select", Section: "payment", Default: "sandbox", Options: []GatewayOption{{"sandbox", "آزمایشی (Sandbox)"}, {"production", "واقعی"}}},
			{Key: "base_url", Label: "آدرس پایه سرویس پرداخت", Kind: "url", Section: "payment", Default: "https://sandbox.zarinpal.com/pg/v4/payment", MaxLength: 2048, Helper: "برای request.json و verify.json؛ با آدرس API حساب فرق دارد و باید متناسب با محیط پرداخت باشد."},
			{Key: "currency", Label: "واحد مبلغ ارسالی به درگاه", Kind: "select", Section: "payment", Default: string(commerce.SiteCurrency), Options: []GatewayOption{{string(commerce.CurrencyToman), "تومان (IRT)"}, {string(commerce.CurrencyRial), "ریال (IRR)"}}, Helper: "واحد پیش‌فرض سایت تومان است؛ واحد ارسالی به این درگاه را انتخاب کنید."},
			{Key: "exchange_rate", Label: "ضریب تبدیل مبلغ سایت به مبلغ درگاه", Kind: "decimal", Section: "payment", Default: commerce.SiteCurrency.DefaultGatewayRate(), MaxLength: 14, Helper: "مبلغ سایت به تومان × ضریب تبدیل = مبلغ ارسالی به درگاه؛ تومان به ریال ×۱۰ و تومان به تومان ×۱. تا ۶ رقم اعشار؛ بدون گردکردن پنهان."},
			{Key: "auto_verify", Label: "آغاز تأیید تراکنش نزد بانک", Kind: "select", Section: "advanced", Default: "false", Options: []GatewayOption{{"false", "با verify سرور چریوا پس از callback (پیش‌فرض)"}, {"true", "خودکار توسط زرین‌پال"}}, Helper: "در حالت پیش‌فرض، verify چریوا تأیید نزد بانک را آغاز می‌کند. حالت خودکار فقط آغاز این کار را به زرین‌پال می‌سپارد؛ نتیجه در چریوا همچنان باید سروری بررسی و یک‌بار ثبت شود."},
		}}}
}

func gatewayProviderDefinition(id string) (GatewayProvider, bool) {
	for _, p := range GatewayProviders() {
		if p.ID == id {
			return p, true
		}
	}
	return GatewayProvider{}, false
}
func gatewayCredentialField(f GatewayField) bool { return f.Kind == "secret" || f.Kind == "credential" }

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
	provider, _ := gatewayProviderDefinition(g.Provider)
	return GatewaySettingsView{GatewaySettingsInput: GatewaySettingsInput{ID: g.ID, Version: g.Version, Provider: g.Provider, Name: g.Name, Active: g.Active, MinToman: commerce.Toman(g.Min), MaxToman: maxValue, AllCategories: g.AllCategories, AllowedCategories: append([]string{}, g.AllowedCategories...), ExcludedCategories: append([]string{}, g.ExcludedCategories...), Instructions: g.Instructions, MaxAutomaticRetries: g.MaxAutomaticRetries, AllowWalletTopup: g.AllowWalletTopup, Values: gatewayNonCredentialValues(provider, g.Values)}, ConfiguredCredentials: append([]string{}, g.CredentialNames...)}
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
	for _, p := range out.Providers {
		found := false
		for _, g := range out.Items {
			if g.Provider == p.ID {
				found = true
				break
			}
		}
		if !found {
			out.Items = append(out.Items, gatewaySettingsView(commerce.GatewaySettings{Provider: p.ID, Name: p.Name, AllCategories: true}))
		}
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
var gatewayPositiveID = regexp.MustCompile(`^[1-9][0-9]{0,19}$`)

func gatewayNonCredentialValues(provider GatewayProvider, stored map[string]string) map[string]string {
	values := map[string]string{}
	for _, f := range provider.Fields {
		if !gatewayCredentialField(f) {
			values[f.Key] = f.Default
			if v, ok := stored[f.Key]; ok {
				values[f.Key] = v
			}
		}
	}
	if _, exists := values["exchange_rate"]; exists && stored["exchange_rate"] == "" {
		values["exchange_rate"] = commerce.Currency(values["currency"]).DefaultGatewayRate()
	}
	return values
}

// Credential formats belong to a provider's field schema. Opaque secrets must
// preserve case, spaces and length; the UUID rule is never applied globally.
func validateGatewayCredentials(provider GatewayProvider, in GatewaySettingsInput) (map[string]string, map[string]string) {
	definitions := map[string]GatewayField{}
	for _, f := range provider.Fields {
		if gatewayCredentialField(f) {
			definitions[f.Key] = f
		}
	}
	updates := map[string]string{}
	fields := map[string]string{}
	for k, v := range in.Credentials {
		f, ok := definitions[k]
		if !ok {
			fields["credentials."+k] = "این credential برای درگاه انتخاب‌شده تعریف نشده است."
			continue
		}
		if v == "" {
			continue
		}
		limit := f.MaxLength
		if limit == 0 {
			limit = 4096
		}
		if !textBound(v, limit) || strings.TrimSpace(v) == "" {
			fields["credentials."+k] = "مقدار اطلاعات اتصال معتبر نیست یا از طول مجاز بیشتر است."
			continue
		}
		if f.Kind != "secret" {
			v = strings.TrimSpace(v)
		}
		switch f.Format {
		case "uuid":
			if !merchantUUID.MatchString(v) {
				fields["credentials."+k] = "شناسه پذیرنده باید UUID با ۳۶ نویسه باشد."
			}
			v = strings.ToLower(v)
		case "positive-integer":
			v = normalizeDigits(v)
			if !gatewayPositiveID.MatchString(v) {
				fields["credentials."+k] = "شناسه کلاینت باید عدد صحیح مثبت باشد."
			}
		}
		updates[k] = v
	}
	seen := map[string]bool{}
	for _, k := range in.ClearCredentials {
		_, known := definitions[k]
		if !known || seen[k] || in.Credentials[k] != "" {
			fields["credentials."+k] = "حذف و جایگزینی هم‌زمان یا حذف اطلاعات نامرتبط معتبر نیست."
		}
		seen[k] = true
	}
	return updates, fields
}

func validateGatewayValues(provider GatewayProvider, in GatewaySettingsInput) (map[string]string, map[string]string) {
	fields := map[string]string{}
	values := map[string]string{}
	for _, f := range provider.Fields {
		if !gatewayCredentialField(f) {
			values[f.Key] = f.Default
		}
	}
	for k, v := range in.Values {
		// Accept an old form submission without letting a historical global
		// description become the description of any future payment operation.
		if k == "description" {
			continue
		}
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
	currency := commerce.Currency(values["currency"])
	if !currency.Valid() {
		fields["values.currency"] = "واحد ارسالی باید تومان یا ریال باشد."
	}
	if _, supplied := in.Values["exchange_rate"]; !supplied {
		values["exchange_rate"] = currency.DefaultGatewayRate()
	}
	rate, rateErr := commerce.NormalizeGatewayRate(normalizeDigits(values["exchange_rate"]))
	if rateErr != nil {
		fields["values.exchange_rate"] = "نرخ باید عددی مثبت، حداکثر ۱۰۰۰۰۰۰ و با حداکثر ۶ رقم اعشار باشد."
	} else {
		values["exchange_rate"] = rate
	}
	if strings.TrimRight(values["api_base_url"], "/") != "https://next.zarinpal.com/api" {
		fields["values.api_base_url"] = "آدرس رسمی API زرین‌پال معتبر نیست."
	}
	values["api_base_url"] = "https://next.zarinpal.com/api"
	if values["channel"] != "ussd" && values["channel"] != "sms" {
		fields["values.channel"] = "روش دریافت کد حساب پذیرنده معتبر نیست."
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
		} else if in.Version != 0 {
			return nil, ErrInvalid
		}
		if index < 0 {
			for _, saved := range c.GatewaySettings {
				if saved.Provider == in.Provider {
					return nil, ErrConflict
				}
			}
		}
		provider, knownProvider := gatewayProviderDefinition(in.Provider)
		if !knownProvider {
			return nil, &GatewayFieldError{map[string]string{"provider": "ارائه‌دهنده درگاه پشتیبانی نمی‌شود."}}
		}
		values, fields := validateGatewayValues(provider, in)
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
		updates, credentialErrors := validateGatewayCredentials(provider, in)
		for k, v := range credentialErrors {
			fields[k] = v
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
		for k, v := range updates {
			credentials[k] = v
		}
		if err = s.sealGatewayCredentials(&g, credentials); err != nil {
			return nil, err
		}
		g.Version++
		g.Active = in.Active
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
