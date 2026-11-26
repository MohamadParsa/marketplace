package backoffice

import (
	"context"
	"fmt"
	"slices"
	"sort"
	"strings"
	"unicode/utf8"

	"github.com/cherriva/marketplace/apps/api/internal/domain/commerce"
	"github.com/cherriva/marketplace/apps/api/internal/domain/identity"
	"github.com/cherriva/marketplace/apps/api/internal/domain/shipping"
)

type ShippingFieldError struct {
	Field   string
	Message string
}

func (e *ShippingFieldError) Error() string {
	return fmt.Sprintf("invalid shipping field: %s", e.Field)
}
func (e *ShippingFieldError) Unwrap() error { return ErrInvalid }
func shippingInvalid(field, message string) error {
	return &ShippingFieldError{Field: field, Message: message}
}

func shippingMethodInput(m commerce.ShippingMethod) ShippingMethodInput {
	if m.CollectionMode == "" {
		m.CollectionMode = "prepaid"
	}
	if m.DeliveryMode == "" {
		m.DeliveryMode = "range"
	}
	if m.CategoryMode == "" {
		m.CategoryMode = "all"
	}
	if m.DayBasis == "" {
		m.DayBasis = shipping.WorkingDays
	}
	fee := commerce.Toman(m.Fee)
	if m.CollectionMode == "collect" {
		fee = ""
	}
	return ShippingMethodInput{Schedule: m.Schedule, Version: m.Version, CollectionMode: m.CollectionMode, DeliveryMode: m.DeliveryMode, DailyCapacity: m.DailyCapacity, CategoryMode: m.CategoryMode, AllowedCategories: append([]string{}, m.AllowedCategories...), DeniedCategories: append([]string{}, m.DeniedCategories...), CityIDs: append([]string{}, m.CityIDs...), DayBasis: m.DayBasis, ID: m.ID, Name: m.Name, Active: m.Active, AllCities: m.AllCities, Cities: append([]string{}, m.Cities...), StoreIDs: append([]string{}, m.StoreIDs...), FeeToman: fee, MinDays: m.MinDays, MaxDays: m.MaxDays}
}

func shippingCityIDs(geography []commerce.Province, legacy []string) (ids, unresolved []string) {
	ids = []string{}
	unresolved = []string{}
	for _, name := range legacy {
		matches := []string{}
		for _, province := range geography {
			if !province.Active {
				continue
			}
			for _, city := range province.Cities {
				if city.Active && NormalizeSearch(city.Name) == NormalizeSearch(name) {
					matches = append(matches, city.ID)
				}
			}
		}
		if len(matches) != 1 {
			unresolved = append(unresolved, name)
			continue
		}
		if !slices.Contains(ids, matches[0]) {
			ids = append(ids, matches[0])
		}
	}
	return
}
func activeShippingCity(geography []commerce.Province, id string) bool {
	for _, province := range geography {
		if !province.Active {
			continue
		}
		for _, city := range province.Cities {
			if city.ID == id && city.Active {
				return true
			}
		}
	}
	return false
}

