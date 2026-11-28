package backoffice

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/mail"
	"sort"
	"strings"
	"unicode/utf8"

	"github.com/cherriva/marketplace/apps/api/internal/domain/catalog"
	"github.com/cherriva/marketplace/apps/api/internal/domain/commerce"
	"github.com/cherriva/marketplace/apps/api/internal/domain/identity"
	"github.com/cherriva/marketplace/apps/api/internal/domain/store"
)

const StoresRead = "stores.read"
const StoresWrite = "stores.write"
const StoresMembersRead = "stores.members.read"
const StoresMediaWrite = "stores.media.write"

type StoreProfileInput struct {
	Description string `json:"description"`
	Email       string `json:"email"`
	Province    string `json:"province"`
	Address     string `json:"address"`
	PostalCode  string `json:"postal_code"`
	Phone       string `json:"phone"`
	Mobile      string `json:"mobile"`
}
type StoreInput struct {
	ID      string            `json:"id"`
	Version int               `json:"version"`
	Name    string            `json:"name"`
	City    string            `json:"city"`
	Active  bool              `json:"active"`
	Profile StoreProfileInput `json:"profile"`
}
type StoreMember struct {
	UserID string   `json:"user_id"`
	Name   string   `json:"name"`
	Active bool     `json:"active"`
	Roles  []string `json:"roles"`
}
type AdminStoreRow struct {
	store.Store
	Version int           `json:"version"`
	Members []StoreMember `json:"members,omitempty"`
}
type StoreFilter struct {
	SalesMethod string `json:"sales_method"`
	ID          string `json:"id"`
	Name        string `json:"name"`
	Active      string `json:"active"`
	Province    string `json:"province"`
	City        string `json:"city"`
	Member      string `json:"member"`
	Page        int    `json:"page"`
	Size        int    `json:"size"`
}
type StoreDirectory struct {
	CanEligibility bool                    `json:"can_eligibility"`
	SalesMethods   []StoreSaleMethodChoice `json:"sales_methods"`
	Geography      []commerce.Province     `json:"geography"`
	Rows           []AdminStoreRow         `json:"rows"`
	Total          int                     `json:"total"`
	Page           int                     `json:"page"`
	Size           int                     `json:"size"`
	CanWrite       bool                    `json:"can_write"`
	CanMedia       bool                    `json:"can_media"`
	CanMembers     bool                    `json:"can_members"`
	CanUsers       bool                    `json:"can_users"`
}
type StoreMutation struct {
	Key         string `json:"key"`
	Actor       string `json:"actor"`
	Fingerprint string `json:"fingerprint"`
	StoreID     string `json:"store_id"`
}

