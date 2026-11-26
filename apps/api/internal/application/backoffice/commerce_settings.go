package backoffice

import (
	"context"
	"encoding/json"
	"github.com/cherriva/marketplace/apps/api/internal/domain/commerce"
	"github.com/cherriva/marketplace/apps/api/internal/domain/identity"
	"github.com/cherriva/marketplace/apps/api/internal/domain/shipping"
	"reflect"
	"sort"
	"strings"
	"time"
	_ "time/tzdata"
)

const CommerceSettingsRead = "commerce.settings.read"
const CommercePaymentWrite = "commerce.payment.settings.write"
const CommerceShippingWrite = "commerce.shipping.write"
const CommerceGeographyWrite = "commerce.geography.write"
const CommerceBusinessWrite = "commerce.business.write"
const CommerceOrdersRead = "commerce.orders.read"
const CommerceOrdersConfirm = "commerce.orders.confirm"
const CommerceDepositsRead = "commerce.deposits.read"
const CommerceDepositsReview = "commerce.deposits.review"
const CommerceRefundsResolve = "commerce.refunds.resolve"

type PaymentMethodInput struct {
	ID                 string   `json:"id"`
	Name               string   `json:"name"`
	Kind               string   `json:"kind"`
	Active             bool     `json:"active"`
	MinToman           string   `json:"min_toman"`
	MaxToman           string   `json:"max_toman"`
	AllCategories      bool     `json:"all_categories"`
	AllowedCategories  []string `json:"allowed_categories"`
	ExcludedCategories []string `json:"excluded_categories"`
	Instructions       string   `json:"instructions"`
}
type ShippingMethodInput struct {
	Version           int               `json:"version"`
	CollectionMode    string            `json:"collection_mode,omitempty"`
	DeliveryMode      string            `json:"delivery_mode,omitempty"`
	DailyCapacity     *int              `json:"daily_capacity,omitempty"`
	CategoryMode      string            `json:"category_mode,omitempty"`
	AllowedCategories []string          `json:"allowed_categories,omitempty"`
	DeniedCategories  []string          `json:"denied_categories,omitempty"`
	CityIDs           []string          `json:"city_ids,omitempty"`
	DayBasis          shipping.DayBasis `json:"day_basis,omitempty"`
	ID                string            `json:"id"`
	Name              string            `json:"name"`
	Active            bool              `json:"active"`
	AllCities         bool              `json:"all_cities"`
	Cities            []string          `json:"cities"`
	StoreIDs          []string          `json:"store_ids"`
	FeeToman          string            `json:"fee_toman"`
	MinDays           int               `json:"min_days"`
	MaxDays           int               `json:"max_days"`
}
type PaymentSettingsInput struct {
	Methods     []PaymentMethodInput `json:"methods"`
	Destination commerce.Destination `json:"destination"`
}
type ShippingSettingsInput struct {
	Timing         *shipping.TimingPolicy `json:"timing,omitempty"`
	Methods        []ShippingMethodInput  `json:"methods"`
	ClosedWeekdays []int                  `json:"closed_weekdays"`
	Holidays       []string               `json:"holidays"`
}
type CommerceSettingsInput struct {
	Geography *[]commerce.Province   `json:"geography,omitempty"`
	Version   int                    `json:"version"`
	Payment   *PaymentSettingsInput  `json:"payment,omitempty"`
	Shipping  *ShippingSettingsInput `json:"shipping,omitempty"`
	Business  *commerce.Business     `json:"business,omitempty"`
}
type CommerceSettingsView struct {
	ShippingTiming     shipping.TimingPolicy `json:"shipping_timing"`
	Geography          []commerce.Province   `json:"geography"`
	Version            int                   `json:"version"`
	Methods            []PaymentMethodInput  `json:"methods"`
	Shipping           []ShippingMethodInput `json:"shipping"`
	ClosedWeekdays     []int                 `json:"closed_weekdays"`
	Holidays           []string              `json:"holidays"`
	Destination        commerce.Destination  `json:"destination"`
	Business           commerce.Business     `json:"business"`
	Currency           string                `json:"currency"`
	ReservationMinutes int                   `json:"reservation_minutes"`
}

