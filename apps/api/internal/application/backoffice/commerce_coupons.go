package backoffice

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/cherriva/marketplace/apps/api/internal/domain/commerce"
	"github.com/cherriva/marketplace/apps/api/internal/domain/identity"
)

const CouponsRead = "commerce.coupons.read"
const CouponsCreate = "commerce.coupons.create"
const CouponsEdit = "commerce.coupons.edit"
const CouponsActivate = "commerce.coupons.activate"

var ErrCoupon = errors.New("coupon_not_eligible")
var couponCodePattern = regexp.MustCompile(`^[A-Z0-9][A-Z0-9-]{2,63}$`)

type CouponInput struct {
	ID              string     `json:"id"`
	Version         int        `json:"version"`
	Code            string     `json:"code"`
	Name            string     `json:"name"`
	Active          bool       `json:"active"`
	Kind            string     `json:"kind"`
	Value           string     `json:"value"`
	Target          string     `json:"target"`
	AllCategories   bool       `json:"all_categories"`
	Categories      []string   `json:"categories"`
	MinToman        string     `json:"min_toman"`
	MaxToman        string     `json:"max_toman"`
	CapToman        string     `json:"cap_toman"`
	StartsAt        *time.Time `json:"starts_at"`
	EndsAt          *time.Time `json:"ends_at"`
	GlobalLimit     int        `json:"global_limit"`
	UserLimit       int        `json:"user_limit"`
	GlobalUnlimited bool       `json:"global_unlimited"`
	UserUnlimited   bool       `json:"user_unlimited"`
	PlatformBPS     int        `json:"platform_bps"`
}
type CouponView struct {
	CouponInput
	Reserved int `json:"reserved"`
	Consumed int `json:"consumed"`
}

func parsePercent(s string) (int, error) {
	if s == "" || len(s) > 6 {
		return 0, ErrInvalid
	}
	parts := strings.Split(s, ".")
	if len(parts) > 2 {
		return 0, ErrInvalid
	}
	for _, p := range parts {
		if p == "" {
			return 0, ErrInvalid
		}
		for _, r := range p {
			if r < '0' || r > '9' {
				return 0, ErrInvalid
			}
		}
	}
	if len(parts) == 2 && len(parts[1]) > 2 {
		return 0, ErrInvalid
	}
	n, _ := strconv.Atoi(parts[0])
	n *= 100
	if len(parts) == 2 {
		f, _ := strconv.Atoi(parts[1] + strings.Repeat("0", 2-len(parts[1])))
		n += f
	}
	if n < 1 || n > 10000 {
		return 0, ErrInvalid
	}
	return n, nil
}
func couponInput(c commerce.Coupon) CouponInput {
	value := commerce.Toman(c.Value)
	if c.Kind == "percent" {
		value = fmt.Sprintf("%d.%02d", c.Value/100, c.Value%100)
	}
	return CouponInput{ID: c.ID, Version: c.Version, Code: c.Code, Name: c.Name, Active: c.Active, Kind: c.Kind, Value: value, Target: c.Target, AllCategories: c.AllCategories, Categories: append([]string{}, c.Categories...), MinToman: commerce.Toman(c.Min), MaxToman: commerce.Toman(c.Max), CapToman: commerce.Toman(c.Cap), StartsAt: c.StartsAt, EndsAt: c.EndsAt, GlobalLimit: c.GlobalLimit, UserLimit: c.UserLimit, GlobalUnlimited: c.GlobalLimit == 0, UserUnlimited: c.UserLimit == 0, PlatformBPS: c.PlatformBPS}
}
func couponCounts(c *commerce.State, id, user, except string) (held, consumed int) {
	for _, q := range c.Quotes {
		if q.ID == except || q.Discount == nil || q.Discount.Coupon.ID != id || (user != "" && q.UserID != user) {
			continue
		}
		if q.Status == "reserved" {
			held++
		} else if q.Status == "consumed" {
			consumed++
		}
	}
	return
}
func couponView(c *commerce.State, coupon commerce.Coupon) CouponView {
	h, u := couponCounts(c, coupon.ID, "", "")
	return CouponView{couponInput(coupon), h, u}
}

type CouponDirectory struct {
	Items       []CouponView    `json:"items"`
	Permissions map[string]bool `json:"permissions"`
}

