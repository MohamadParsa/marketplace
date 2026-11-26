package backoffice

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"slices"
	"sort"
	"strings"
	"time"

	"github.com/cherriva/marketplace/apps/api/internal/domain/catalog"
	"github.com/cherriva/marketplace/apps/api/internal/domain/identity"
)

const CategoryAxesConvert = "category.axes.convert"
const MaxAxisProducts = 1000
const MaxAxisVariants = 10000
const AxisPlanMinutes = 15

type AxisMapping struct {
	VariantID string                   `json:"variant_id"`
	Values    []catalog.AttributeValue `json:"values"`
}
type AxisConversionInput struct {
	CategoryID string        `json:"category_id"`
	Version    int           `json:"version"`
	AxisIDs    []string      `json:"axis_ids"`
	Mapping    []AxisMapping `json:"mapping"`
}
type AxisProduct struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}
type AxisSource struct {
	Category   catalog.Category              `json:"category"`
	Attributes []catalog.AttributeDefinition `json:"attributes"`
	Products   []AxisProduct                 `json:"products"`
	Variants   []catalog.Variant             `json:"variants"`
}
type AxisConflict struct {
	Code           string `json:"code"`
	VariantID      string `json:"variant_id,omitempty"`
	OtherVariantID string `json:"other_variant_id,omitempty"`
	AttributeID    string `json:"attribute_id,omitempty"`
}
type AxisImpact struct {
	Products         int `json:"products"`
	Variants         int `json:"variants"`
	Offers           int `json:"offers"`
	Media            int `json:"media"`
	Carts            int `json:"carts"`
	Quotes           int `json:"quotes"`
	ReservedOffers   int `json:"reserved_offers"`
	ReservedQuotes   int `json:"reserved_quotes"`
	PreservedAliases int `json:"preserved_aliases"`
}
type AxisPreview struct {
	Fingerprint string         `json:"fingerprint"`
	Impact      AxisImpact     `json:"impact"`
	Conflicts   []AxisConflict `json:"conflicts"`
	Ready       bool           `json:"ready"`
}
type AxisReceipt struct {
	PlanID     string     `json:"plan_id"`
	CategoryID string     `json:"category_id"`
	AppliedAt  time.Time  `json:"applied_at"`
	Impact     AxisImpact `json:"impact"`
}
type AxisPlan struct {
	Impact           AxisImpact                 `json:"impact"`
	ID               string                     `json:"id"`
	Key              string                     `json:"key"`
	Actor            string                     `json:"actor"`
	Input            AxisConversionInput        `json:"input"`
	Fingerprint      string                     `json:"fingerprint"`
	CreatedAt        time.Time                  `json:"created_at"`
	ExpiresAt        time.Time                  `json:"expires_at"`
	Receipt          *AxisReceipt               `json:"receipt,omitempty"`
	PreviousVariants []catalog.Variant          `json:"previous_variants,omitempty"`
	PreviousBindings []catalog.AttributeBinding `json:"previous_bindings,omitempty"`
}
type AxisPlanView struct {
	ID          string     `json:"id"`
	Fingerprint string     `json:"fingerprint"`
	ExpiresAt   time.Time  `json:"expires_at"`
	Impact      AxisImpact `json:"impact"`
}