func normalizeShippingMethod(w Workspace, in ShippingMethodInput, newMethod bool) (commerce.ShippingMethod, error) {
	var out commerce.ShippingMethod
	in.Name = strings.TrimSpace(in.Name)
	if in.ID == "" || len(in.ID) > 80 || in.Version < 0 {
		return out, shippingInvalid("id", "شناسه یا نسخهٔ روش معتبر نیست.")
	}
	if in.Name == "" || utf8.RuneCountInString(in.Name) > 100 || strings.ContainsRune(in.Name, '\x00') {
		return out, shippingInvalid("name", "نام روش بین یک تا صد حرف باشد.")
	}
	if in.CollectionMode == "" {
		in.CollectionMode = "prepaid"
	}
	if in.CollectionMode != "prepaid" && in.CollectionMode != "collect" {
		return out, shippingInvalid("collection_mode", "شیوه دریافت هزینه را انتخاب کنید.")
	}
	if in.DeliveryMode == "" {
		in.DeliveryMode = "range"
	}
	if in.DeliveryMode != "range" && in.DeliveryMode != "slots" {
		return out, shippingInvalid("delivery_mode", "نوع زمان تحویل معتبر نیست.")
	}
	if in.DayBasis == "" {
		in.DayBasis = shipping.WorkingDays
	}
	if !shipping.ValidBasis(in.DayBasis) {
		return out, shippingInvalid("day_basis", "واحد مدت حمل معتبر نیست.")
	}
	if in.MinDays < 0 || in.MaxDays < in.MinDays || in.MaxDays > 60 {
		return out, shippingInvalid("min_days", "بازهٔ حمل باید از صفر تا شصت و حداقل کمتر یا مساوی حداکثر باشد.")
	}
	if in.DailyCapacity != nil && (*in.DailyCapacity < 0 || *in.DailyCapacity > 1000000) || newMethod && in.DailyCapacity == nil {
		return out, shippingInvalid("daily_capacity", "ظرفیت روزانه را به‌صورت عدد صحیح صفر تا یک میلیون تعیین کنید.")
	}
	fee := commerce.Amount(0)
	if in.CollectionMode == "collect" {
		if in.FeeToman != "" && in.FeeToman != "0" {
			return out, shippingInvalid("fee_toman", "پس‌کرایه مبلغ قابل وصول چریوا ندارد؛ مبلغ پیش‌کرایه را خالی کنید.")
		}
	} else {
		var err error
		fee, err = commerce.ParseToman(in.FeeToman)
		if err != nil {
			return out, shippingInvalid("fee_toman", "هزینهٔ پیش‌کرایه معتبر و نامنفی باشد.")
		}
	}
	if in.CategoryMode == "" {
		in.CategoryMode = "all"
	}
	if in.CategoryMode != "all" && in.CategoryMode != "selected" {
		return out, shippingInvalid("category_mode", "محدودهٔ دسته را انتخاب کنید.")
	}
	normalizeCategories := func(ids []string) ([]string, error) {
		if len(ids) > 2000 {
			return nil, ErrInvalid
		}
		out := []string{}
		for _, id := range ids {
			resolved, err := ResolvePublicReference("c", id)
			if err != nil || slices.Contains(out, resolved) {
				return nil, ErrInvalid
			}
			found := false
			for _, category := range w.Categories {
				if category.ID == resolved && !category.Archived {
					found = true
					break
				}
			}
			if !found {
				return nil, ErrInvalid
			}
			out = append(out, resolved)
		}
		sort.Strings(out)
		return out, nil
	}
	allowed, err := normalizeCategories(in.AllowedCategories)
	if err != nil {
		return out, shippingInvalid("allowed_categories", "دستهٔ مجاز نامعتبر یا تکراری است.")
	}
	denied, err := normalizeCategories(in.DeniedCategories)
	if err != nil {
		return out, shippingInvalid("denied_categories", "دستهٔ ممنوع نامعتبر یا تکراری است.")
	}
	if in.Active && in.CategoryMode == "selected" && len(allowed) == 0 {
		return out, shippingInvalid("allowed_categories", "برای روش فعال حداقل یک دسته مجاز تعیین کنید.")
	}
	geography := preparationSettings(w).Geography
	if len(in.Cities) > 2000 || len(in.CityIDs) > 2000 {
		return out, shippingInvalid("city_ids", "تعداد شهرها بیش از حد مجاز است.")
	}
	cities := append([]string{}, in.CityIDs...)
	if len(in.Cities) > 0 {
		mapped, unresolved := shippingCityIDs(geography, in.Cities)
		if len(unresolved) > 0 {
			return out, shippingInvalid("city_ids", "شهر متنی قدیمی مبهم یا نامعتبر است؛ شهر را از فهرست جغرافیا دوباره انتخاب کنید.")
		}
		for _, id := range mapped {
			if !slices.Contains(cities, id) {
				cities = append(cities, id)
			}
		}
	}
	sort.Strings(cities)
	for index, id := range cities {
		if !activeShippingCity(geography, id) || index > 0 && cities[index-1] == id {
			return out, shippingInvalid("city_ids", "شهر انتخاب‌شده نامعتبر، غیرفعال یا تکراری است.")
		}
	}
	if in.AllCities && len(cities) > 0 {
		return out, shippingInvalid("city_ids", "در حالت همهٔ شهرها فهرست انتخابی خالی باشد.")
	}
	if in.Active && !in.AllCities && len(cities) == 0 {
		return out, shippingInvalid("city_ids", "برای روش فعال حداقل یک شهر یا همهٔ شهرها را انتخاب کنید.")
	}
	stores := append([]string{}, in.StoreIDs...)
	sort.Strings(stores)
	if len(stores) > 10000 {
		return out, ErrInvalid
	}
	for index, id := range stores {
		if index > 0 && stores[index-1] == id {
			return out, ErrInvalid
		}
		if _, err := storeByID(&w, id); err != nil {
			return out, ErrInvalid
		}
	}
	schedule, err := shipping.NormalizeSchedule(in.Schedule)
	if err != nil || in.Active && in.DeliveryMode == "slots" && schedule == nil {
		return out, shippingInvalid("schedule", "برنامهٔ روزها، بازهٔ ساعت و ظرفیت را کامل کنید؛ زمان‌ها و انتساب‌ها نباید تکراری یا هم‌پوشان باشند.")
	}
	var capacity *int
	if in.DailyCapacity != nil {
		value := *in.DailyCapacity
		capacity = &value
	}
	out = commerce.ShippingMethod{Schedule: schedule, Version: in.Version, ID: in.ID, Name: in.Name, Active: in.Active, CollectionMode: in.CollectionMode, DeliveryMode: in.DeliveryMode, DailyCapacity: capacity, CategoryMode: in.CategoryMode, AllowedCategories: allowed, DeniedCategories: denied, CityIDs: cities, Cities: []string{}, StoreIDs: stores, AllCities: in.AllCities, Fee: fee, MinDays: in.MinDays, MaxDays: in.MaxDays, DayBasis: in.DayBasis}
	if in.Active && in.CategoryMode == "selected" {
		covered := false
		for _, category := range w.Categories {
			if ShippingAllowsCategory(w, out, category.ID) {
				covered = true
				break
			}
		}
		if !covered {
			return commerce.ShippingMethod{}, shippingInvalid("allowed_categories", "محدودیت ممنوعیت‌ها هیچ دستهٔ مجازی باقی نگذاشته است.")
		}
	}
	return out, nil
}