func (s *Service) ReadCoupons(ctx context.Context, a identity.User) (*CouponDirectory, error) {
	if e := s.ReleaseExpiredCommerce(ctx); e != nil {
		return nil, e
	}
	w, e := s.repository.Snapshot(ctx)
	if e != nil {
		return nil, e
	}
	if e = requireCommerceAdmin(w, a, CouponsRead); e != nil {
		return nil, e
	}
	c := commerceState(&w)
	result := []CouponView{}
	for _, v := range c.Coupons {
		result = append(result, couponView(c, v))
	}
	actor, _ := currentActor(w, a)
	permissions := map[string]bool{}
	for _, p := range []string{CouponsCreate, CouponsEdit, CouponsActivate} {
		permissions[p] = allowsAction(actor, p)
	}
	return &CouponDirectory{Items: result, Permissions: permissions}, nil
}
func (s *Service) SaveCoupon(ctx context.Context, a identity.User, in CouponInput, key string) (json.RawMessage, error) {
	permission := CouponsEdit
	if in.ID == "" {
		permission = CouponsCreate
	}
	return s.mutateCommerce(ctx, "coupon-save:"+a.ID, key, in, func(w Workspace) error { return requireCommerceAdmin(w, a, permission) }, func(w *Workspace, id string, now time.Time) (any, error) {
		c := commerceState(w)
		index := -1
		for i, v := range c.Coupons {
			if v.ID == in.ID {
				index = i
			}
			if strings.EqualFold(v.Code, strings.TrimSpace(in.Code)) && v.ID != in.ID {
				return nil, ErrConflict
			}
		}
		if in.ID != "" && index < 0 {
			return nil, ErrNotFound
		}
		if index >= 0 && c.Coupons[index].Version != in.Version {
			return nil, ErrConflict
		}
		wasActive := index >= 0 && c.Coupons[index].Active
		if in.Active != wasActive {
			if e := requireCommerceAdmin(*w, a, CouponsActivate); e != nil {
				return nil, e
			}
		}
		v := commerce.Coupon{ID: id, Version: 1, Code: strings.ToUpper(strings.TrimSpace(in.Code)), Name: strings.TrimSpace(in.Name), Active: in.Active, Kind: in.Kind, Target: in.Target, AllCategories: in.AllCategories, Categories: append([]string{}, in.Categories...), StartsAt: in.StartsAt, EndsAt: in.EndsAt, GlobalLimit: in.GlobalLimit, UserLimit: in.UserLimit, PlatformBPS: in.PlatformBPS}
		if !couponCodePattern.MatchString(v.Code) || v.Name == "" || !textBound(v.Name, 120) || v.PlatformBPS < 0 || v.PlatformBPS > 10000 || (v.Target != "items" && v.Target != "order" && v.Target != "shipping") || !categoryIDsValid(*w, v.Categories) || (!v.AllCategories && len(v.Categories) == 0) || (v.AllCategories && len(v.Categories) != 0) {
			return nil, ErrInvalid
		}
		if in.GlobalUnlimited {
			if in.GlobalLimit != 0 {
				return nil, ErrInvalid
			}
		} else if in.GlobalLimit < 1 || in.GlobalLimit > 1000000000 {
			return nil, ErrInvalid
		}
		if in.UserUnlimited {
			if in.UserLimit != 0 {
				return nil, ErrInvalid
			}
		} else if in.UserLimit < 1 || in.UserLimit > 1000000000 {
			return nil, ErrInvalid
		}
		if v.StartsAt != nil && v.EndsAt != nil && !v.StartsAt.Before(*v.EndsAt) {
			return nil, ErrInvalid
		}
		var e error
		if v.Kind == "percent" {
			n, err := parsePercent(normalizeDigits(in.Value))
			if err != nil {
				return nil, err
			}
			v.Value = commerce.Amount(n)
		} else if v.Kind == "fixed" {
			v.Value, e = commerce.ParseToman(normalizeDigits(in.Value))
			if e != nil || v.Value <= 0 {
				return nil, ErrInvalid
			}
		} else {
			return nil, ErrInvalid
		}
		for _, f := range []struct {
			in  string
			out *commerce.Amount
		}{{in.MinToman, &v.Min}, {in.MaxToman, &v.Max}, {in.CapToman, &v.Cap}} {
			text := normalizeDigits(f.in)
			if text == "" {
				text = "0"
			}
			*f.out, e = commerce.ParseToman(text)
			if e != nil {
				return nil, ErrInvalid
			}
		}
		if v.Max > 0 && v.Max < v.Min || (v.Kind == "fixed" && v.Cap > 0) {
			return nil, ErrInvalid
		}
		sort.Strings(v.Categories)
		if index >= 0 {
			v.ID = in.ID
			v.Version = in.Version + 1
			c.Coupons[index] = v
		} else {
			c.Coupons = append(c.Coupons, v)
		}
		commerceAudit(w, id, a, "coupon.save", v.ID, "", now)
		return couponView(c, v), nil
	})
}