func methodInput(m commerce.PaymentMethod) PaymentMethodInput {
	max := ""
	if m.Max > 0 {
		max = commerce.Toman(m.Max)
	}
	return PaymentMethodInput{m.ID, m.Name, m.Kind, m.Active, commerce.Toman(m.Min), max, m.AllCategories, append([]string{}, m.AllowedCategories...), append([]string{}, m.ExcludedCategories...), m.Instructions}
}
func settingsView(s commerce.Settings) CommerceSettingsView {
	v := CommerceSettingsView{ShippingTiming: shippingTiming(s), Geography: append([]commerce.Province{}, s.Geography...), Version: s.Version, Methods: []PaymentMethodInput{}, Shipping: []ShippingMethodInput{}, ClosedWeekdays: s.ClosedWeekdays, Holidays: s.Holidays, Destination: s.Destination, Business: s.Business, Currency: "toman", ReservationMinutes: commerce.ReservationMinutes}
	for _, m := range s.Methods {
		v.Methods = append(v.Methods, methodInput(m))
	}
	for _, m := range s.Shipping {
		v.Shipping = append(v.Shipping, shippingMethodInput(m))
	}
	return v
}
func requireCommerceAdmin(w Workspace, a identity.User, permission string) error {
	if a.Surface != identity.SurfaceBackoffice {
		return ErrForbidden
	}
	current, e := currentActor(w, a)
	if e != nil {
		return e
	}
	if !allowsAction(current, permission) {
		return ErrForbidden
	}
	return nil
}
func (s *Service) ReadCommerceSettings(ctx context.Context, a identity.User) (CommerceSettingsView, error) {
	w, e := s.repository.Snapshot(ctx)
	if e != nil {
		return CommerceSettingsView{}, e
	}
	if e = requireCommerceAdmin(w, a, CommerceSettingsRead); e != nil {
		return CommerceSettingsView{}, e
	}
	return settingsView(commerceState(&w).Settings), nil
}
func categoryPath(w Workspace, id string) map[string]bool {
	path := map[string]bool{}
	for id != "" && !path[id] {
		path[id] = true
		next := ""
		for _, c := range w.Categories {
			if c.ID == id {
				next = c.ParentID
				break
			}
		}
		id = next
	}
	return path
}
func categoryIDsValid(w Workspace, ids []string) bool {
	seen := map[string]bool{}
	for _, id := range ids {
		if seen[id] {
			return false
		}
		seen[id] = true
		found := false
		for _, c := range w.Categories {
			if c.ID == id && !c.Archived {
				found = true
			}
		}
		if !found {
			return false
		}
	}
	return len(ids) <= 200
}
func methodEligible(w Workspace, m commerce.PaymentMethod, lines []commerce.LineSnapshot, total commerce.Amount) bool {
	if !m.Active || total < m.Min || m.Max > 0 && total > m.Max {
		return false
	}
	for _, l := range lines {
		p, e := findProduct(&w, l.ProductID)
		if e != nil {
			return false
		}
		path := categoryPath(w, p.CategoryID)
		for _, id := range m.ExcludedCategories {
			if path[id] {
				return false
			}
		}
		if !m.AllCategories {
			allowed := false
			for _, id := range m.AllowedCategories {
				allowed = allowed || path[id]
			}
			if !allowed {
				return false
			}
		}
	}
	return true
}
func validCard(s string) bool {
	s = normalizeDigits(s)
	if len(s) != 16 {
		return false
	}
	sum := 0
	allSame := true
	for i, r := range s {
		if r < '0' || r > '9' {
			return false
		}
		if i > 0 && byte(r) != s[0] {
			allSame = false
		}
		n := int(r - '0')
		if i%2 == 0 {
			n *= 2
			if n > 9 {
				n -= 9
			}
		}
		sum += n
	}
	return !allSame && sum%10 == 0
}
func parseMethods(w Workspace, inputs []PaymentMethodInput, d commerce.Destination) ([]commerce.PaymentMethod, error) {
	if len(inputs) != 3 {
		return nil, ErrInvalid
	}
	out := []commerce.PaymentMethod{}
	seen := map[string]bool{}
	for _, in := range inputs {
		if seen[in.ID] || in.ID != in.Kind || (in.Kind != "wallet" && in.Kind != "card-transfer" && in.Kind != "local") || strings.TrimSpace(in.Name) == "" || !textBound(in.Name, 80) || !textBound(in.Instructions, 1000) || !categoryIDsValid(w, in.AllowedCategories) || !categoryIDsValid(w, in.ExcludedCategories) {
			return nil, ErrInvalid
		}
		seen[in.ID] = true
		min, e := commerce.ParseToman(in.MinToman)
		if e != nil {
			return nil, ErrInvalid
		}
		var max commerce.Amount
		if in.MaxToman != "" {
			max, e = commerce.ParseToman(in.MaxToman)
			if e != nil || max <= 0 || max < min {
				return nil, ErrInvalid
			}
		}
		if in.Active && !in.AllCategories && len(in.AllowedCategories) == 0 {
			return nil, ErrInvalid
		}
		if in.Active && in.Kind == "card-transfer" && (!validCard(d.CardNumber) || strings.TrimSpace(d.Holder) == "" || strings.TrimSpace(d.Bank) == "") {
			return nil, ErrInvalid
		}
		if in.Active && in.Kind == "local" && strings.TrimSpace(in.Instructions) == "" {
			return nil, ErrInvalid
		}
		out = append(out, commerce.PaymentMethod{ID: in.ID, Name: strings.TrimSpace(in.Name), Kind: in.Kind, Active: in.Active, Min: min, Max: max, AllCategories: in.AllCategories, AllowedCategories: in.AllowedCategories, ExcludedCategories: in.ExcludedCategories, Instructions: in.Instructions})
	}
	return out, nil
}
func (s *Service) SaveCommerceSettings(ctx context.Context, a identity.User, in CommerceSettingsInput, key string) (json.RawMessage, error) {
	guard := func(w Workspace) error {
		if in.Payment == nil && in.Shipping == nil && in.Business == nil && in.Geography == nil {
			return ErrInvalid
		}
		for _, v := range []struct {
			yes        bool
			permission string
		}{{in.Payment != nil, CommercePaymentWrite}, {in.Shipping != nil, CommerceShippingWrite}, {in.Geography != nil, CommerceGeographyWrite}, {in.Business != nil, CommerceBusinessWrite}} {
			if v.yes {
				if e := requireCommerceAdmin(w, a, v.permission); e != nil {
					return e
				}
			}
		}
		return nil
	}
	return s.mutateCommerce(ctx, "commerce-settings:"+a.ID, key, in, guard, func(w *Workspace, id string, now time.Time) (any, error) {
		c := commerceState(w)
		if c.Settings.Version != in.Version {
			return nil, ErrConflict
		}
		if in.Geography != nil {
			directory, err := parseGeography(c.Settings.Geography, *in.Geography)
			if err != nil {
				return nil, err
			}
			c.Settings.Geography = directory
		}
		if in.Payment != nil {
			d := in.Payment.Destination
			d.CardNumber = normalizeDigits(d.CardNumber)
			if !textBound(d.Holder, 160) || !textBound(d.Bank, 100) {
				return nil, ErrInvalid
			}
			methods, e := parseMethods(*w, in.Payment.Methods, d)
			if e != nil {
				return nil, e
			}
			if d != c.Settings.Destination {
				if c.Settings.Destination.ID != "" {
					c.Settings.DestinationHistory = append(c.Settings.DestinationHistory, c.Settings.Destination)
				}
				d.ID = id
			}
			c.Settings.Destination = d
			oldMethods := c.Settings.Methods
			c.Settings.Methods = methods
			if c.Sales != nil {
				if err := validateSales(*w, *c.Sales); err != nil {
					c.Settings.Methods = oldMethods
					return nil, err
				}
			}
		}
		if in.Shipping != nil {
			v := in.Shipping
			if v.Timing != nil {
				timing, err := shipping.NormalizeTiming(*v.Timing)
				if err != nil {
					return nil, ErrInvalid
				}
				c.Settings.ShippingTiming = &timing
			}
			if len(v.Methods) > 1000 || len(v.ClosedWeekdays) > 6 || len(v.Holidays) > 366 {
				return nil, ErrInvalid
			}
			days := map[int]bool{}
			for _, d := range v.ClosedWeekdays {
				if d < 0 || d > 6 || days[d] {
					return nil, ErrInvalid
				}
				days[d] = true
			}
			calendarSettings := c.Settings
			calendarSettings.ClosedWeekdays = v.ClosedWeekdays
			calendarSettings.Holidays = v.Holidays
			calendar, err := siteShippingCalendar(calendarSettings)
			if err != nil {
				return nil, ErrInvalid
			}
			methods := []commerce.ShippingMethod{}
			seen := map[string]bool{}
			for _, m := range v.Methods {
				if seen[m.ID] {
					return nil, ErrInvalid
				}
				seen[m.ID] = true
				var previous *commerce.ShippingMethod
				for index := range c.Settings.Shipping {
					if c.Settings.Shipping[index].ID == m.ID {
						previous = &c.Settings.Shipping[index]
						break
					}
				}
				method, err := normalizeShippingMethod(*w, m, previous == nil)
				if err != nil {
					return nil, err
				}
				if previous != nil {
					if previous.Version != m.Version {
						return nil, ErrConflict
					}
					old, err := normalizeShippingMethod(*w, shippingMethodInput(*previous), false)
					if err != nil {
						return nil, err
					}
					if !reflect.DeepEqual(old, method) {
						method.Version++
					}
				} else {
					method.Version = 1
				}
				methods = append(methods, method)
			}
			for _, existing := range c.Settings.Shipping {
				if !seen[existing.ID] {
					return nil, shippingInvalid("methods", "روش موجود حذف نمی‌شود؛ برای توقف، آن را غیرفعال کنید.")
				}
			}
			c.Settings.Shipping = methods
			c.Settings.ClosedWeekdays = v.ClosedWeekdays
			c.Settings.Holidays = calendar.Holidays
		}
		if in.Business != nil {
			b, validationError := normalizeBusiness(*in.Business)
			if validationError != nil {
				return nil, validationError
			}
			for _, s := range []string{b.Name, b.SupportPhone, b.SupportEmail} {
				if !textBound(s, 160) {
					return nil, ErrInvalid
				}
			}
			for _, s := range []string{b.About, b.Contact, b.ShippingPolicy, b.ReturnPolicy, b.Privacy, b.Terms} {
				if !textBound(s, 20000) {
					return nil, ErrInvalid
				}
			}
			c.Settings.Business = b
		}
		c.Settings.Version++
		commerceAudit(w, id, a, "settings.update", "settings", "", now)
		return settingsView(c.Settings), nil
	})
}
func sortedStoreIDs(lines []commerce.LineSnapshot) []string {
	m := map[string]bool{}
	for _, l := range lines {
		m[l.StoreID] = true
	}
	ids := []string{}
	for id := range m {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	return ids
}