func storeAdmin(w Workspace, a identity.User, permission string) (identity.User, error) {
	if a.Surface != identity.SurfaceBackoffice {
		return identity.User{}, ErrForbidden
	}
	current, e := currentActor(w, a)
	if e != nil || !allowsAction(current, permission) {
		return current, ErrForbidden
	}
	return current, nil
}
func privateStoreRow(w Workspace, st store.Store, members bool) AdminStoreRow {
	row := AdminStoreRow{Store: st, Version: st.Version}
	if members {
		row.Members = []StoreMember{}
		for _, m := range w.StoreMemberships {
			if m.StoreID != st.ID {
				continue
			}
			for _, u := range w.Users {
				if u.ID == m.UserID {
					member := StoreMember{UserID: u.ID, Name: u.Name, Active: u.Active, Roles: []string{}}
					if w.Access != nil {
						for _, g := range w.Access.Grants {
							if g.UserID == u.ID && g.StoreID == st.ID {
								for _, r := range w.Access.Roles {
									if r.ID == g.RoleID {
										member.Roles = append(member.Roles, r.Name)
									}
								}
							}
						}
					}
					row.Members = append(row.Members, member)
				}
			}
		}
		sort.Slice(row.Members, func(i, j int) bool { return row.Members[i].UserID < row.Members[j].UserID })
	}
	return row
}
func (s *Service) ListStores(ctx context.Context, a identity.User, f StoreFilter, export bool, ids []string) (StoreDirectory, error) {
	out := StoreDirectory{Rows: []AdminStoreRow{}}
	if len(f.SalesMethod) > 64 || f.Page < 0 || f.Size < 0 || f.Size > 100 || len(ids) > 10000 || (f.Active != "" && f.Active != "true" && f.Active != "false") {
		return out, ErrInvalid
	}
	for _, v := range []string{f.ID, f.Name, f.Province, f.City, f.Member} {
		if utf8.RuneCountInString(v) > 160 {
			return out, ErrInvalid
		}
	}
	w, e := s.repository.Snapshot(ctx)
	if e != nil {
		return out, e
	}
	actor, e := storeAdmin(w, a, StoresRead)
	if e != nil {
		return out, e
	}
	if export {
		if e = s.AuthorizeGridExport(ctx, a); e != nil {
			return out, e
		}
	}
	canMembers := allowsAction(actor, StoresMembersRead)
	if f.Member != "" && !canMembers {
		return out, ErrForbidden
	}
	if w.Commerce != nil {
		out.Geography = publicGeography(w.Commerce.Settings.Geography)
	}
	out.CanEligibility = allowsAction(actor, StoresEligibilityWrite)
	for _, m := range salesSettings(w).Methods {
		out.SalesMethods = append(out.SalesMethods, StoreSaleMethodChoice{ID: m.ID, Name: m.Name, Active: m.Active})
	}
	if f.SalesMethod != "" {
		known := false
		for _, m := range out.SalesMethods {
			known = known || m.ID == f.SalesMethod
		}
		if !known {
			return out, ErrInvalid
		}
	}
	out.CanWrite = allowsAction(actor, StoresWrite)
	out.CanMedia = allowsAction(actor, StoresMediaWrite)
	out.CanMembers = canMembers
	out.CanUsers = allowsAction(actor, UsersRead) || allowsAction(actor, CommerceWalletCredit)
	out.Page = f.Page
	out.Size = f.Size
	if out.Size == 0 {
		out.Size = 20
	}
	selected := map[string]bool{}
	for _, id := range ids {
		selected[id] = true
	}
	for _, st := range w.Stores {
		if f.SalesMethod != "" && !StoreAllowsMethod(st, f.SalesMethod) {
			continue
		}
		if f.ID != "" && st.ID != f.ID || !strings.Contains(foldUser(st.Name), foldUser(f.Name)) || !strings.Contains(foldUser(st.City), foldUser(f.City)) || f.Active != "" && st.Active != (f.Active == "true") {
			continue
		}
		if f.Province != "" && (st.Profile == nil || st.Profile.Province != f.Province) {
			continue
		}
		row := privateStoreRow(w, st, canMembers)
		if f.Member != "" {
			matched := false
			for _, m := range row.Members {
				matched = matched || m.UserID == f.Member || strings.Contains(foldUser(m.Name), foldUser(f.Member))
			}
			if !matched {
				continue
			}
		}
		if export && len(ids) > 0 && !selected[st.ID] {
			continue
		}
		out.Rows = append(out.Rows, row)
	}
	sort.Slice(out.Rows, func(i, j int) bool {
		if out.Rows[i].Name == out.Rows[j].Name {
			return out.Rows[i].ID < out.Rows[j].ID
		}
		return out.Rows[i].Name < out.Rows[j].Name
	})
	out.Total = len(out.Rows)
	if export {
		if out.Total > 10000 {
			return out, ErrInvalid
		}
		return out, nil
	}
	if out.Page > out.Total/out.Size {
		out.Page = out.Total / out.Size
	}
	if out.Total > 0 && out.Page*out.Size == out.Total {
		out.Page--
	}
	start := out.Page * out.Size
	end := start + out.Size
	if end > out.Total {
		end = out.Total
	}
	out.Rows = out.Rows[start:end]
	return out, nil
}
func validLandline(v string) bool {
	if v == "" {
		return true
	}
	v = strings.TrimPrefix(v, "+")
	if len(v) < 8 || len(v) > 15 {
		return false
	}
	for _, r := range v {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}
func normalizeStoreInput(w Workspace, in StoreInput) (StoreInput, error) {
	in.Name = strings.TrimSpace(in.Name)
	in.City = strings.TrimSpace(in.City)
	p := &in.Profile
	p.Description = strings.TrimSpace(p.Description)
	p.Email = strings.ToLower(strings.TrimSpace(p.Email))
	p.Province = strings.TrimSpace(p.Province)
	p.Address = strings.TrimSpace(p.Address)
	p.PostalCode = normalizeDigits(p.PostalCode)
	p.Phone = normalizeDigits(strings.TrimSpace(p.Phone))
	p.Mobile = identity.NormalizePhone(p.Mobile)
	if in.Version < 0 || in.Name == "" || utf8.RuneCountInString(in.Name) > 160 || utf8.RuneCountInString(in.City) > 100 || utf8.RuneCountInString(p.Description) > 4000 || utf8.RuneCountInString(p.Province) > 100 || utf8.RuneCountInString(p.Address) > 1500 || len(p.Email) > 254 || !validLandline(p.Phone) || p.Mobile != "" && !validPhone(p.Mobile) {
		return in, ErrInvalid
	}
	if p.Email != "" {
		a, e := mail.ParseAddress(p.Email)
		if e != nil || a.Address != p.Email {
			return in, ErrInvalid
		}
	}
	if p.PostalCode != "" {
		if len(p.PostalCode) != 10 {
			return in, ErrInvalid
		}
		for _, r := range p.PostalCode {
			if r < '0' || r > '9' {
				return in, ErrInvalid
			}
		}
	}
	if p.Province != "" {
		if w.Commerce == nil {
			return in, ErrInvalid
		}
		province, city, ok := canonicalDestination(w.Commerce.Settings.Geography, p.Province, in.City)
		if !ok {
			return in, ErrInvalid
		}
		p.Province, in.City = province, city
	} else if p.Address != "" || p.PostalCode != "" {
		return in, ErrInvalid
	}
	return in, nil
}
func storeByID(w *Workspace, id string) (*store.Store, error) {
	for i := range w.Stores {
		if w.Stores[i].ID == id {
			return &w.Stores[i], nil
		}
	}
	return nil, ErrNotFound
}
func (s *Service) SaveStore(ctx context.Context, a identity.User, in StoreInput, key string) (AdminStoreRow, error) {
	var out AdminStoreRow
	if strings.TrimSpace(key) == "" || len(key) > 128 {
		return out, ErrInvalid
	}
	id, e := s.identifiers.Token()
	if e != nil {
		return out, e
	}
	e = s.repository.Update(ctx, func(w *Workspace) error {
		actor, e := storeAdmin(*w, a, StoresWrite)
		if e != nil {
			return e
		}
		if !allowsAction(actor, StoresRead) {
			return ErrForbidden
		}
		if in.Profile.Province == "" && in.City != "" {
			old, oldErr := storeByID(w, in.ID)
			if oldErr != nil || old.City != in.City || old.Profile != nil && old.Profile.Province != "" {
				return ErrInvalid
			}
		}
		normalized, e := normalizeStoreInput(*w, in)
		if e != nil {
			return e
		}
		b, _ := json.Marshal(normalized)
		sum := sha256.Sum256(b)
		fingerprint := hex.EncodeToString(sum[:])
		for _, m := range w.StoreMutations {
			if m.Key == key {
				if m.Actor != a.ID || m.Fingerprint != fingerprint {
					return ErrConflict
				}
				st, e := storeByID(w, m.StoreID)
				if e != nil {
					return e
				}
				out = privateStoreRow(*w, *st, allowsAction(actor, StoresMembersRead))
				return nil
			}
		}
		for _, event := range w.Audit {
			if event.ID == key {
				return ErrConflict
			}
		}
		var st *store.Store
		if normalized.ID == "" {
			if normalized.Version != 0 {
				return ErrInvalid
			}
			w.Stores = append(w.Stores, store.Store{ID: id, Model: "shared"})
			st = &w.Stores[len(w.Stores)-1]
		} else {
			st, e = storeByID(w, normalized.ID)
			if e != nil {
				return e
			}
			if st.Version != normalized.Version {
				return ErrConflict
			}
		}
		image := ""
		if st.Profile != nil {
			image = st.Profile.ImageID
		}
		p := normalized.Profile
		st.Profile = &store.Profile{Description: p.Description, Email: p.Email, Province: p.Province, Address: p.Address, PostalCode: p.PostalCode, Phone: p.Phone, Mobile: p.Mobile, ImageID: image}
		st.Name = normalized.Name
		st.City = normalized.City
		st.Active = normalized.Active
		st.Version++
		out = privateStoreRow(*w, *st, allowsAction(actor, StoresMembersRead))
		w.StoreMutations = append(w.StoreMutations, StoreMutation{Key: key, Actor: a.ID, Fingerprint: fingerprint, StoreID: st.ID})
		w.Audit = append(w.Audit, Audit{ID: key, Actor: a.ID, Action: "stores.save", Target: st.ID, Fingerprint: fingerprint, At: s.clock.Now()})
		return nil
	})
	return out, e
}
func (s *Service) ChangeStoreImage(ctx context.Context, a identity.User, id string, version int, key string, input []byte, remove bool) (AdminStoreRow, error) {
	var out AdminStoreRow
	if a.Surface != identity.SurfaceBackoffice || !allowsAction(a, StoresMediaWrite) {
		return out, ErrForbidden
	}
	if version < 0 || key == "" || len(key) > 128 {
		return out, ErrInvalid
	}
	var data []byte
	var e error
	if !remove {
		data, e = normalizeIcon(input)
		if e != nil {
			return out, e
		}
	}
	mediaID, e := s.identifiers.Token()
	if e != nil {
		return out, e
	}
	digest := sha256.Sum256(input)
	b, _ := json.Marshal([]any{id, version, remove, hex.EncodeToString(digest[:])})
	sum := sha256.Sum256(b)
	fingerprint := hex.EncodeToString(sum[:])
	e = s.repository.Update(ctx, func(w *Workspace) error {
		actor, e := storeAdmin(*w, a, StoresMediaWrite)
		if e != nil {
			return e
		}
		if !allowsAction(actor, StoresRead) {
			return ErrForbidden
		}
		st, e := storeByID(w, id)
		if e != nil {
			return e
		}
		for _, event := range w.Audit {
			if event.ID == key {
				if event.Actor == a.ID && event.Action == "stores.image" && event.Target == id && event.Fingerprint == fingerprint {
					out = privateStoreRow(*w, *st, allowsAction(actor, StoresMembersRead))
					return nil
				}
				return ErrConflict
			}
		}
		if st.Version != version {
			return ErrConflict
		}
		if st.Profile == nil {
			st.Profile = &store.Profile{}
		}
		st.Profile.ImageID = ""
		if !remove {
			st.Profile.ImageID = mediaID
			w.IconMedia = append(w.IconMedia, catalog.IconMedia{ID: mediaID, Data: data})
		}
		st.Version++
		out = privateStoreRow(*w, *st, allowsAction(actor, StoresMembersRead))
		collectUnusedIcons(w)
		w.Audit = append(w.Audit, Audit{ID: key, Actor: a.ID, Action: "stores.image", Target: id, Fingerprint: fingerprint, At: s.clock.Now()})
		return nil
	})
	return out, e
}
func (s *Service) ReadStoreImage(ctx context.Context, a identity.User, id string) ([]byte, error) {
	w, e := s.repository.Snapshot(ctx)
	if e != nil {
		return nil, e
	}
	if _, e = storeAdmin(w, a, StoresRead); e != nil {
		return nil, e
	}
	st, e := storeByID(&w, id)
	if e != nil {
		return nil, e
	}
	if st.Profile == nil || st.Profile.ImageID == "" {
		return nil, ErrNotFound
	}
	for _, m := range w.IconMedia {
		if m.ID == st.Profile.ImageID {
			return s.catalogMediaBytes(ctx, m)
		}
	}
	return nil, ErrNotFound
}