// Each allocation key identifies a unit or a shipment and survives later rule changes.
func couponWeights(w Workspace, q commerce.Quote, rule commerce.Coupon) ([]commerce.WeightedAmount, map[string]string, error) {
	weights := []commerce.WeightedAmount{}
	stores := map[string]bool{}
	owners := map[string]string{}
	for _, l := range q.Lines {
		eligible := rule.AllCategories
		for _, p := range w.Products {
			if p.ID == l.ProductID {
				path := categoryPath(w, p.CategoryID)
				for _, id := range rule.Categories {
					eligible = eligible || path[id]
				}
				break
			}
		}
		if !eligible {
			continue
		}
		stores[l.StoreID] = true
		if rule.Target != "shipping" {
			for i := 0; i < l.Quantity; i++ {
				key := fmt.Sprintf("item:%s:%03d", l.OfferID, i)
				weights = append(weights, commerce.WeightedAmount{Key: key, Amount: l.UnitPrice})
				owners[key] = l.StoreID
			}
		}
	}
	if rule.Target != "items" {
		for _, ship := range q.Shipments {
			if stores[ship.StoreID] {
				key := "shipping:" + ship.StoreID
				weights = append(weights, commerce.WeightedAmount{Key: key, Amount: ship.Method.Fee})
				owners[key] = ship.StoreID
			}
		}
	}
	return weights, owners, nil
}
func evaluateCoupon(w Workspace, q commerce.Quote, code, except string, now time.Time) (*commerce.DiscountUsage, error) {
	code = strings.ToUpper(strings.TrimSpace(code))
	if code == "" {
		return nil, nil
	}
	c := commerceState(&w)
	var rule *commerce.Coupon
	for i := range c.Coupons {
		if c.Coupons[i].Code == code {
			rule = &c.Coupons[i]
			break
		}
	}
	if rule == nil || !rule.Active || (rule.StartsAt != nil && now.Before(*rule.StartsAt)) || (rule.EndsAt != nil && !now.Before(*rule.EndsAt)) {
		return nil, ErrCoupon
	}
	h, u := couponCounts(c, rule.ID, "", except)
	uh, uu := couponCounts(c, rule.ID, q.UserID, except)
	if (rule.GlobalLimit > 0 && h+u >= rule.GlobalLimit) || (rule.UserLimit > 0 && uh+uu >= rule.UserLimit) {
		return nil, ErrCoupon
	}
	weights, owners, e := couponWeights(w, q, *rule)
	if e != nil {
		return nil, e
	}
	var base commerce.Amount
	for _, v := range weights {
		base, e = commerce.Add(base, v.Amount)
		if e != nil {
			return nil, ErrInvalid
		}
	}
	if base <= 0 || base < rule.Min || (rule.Max > 0 && base > rule.Max) {
		return nil, ErrCoupon
	}
	amount := min(rule.Value, base)
	if rule.Kind == "percent" {
		amount, e = commerce.Percent(base, int(rule.Value))
		if e != nil {
			return nil, ErrInvalid
		}
		if rule.Cap > 0 {
			amount = min(amount, rule.Cap)
		}
	}
	if amount <= 0 {
		return nil, ErrCoupon
	}
	allocations, e := commerce.Allocate(amount, weights)
	if e != nil {
		return nil, ErrInvalid
	}
	usage := &commerce.DiscountUsage{Coupon: *rule, Amount: amount, Allocations: []commerce.DiscountAllocation{}}
	// Split the complete funded amount once, then allocate the platform share; unit rounding sums exactly.
	platform, _ := commerce.Percent(amount, rule.PlatformBPS)
	fundingWeights := []commerce.WeightedAmount{}
	for _, v := range weights {
		fundingWeights = append(fundingWeights, commerce.WeightedAmount{Key: v.Key, Amount: allocations[v.Key]})
	}
	platforms, _ := commerce.Allocate(platform, fundingWeights)
	for _, v := range weights {
		n := allocations[v.Key]
		usage.Allocations = append(usage.Allocations, commerce.DiscountAllocation{Key: v.Key, StoreID: owners[v.Key], Amount: n, Platform: platforms[v.Key], Seller: n - platforms[v.Key]})
	}
	return usage, nil
}