func axisActor(w Workspace, a identity.User) (identity.User, error) {
	if a.Surface != identity.SurfaceBackoffice {
		return identity.User{}, ErrForbidden
	}
	current, e := currentActor(w, a)
	if e != nil || !allowsAction(current, CategoryAxesConvert) {
		return current, ErrForbidden
	}
	return current, nil
}
func axisSource(w Workspace, id string) (AxisSource, error) {
	out := AxisSource{Attributes: []catalog.AttributeDefinition{}, Products: []AxisProduct{}, Variants: []catalog.Variant{}}
	found := false
	for _, c := range w.Categories {
		if c.ID == id && !c.Archived {
			out.Category = c
			found = true
			break
		}
	}
	if !found {
		return out, ErrNotFound
	}
	for _, a := range w.Attributes {
		if a.Kind == "enum" {
			out.Attributes = append(out.Attributes, a)
		}
	}
	products := map[string]bool{}
	for _, p := range w.Products {
		if p.CategoryID == id {
			products[p.ID] = true
			out.Products = append(out.Products, AxisProduct{ID: p.ID, Name: p.Name})
		}
	}
	for _, v := range w.Variants {
		if products[v.ProductID] {
			out.Variants = append(out.Variants, v)
		}
	}
	if len(out.Products) > MaxAxisProducts || len(out.Variants) > MaxAxisVariants {
		return out, ErrInvalid
	}
	sort.Slice(out.Products, func(i, j int) bool { return out.Products[i].ID < out.Products[j].ID })
	sort.Slice(out.Variants, func(i, j int) bool { return out.Variants[i].ID < out.Variants[j].ID })
	return out, nil
}
func (s *Service) ReadAxisSource(ctx context.Context, a identity.User, id string) (AxisSource, error) {
	w, e := s.repository.Snapshot(ctx)
	if e != nil {
		return AxisSource{}, e
	}
	if _, e = axisActor(w, a); e != nil {
		return AxisSource{}, e
	}
	return axisSource(w, id)
}
func normalizeAxisInput(in AxisConversionInput) (AxisConversionInput, error) {
	out := in
	out.AxisIDs = append([]string{}, in.AxisIDs...)
	out.Mapping = append([]AxisMapping{}, in.Mapping...)
	if len(out.AxisIDs) > 100 || len(out.Mapping) > MaxAxisVariants || out.Version < 0 {
		return out, ErrInvalid
	}
	var e error
	for i := range out.AxisIDs {
		out.AxisIDs[i], e = ResolvePublicReference("a", out.AxisIDs[i])
		if e != nil {
			return out, e
		}
	}
	sort.Strings(out.AxisIDs)
	for i, id := range out.AxisIDs {
		if id == "" || i > 0 && out.AxisIDs[i-1] == id {
			return out, ErrInvalid
		}
	}
	for i := range out.Mapping {
		m := &out.Mapping[i]
		m.VariantID, e = ResolvePublicReference("v", m.VariantID)
		if e != nil {
			return out, e
		}
		m.Values, e = normalizeAttributeValues(m.Values)
		if e != nil {
			return out, e
		}
		if len(m.Values) > 100 {
			return out, ErrInvalid
		}
		sort.Slice(m.Values, func(i, j int) bool { return m.Values[i].AttributeID < m.Values[j].AttributeID })
	}
	sort.Slice(out.Mapping, func(i, j int) bool { return out.Mapping[i].VariantID < out.Mapping[j].VariantID })
	return out, nil
}
func axisHash(v any) string {
	b, _ := json.Marshal(v)
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:])
}
func analyzeAxes(w Workspace, in AxisConversionInput) (AxisPreview, error) {
	out := AxisPreview{Conflicts: []AxisConflict{}}
	source, e := axisSource(w, in.CategoryID)
	if e != nil {
		return out, e
	}
	if source.Category.Version != in.Version {
		return out, ErrConflict
	}
	products := map[string]bool{}
	for _, p := range source.Products {
		products[p.ID] = true
	}
	variants := map[string]catalog.Variant{}
	for _, v := range source.Variants {
		variants[v.ID] = v
	}
	definitions := map[string]catalog.AttributeDefinition{}
	for _, a := range w.Attributes {
		definitions[a.ID] = a
	}
	for _, id := range in.AxisIDs {
		a, ok := definitions[id]
		if !ok || a.Kind != "enum" {
			return out, ErrInvalid
		}
	}
	if len(axisBindings(source.Category.Bindings, in.AxisIDs)) > 100 {
		return out, ErrInvalid
	}
	if sameAxes(source.Category.Bindings, axisBindings(source.Category.Bindings, in.AxisIDs)) {
		out.Conflicts = append(out.Conflicts, AxisConflict{Code: "unchanged_axes"})
	}
	seen := map[string]bool{}
	combos := map[string]string{}
	for _, m := range in.Mapping {
		v, ok := variants[m.VariantID]
		if !ok || seen[m.VariantID] {
			return out, ErrInvalid
		}
		seen[m.VariantID] = true
		values := map[string]string{}
		for _, av := range m.Values {
			if _, exists := values[av.AttributeID]; exists {
				return out, ErrInvalid
			}
			values[av.AttributeID] = av.Value
			if !slices.Contains(in.AxisIDs, av.AttributeID) {
				return out, ErrInvalid
			}
		}
		valid := true
		for _, id := range in.AxisIDs {
			value, has := values[id]
			if !has || !slices.Contains(definitions[id].Options, value) {
				valid = false
				out.Conflicts = append(out.Conflicts, AxisConflict{Code: "missing_or_invalid_value", VariantID: v.ID, AttributeID: id})
			}
		}
		if valid {
			combo := v.ProductID + ":" + axisHash(m.Values)
			if previous, exists := combos[combo]; exists {
				out.Conflicts = append(out.Conflicts, AxisConflict{Code: "combination_collision", VariantID: v.ID, OtherVariantID: previous})
			} else {
				combos[combo] = v.ID
			}
		}
	}
	for _, v := range source.Variants {
		if !seen[v.ID] {
			out.Conflicts = append(out.Conflicts, AxisConflict{Code: "missing_variant_mapping", VariantID: v.ID})
		}
	}
	out.Impact.Products = len(source.Products)
	out.Impact.Variants = len(source.Variants)
	out.Impact.PreservedAliases = len(source.Variants)
	// Fingerprint only affected state; never return customer IDs or quote contents.
	state := []any{in, source.Category}
	for _, id := range in.AxisIDs {
		state = append(state, definitions[id])
	}
	for _, b := range source.Category.Bindings {
		state = append(state, definitions[b.AttributeID])
	}
	for _, p := range w.Products {
		if products[p.ID] {
			state = append(state, p)
		}
	}
	state = append(state, source.Variants)
	offers := map[string]bool{}
	media := map[string]bool{}
	for _, v := range source.Variants {
		if v.ImageID != "" {
			media[v.ImageID] = true
		}
	}
	for _, o := range w.Offers {
		if products[o.ProductID] {
			offers[o.ID] = true
			if v, ok := variants[o.VariantID]; !ok || v.ProductID != o.ProductID {
				out.Conflicts = append(out.Conflicts, AxisConflict{Code: "inconsistent_offer", VariantID: o.VariantID})
			}
			out.Impact.Offers++
			if o.Reserved > 0 {
				out.Impact.ReservedOffers++
			}
			state = append(state, o)
			out.Impact.PreservedAliases++
		}
	}
	out.Impact.Media = len(media)
	if w.Commerce != nil {
		for _, c := range w.Commerce.Carts {
			if c.MergedInto != "" {
				continue
			}
			affected := false
			for _, l := range c.Lines {
				affected = affected || offers[l.OfferID]
			}
			if affected {
				out.Impact.Carts++
				state = append(state, c)
			}
		}
		for _, q := range w.Commerce.Quotes {
			affected := false
			for _, l := range q.Lines {
				affected = affected || products[l.ProductID]
			}
			if affected {
				if q.Status == "preview" {
					out.Impact.Quotes++
					state = append(state, q)
				}
				if q.Status == "reserved" {
					out.Impact.ReservedQuotes++
					state = append(state, q)
				}
			}
		}
	}
	if out.Impact.ReservedOffers > 0 || out.Impact.ReservedQuotes > 0 {
		out.Conflicts = append(out.Conflicts, AxisConflict{Code: "active_reservation"})
	}
	out.Fingerprint = axisHash(state)
	out.Ready = len(out.Conflicts) == 0
	return out, nil
}
func axisBindings(original []catalog.AttributeBinding, axes []string) []catalog.AttributeBinding {
	out := append([]catalog.AttributeBinding{}, original...)
	position := 0
	for i := range out {
		out[i].Axis = slices.Contains(axes, out[i].AttributeID)
		if out[i].Position >= position {
			position = out[i].Position + 1
		}
	}
	for _, id := range axes {
		if !slices.ContainsFunc(out, func(b catalog.AttributeBinding) bool { return b.AttributeID == id }) {
			out = append(out, catalog.AttributeBinding{AttributeID: id, Axis: true, Position: min(position, 10000)})
			position++
		}
	}
	return out
}
func (s *Service) PreviewAxisConversion(ctx context.Context, a identity.User, input AxisConversionInput) (AxisPreview, error) {
	in, e := normalizeAxisInput(input)
	if e != nil {
		return AxisPreview{}, e
	}
	w, e := s.repository.Snapshot(ctx)
	if e != nil {
		return AxisPreview{}, e
	}
	if _, e = axisActor(w, a); e != nil {
		return AxisPreview{}, e
	}
	return analyzeAxes(w, in)
}
func (s *Service) PrepareAxisConversion(ctx context.Context, a identity.User, input AxisConversionInput, fingerprint, key string) (AxisPlanView, error) {
	var out AxisPlanView
	in, e := normalizeAxisInput(input)
	if e != nil {
		return out, e
	}
	if strings.TrimSpace(key) == "" || len(key) > 128 || len(fingerprint) != 64 {
		return out, ErrInvalid
	}
	id, e := s.identifiers.Token()
	if e != nil {
		return out, e
	}
	e = s.repository.Update(ctx, func(w *Workspace) error {
		if _, e := axisActor(*w, a); e != nil {
			return e
		}
		for _, p := range w.AxisPlans {
			if p.Key == key {
				if p.Actor != a.ID || p.Fingerprint != fingerprint || axisHash(p.Input) != axisHash(in) {
					return ErrConflict
				}
				out = AxisPlanView{ID: p.ID, Fingerprint: p.Fingerprint, ExpiresAt: p.ExpiresAt, Impact: p.Impact}
				return nil
			}
		}
		for _, event := range w.Audit {
			if event.ID == key {
				return ErrConflict
			}
		}
		preview, e := analyzeAxes(*w, in)
		if e != nil {
			return e
		}
		if !preview.Ready || preview.Fingerprint != fingerprint {
			return ErrConflict
		}
		now := s.clock.Now()
		p := AxisPlan{Impact: preview.Impact, ID: id, Key: key, Actor: a.ID, Input: in, Fingerprint: fingerprint, CreatedAt: now, ExpiresAt: now.Add(AxisPlanMinutes * time.Minute)}
		w.AxisPlans = append(w.AxisPlans, p)
		out = AxisPlanView{ID: id, Fingerprint: fingerprint, ExpiresAt: p.ExpiresAt, Impact: preview.Impact}
		w.Audit = append(w.Audit, Audit{ID: key, Actor: a.ID, Action: "category.axes.prepare", Target: in.CategoryID, Fingerprint: fingerprint, At: now})
		return nil
	})
	return out, e
}
func (s *Service) ApplyAxisConversion(ctx context.Context, a identity.User, id, fingerprint string, confirm bool) (AxisReceipt, error) {
	var out AxisReceipt
	if !confirm || id == "" || len(fingerprint) != 64 {
		return out, ErrInvalid
	}
	e := s.repository.Update(ctx, func(w *Workspace) error {
		if _, e := axisActor(*w, a); e != nil {
			return e
		}
		index := -1
		for i, p := range w.AxisPlans {
			if p.ID == id {
				index = i
				break
			}
		}
		if index < 0 {
			return ErrNotFound
		}
		plan := &w.AxisPlans[index]
		if plan.Actor != a.ID {
			return ErrForbidden
		}
		if plan.Fingerprint != fingerprint {
			return ErrConflict
		}
		if plan.Receipt != nil {
			out = *plan.Receipt
			return nil
		}
		now := s.clock.Now()
		if !now.Before(plan.ExpiresAt) {
			return ErrConflict
		}
		preview, e := analyzeAxes(*w, plan.Input)
		if e != nil {
			return e
		}
		if !preview.Ready || preview.Fingerprint != fingerprint {
			return ErrConflict
		}
		products := map[string]bool{}
		for i := range w.Products {
			p := &w.Products[i]
			if p.CategoryID == plan.Input.CategoryID {
				products[p.ID] = true
				p.VariantVersion++
				p.Version++
			}
		}
		mappings := map[string][]catalog.AttributeValue{}
		for _, m := range plan.Input.Mapping {
			mappings[m.VariantID] = m.Values
		}
		for i := range w.Categories {
			c := &w.Categories[i]
			if c.ID == plan.Input.CategoryID {
				plan.PreviousBindings = append([]catalog.AttributeBinding{}, c.Bindings...)
				c.Bindings = axisBindings(c.Bindings, plan.Input.AxisIDs)
				c.Version++
			}
		}
		labels := map[string]string{}
		for i := range w.Variants {
			v := &w.Variants[i]
			if !products[v.ProductID] {
				continue
			}
			plan.PreviousVariants = append(plan.PreviousVariants, *v)
			values := []string{}
			for _, a := range v.Values {
				values = append(values, a.Value)
			}
			labels[v.ID] = strings.Join(values, " / ")
			v.Values = append([]catalog.AttributeValue{}, mappings[v.ID]...)
			v.Version++
		}
		offers := map[string]string{}
		for _, o := range w.Offers {
			if products[o.ProductID] {
				offers[o.ID] = o.VariantID
			}
		}
		if w.Commerce != nil {
			for i := range w.Commerce.Carts {
				c := &w.Commerce.Carts[i]
				if c.MergedInto != "" {
					continue
				}
				changed := false
				for j := range c.Lines {
					l := &c.Lines[j]
					if v, ok := offers[l.OfferID]; ok {
						l.ReviewRequired = true
						l.PreviousVariantLabel = labels[v]
						changed = true
					}
				}
				if changed {
					c.Version++
					c.UpdatedAt = now
				}
			}
			for i := range w.Commerce.Quotes {
				q := &w.Commerce.Quotes[i]
				if q.Status != "preview" || q.PurchaseID != "" {
					continue
				}
				for _, l := range q.Lines {
					if products[l.ProductID] {
						q.Status = "expired"
						break
					}
				}
			}
		}
		out = AxisReceipt{PlanID: id, CategoryID: plan.Input.CategoryID, AppliedAt: now, Impact: preview.Impact}
		plan.Receipt = &out
		w.Audit = append(w.Audit, Audit{ID: fmt.Sprint("axis-apply:", id), Actor: a.ID, Action: CategoryAxesConvert, Target: plan.Input.CategoryID, Fingerprint: fingerprint, At: now})
		return nil
	})
	return out, e
}