func shippingStillAllowed(w Workspace, q commerce.Quote) bool {
	for _, shipment := range q.Shipments {
		found := false
		for _, method := range preparationSettings(w).Shipping {
			effective, allowed := EffectiveStoreShipping(w, shipment.StoreID, method)
			if allowed && method.ID == shipment.Method.ID && storeShippingEligible(w, method, effective, q.Address, shipment.StoreID, q.Lines) && shippingTermsCurrent(w, shipment, effective) {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}

func ShippingAllowsCategory(w Workspace, method commerce.ShippingMethod, category string) bool {
	allowed := method.CategoryMode == "" || method.CategoryMode == "all"
	seen := map[string]bool{}
	for category != "" {
		if seen[category] {
			return false
		}
		seen[category] = true
		if slices.Contains(method.DeniedCategories, category) {
			return false
		}
		if slices.Contains(method.AllowedCategories, category) {
			allowed = true
		}
		found := false
		for _, item := range w.Categories {
			if item.ID == category && !item.Archived {
				category = item.ParentID
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return allowed
}

func ShippingMethodEligible(w Workspace, method commerce.ShippingMethod, address commerce.Address, storeID string, lines []commerce.LineSnapshot) bool {
	if !method.Active || len(method.StoreIDs) > 0 && !slices.Contains(method.StoreIDs, storeID) {
		return false
	}
	geography := preparationSettings(w).Geography
	provinceName, cityName, valid := canonicalDestination(geography, address.Province, address.City)
	if !valid {
		return false
	}
	cityID := ""
	for _, province := range geography {
		if province.Name != provinceName {
			continue
		}
		for _, city := range province.Cities {
			if city.Name == cityName {
				cityID = city.ID
			}
		}
	}
	if !method.AllCities {
		ids := method.CityIDs
		if len(method.Cities) > 0 {
			mapped, unresolved := shippingCityIDs(geography, method.Cities)
			if len(unresolved) > 0 {
				return false
			}
			ids = append(slices.Clone(ids), mapped...)
		}
		if !slices.Contains(ids, cityID) {
			return false
		}
	}
	for _, line := range lines {
		if line.StoreID != storeID {
			continue
		}
		product, err := findProduct(&w, line.ProductID)
		if err != nil || !ShippingAllowsCategory(w, method, product.CategoryID) {
			return false
		}
	}
	return true
}

type ShippingMethodFilter struct {
	Query, Active, CollectionMode, DeliveryMode, CategoryID, CityID string
	Page, Size                                                      int
}
type ShippingMethodRow struct {
	ShippingMethodInput
	UnresolvedCities []string `json:"unresolved_cities"`
}
type ShippingMethodDirectory struct {
	Rows      []ShippingMethodRow `json:"rows"`
	Total     int                 `json:"total"`
	Page      int                 `json:"page"`
	Size      int                 `json:"size"`
	CanWrite  bool                `json:"can_write"`
	Geography []commerce.Province `json:"geography"`
}

func (s *Service) ReadShippingMethods(ctx context.Context, a identity.User, f ShippingMethodFilter) (ShippingMethodDirectory, error) {
	out := ShippingMethodDirectory{Rows: []ShippingMethodRow{}, Page: f.Page, Size: f.Size}
	if utf8.RuneCountInString(f.Query) > 200 || f.Page < 0 || f.Size < 0 || f.Size > 100 || f.Active != "" && f.Active != "true" && f.Active != "false" || f.CollectionMode != "" && f.CollectionMode != "prepaid" && f.CollectionMode != "collect" || f.DeliveryMode != "" && f.DeliveryMode != "range" && f.DeliveryMode != "slots" {
		return out, ErrInvalid
	}
	w, err := s.repository.Snapshot(ctx)
	if err != nil {
		return out, err
	}
	if err = requireCommerceAdmin(w, a, CommerceSettingsRead); err != nil {
		return out, err
	}
	actor, err := currentActor(w, a)
	if err != nil {
		return out, err
	}
	out.CanWrite = allowsAction(actor, CommerceShippingWrite)
	settings := preparationSettings(w)
	out.Geography = publicGeography(settings.Geography)
	if f.CityID != "" && !activeShippingCity(settings.Geography, f.CityID) {
		return out, ErrInvalid
	}
	if f.CategoryID != "" && !categoryIDsValid(w, []string{f.CategoryID}) {
		return out, ErrInvalid
	}
	for _, method := range settings.Shipping {
		in := shippingMethodInput(method)
		if f.Query != "" && !strings.Contains(NormalizeSearch(in.Name+" "+in.ID), NormalizeSearch(f.Query)) || f.Active != "" && in.Active != (f.Active == "true") || f.CollectionMode != "" && in.CollectionMode != f.CollectionMode || f.DeliveryMode != "" && in.DeliveryMode != f.DeliveryMode || f.CategoryID != "" && !ShippingAllowsCategory(w, method, f.CategoryID) {
			continue
		}
		ids, unresolved := shippingCityIDs(settings.Geography, method.Cities)
		in.CityIDs = append(in.CityIDs, ids...)
		in.Cities = []string{}
		if f.CityID != "" && !in.AllCities && !slices.Contains(in.CityIDs, f.CityID) {
			continue
		}
		out.Rows = append(out.Rows, ShippingMethodRow{ShippingMethodInput: in, UnresolvedCities: unresolved})
	}
	sort.Slice(out.Rows, func(i, j int) bool {
		if out.Rows[i].Name == out.Rows[j].Name {
			return out.Rows[i].ID < out.Rows[j].ID
		}
		return out.Rows[i].Name < out.Rows[j].Name
	})
	out.Total = len(out.Rows)
	if out.Size == 0 {
		out.Size = 20
	}
	if out.Total == 0 {
		out.Page = 0
		return out, nil
	}
	if out.Page > (out.Total-1)/out.Size {
		out.Page = (out.Total - 1) / out.Size
	}
	start := out.Page * out.Size
	out.Rows = out.Rows[start:min(start+out.Size, out.Total)]
	return out, nil
}

func (s *Service) SaveShippingMethod(ctx context.Context, a identity.User, in ShippingMethodInput, create bool, key string) (ShippingMethodInput, error) {
	var out ShippingMethodInput
	if strings.TrimSpace(key) == "" || len(key) > 128 {
		return out, ErrInvalid
	}
	err := s.repository.Update(ctx, func(w *Workspace) error {
		if err := requireCommerceAdmin(*w, a, CommerceShippingWrite); err != nil {
			return err
		}
		if err := requireCommerceAdmin(*w, a, CommerceSettingsRead); err != nil {
			return err
		}
		method, err := normalizeShippingMethod(*w, in, create)
		if err != nil {
			return err
		}
		fingerprint := axisHash([]any{create, method})
		if w.Commerce == nil {
			w.Commerce = &commerce.State{Settings: commerce.DefaultSettings()}
		}
		settings := &w.Commerce.Settings
		for _, event := range w.Audit {
			if event.ID == key {
				if event.Actor != a.ID || event.Action != "shipping.method.save" || event.Target != in.ID || event.Fingerprint != fingerprint {
					return ErrConflict
				}
				for _, saved := range settings.Shipping {
					if saved.ID == in.ID {
						out = shippingMethodInput(saved)
						return nil
					}
				}
				return ErrNotFound
			}
		}
		index := slices.IndexFunc(settings.Shipping, func(m commerce.ShippingMethod) bool { return m.ID == in.ID })
		if create {
			if index >= 0 || in.Version != 0 {
				return ErrConflict
			}
			if len(settings.Shipping) >= 1000 {
				return ErrInvalid
			}
			method.Version = 1
			settings.Shipping = append(settings.Shipping, method)
		} else {
			if index < 0 {
				return ErrNotFound
			}
			if settings.Shipping[index].Version != in.Version {
				return ErrConflict
			}
			method.Version++
			settings.Shipping[index] = method
		}
		settings.Version++
		out = shippingMethodInput(method)
		w.Audit = append(w.Audit, Audit{ID: key, Actor: a.ID, Action: "shipping.method.save", Target: in.ID, Fingerprint: fingerprint, At: s.clock.Now()})
		return nil
	})
	return out, err
}