type CouponPreviewInput struct {
	CouponID      string `json:"coupon_id"`
	OfferID       string `json:"offer_id"`
	Quantity      int    `json:"quantity"`
	ShippingToman string `json:"shipping_toman"`
}

func (s *Service) PreviewCoupon(ctx context.Context, a identity.User, in CouponPreviewInput) (map[string]string, error) {
	if e := s.ReleaseExpiredCommerce(ctx); e != nil {
		return nil, e
	}
	w, e := s.repository.Snapshot(ctx)
	if e != nil {
		return nil, e
	}
	if e = requireCommerceAdmin(w, a, CouponsRead); e != nil {
		return nil, e
	}
	c := commerceState(&w)
	c.Coupons = append([]commerce.Coupon{}, c.Coupons...)
	var code string
	for i := range c.Coupons {
		if c.Coupons[i].ID == in.CouponID {
			code = c.Coupons[i].Code
			c.Coupons[i].Active = true
		}
	}
	if code == "" {
		return nil, ErrNotFound
	}
	o, e := findOffer(w, in.OfferID)
	if e != nil {
		return nil, e
	}
	subtotal, e := commerce.Multiply(commerce.Amount(o.Price), in.Quantity)
	if e != nil || in.Quantity < 1 {
		return nil, ErrInvalid
	}
	shipping, e := commerce.ParseToman(normalizeDigits(in.ShippingToman))
	if e != nil {
		return nil, ErrInvalid
	}
	total, e := commerce.Add(subtotal, shipping)
	if e != nil {
		return nil, ErrInvalid
	}
	q := commerce.Quote{UserID: a.ID, Subtotal: subtotal, ShippingTotal: shipping, Total: total, Lines: []commerce.LineSnapshot{{OfferID: o.ID, ProductID: o.ProductID, StoreID: o.StoreID, UnitPrice: commerce.Amount(o.Price), Quantity: in.Quantity}}, Shipments: []commerce.ShipmentSnapshot{{StoreID: o.StoreID, Method: commerce.ShippingMethod{Fee: shipping}}}}
	usage, e := evaluateCoupon(w, q, code, "", s.clock.Now())
	if e != nil {
		return nil, e
	}
	return map[string]string{"subtotal_toman": commerce.Toman(subtotal), "shipping_toman": commerce.Toman(shipping), "discount_toman": commerce.Toman(usage.Amount), "total_toman": commerce.Toman(total - usage.Amount), "note": "پیش‌نمایش قانون ذخیره‌شده؛ غیرفعال بودن نادیده گرفته شد، تاریخ و محدودیت‌ها برقرارند؛ بدون مصرف سهمیه"}, nil
}

type CouponActivationInput struct {
	ID      string `json:"id"`
	Version int    `json:"version"`
	Active  bool   `json:"active"`
}

func (s *Service) ActivateCoupon(ctx context.Context, a identity.User, in CouponActivationInput, key string) (json.RawMessage, error) {
	return s.mutateCommerce(ctx, "coupon-activate:"+a.ID, key, in, func(w Workspace) error { return requireCommerceAdmin(w, a, CouponsActivate) }, func(w *Workspace, id string, now time.Time) (any, error) {
		c := commerceState(w)
		for i := range c.Coupons {
			v := &c.Coupons[i]
			if v.ID != in.ID {
				continue
			}
			if v.Version != in.Version {
				return nil, ErrConflict
			}
			v.Active = in.Active
			v.Version++
			commerceAudit(w, id, a, "coupon.activate", v.ID, "", now)
			return couponView(c, *v), nil
		}
		return nil, ErrNotFound
	})
}
