package backoffice

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/cherriva/marketplace/apps/api/internal/domain/commerce"
	"github.com/cherriva/marketplace/apps/api/internal/domain/identity"
	"math"
	"strings"
	"time"
)

type AddressInput struct {
	DetailVersion       int      `json:"detail_version,omitempty"`
	Title               string   `json:"title,omitempty"`
	FullAddress         string   `json:"full_address,omitempty"`
	HouseNumber         string   `json:"house_number,omitempty"`
	Unit                string   `json:"unit,omitempty"`
	RecipientSelf       *bool    `json:"recipient_self,omitempty"`
	RecipientNationalID string   `json:"recipient_national_id,omitempty"`
	Latitude            *float64 `json:"latitude,omitempty"`
	Longitude           *float64 `json:"longitude,omitempty"`

	ID         string `json:"id,omitempty"`
	Version    int    `json:"version"`
	Recipient  string `json:"recipient"`
	Phone      string `json:"phone"`
	Province   string `json:"province"`
	City       string `json:"city"`
	Street     string `json:"street"`
	PostalCode string `json:"postal_code"`
	Delete     bool   `json:"delete,omitempty"`
}

func validPhone(s string) bool {
	if len(s) != 11 || !strings.HasPrefix(s, "09") {
		return false
	}
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

var ErrProfileIncomplete = errors.New("recipient_profile_incomplete")
var ErrAddressIncomplete = errors.New("address_incomplete")

func validCoordinates(lat, lng *float64) bool {
	if lat == nil || lng == nil {
		return lat == nil && lng == nil
	}
	return !math.IsNaN(*lat) && !math.IsInf(*lat, 0) && !math.IsNaN(*lng) && !math.IsInf(*lng, 0) && *lat >= -90 && *lat <= 90 && *lng >= -180 && *lng <= 180
}
func addressFromInput(a identity.User, in AddressInput, id string) (commerce.Address, error) {
	phone, postal := normalizeDigits(in.Phone), normalizeDigits(in.PostalCode)
	nid := normalizeDigits(in.RecipientNationalID)
	if in.RecipientSelf == nil || strings.TrimSpace(in.HouseNumber) == "" {
		return commerce.Address{}, ErrAddressIncomplete
	}
	if *in.RecipientSelf {
		if strings.TrimSpace(a.Name) == "" || a.Name == "مشتری" || !identity.ValidNationalID(a.NationalID) || !validPhone(a.Phone) {
			return commerce.Address{}, ErrProfileIncomplete
		}
		in.Recipient, phone, nid = a.Name, a.Phone, a.NationalID
	}
	for _, v := range []struct {
		s   string
		max int
	}{{in.Recipient, 100}, {in.Province, 100}, {in.City, 100}, {in.Street, 1000}, {in.HouseNumber, 32}} {
		if strings.TrimSpace(v.s) == "" || !textBound(v.s, v.max) {
			return commerce.Address{}, ErrInvalid
		}
	}
	if !textBound(in.Title, 100) || !textBound(in.Unit, 32) || !textBound(in.FullAddress, 1500) || !validCoordinates(in.Latitude, in.Longitude) || !validPhone(phone) || !identity.ValidNationalID(nid) || len(postal) != 10 {
		return commerce.Address{}, ErrInvalid
	}
	for _, r := range postal {
		if r < '0' || r > '9' {
			return commerce.Address{}, ErrInvalid
		}
	}
	full := strings.TrimSpace(in.FullAddress)
	if full == "" {
		full = strings.Join([]string{strings.TrimSpace(in.Province), strings.TrimSpace(in.City), strings.TrimSpace(in.Street), "پلاک " + strings.TrimSpace(in.HouseNumber)}, "، ")
	}
	return commerce.Address{ID: id, UserID: a.ID, Version: in.Version + 1, DetailVersion: 2, Recipient: strings.TrimSpace(in.Recipient), Phone: phone, RecipientNationalID: nid, RecipientSelf: in.RecipientSelf, Province: strings.TrimSpace(in.Province), City: strings.TrimSpace(in.City), Street: strings.TrimSpace(in.Street), PostalCode: postal, Title: strings.TrimSpace(in.Title), FullAddress: full, HouseNumber: strings.TrimSpace(in.HouseNumber), Unit: strings.TrimSpace(in.Unit), Latitude: in.Latitude, Longitude: in.Longitude}, nil
}

// Fresh profile is authoritative for self; an old authenticated session is not.
func addressIdentity(w Workspace, a identity.User) identity.User {
	for _, u := range w.Users {
		if u.ID == a.ID {
			a.Name, a.Phone, a.NationalID = u.Name, u.Phone, u.NationalID
			break
		}
	}
	return a
}
func validateCheckoutAddress(w Workspace, a identity.User, address commerce.Address) error {
	if address.DetailVersion < 2 {
		return ErrAddressIncomplete
	}
	if address.RecipientSelf != nil && *address.RecipientSelf {
		current := addressIdentity(w, a)
		_, err := addressFromInput(current, AddressInput{RecipientSelf: address.RecipientSelf, HouseNumber: address.HouseNumber, Province: address.Province, City: address.City, Street: address.Street, PostalCode: address.PostalCode}, address.ID)
		if err != nil {
			return err
		}
	}
	return nil
}
func (s *Service) ReadAddresses(ctx context.Context, a identity.User) ([]commerce.Address, error) {
	w, e := s.repository.Snapshot(ctx)
	if e != nil {
		return nil, e
	}
	if e = customerActor(w, a); e != nil {
		return nil, e
	}
	result := []commerce.Address{}
	if w.Commerce != nil {
		for _, v := range w.Commerce.Addresses {
			if v.UserID == a.ID {
				result = append(result, v)
			}
		}
	}
	return result, nil
}
func (s *Service) SaveAddress(ctx context.Context, a identity.User, in AddressInput, key string) (json.RawMessage, error) {
	return s.mutateCommerce(ctx, "address:"+a.ID, key, in, func(w Workspace) error { return customerActor(w, a) }, func(w *Workspace, id string, now time.Time) (any, error) {
		c := commerceState(w)
		index := -1
		count := 0
		for i, v := range c.Addresses {
			if v.UserID == a.ID {
				count++
			}
			if v.ID == in.ID && v.UserID == a.ID {
				index = i
			}
		}
		if in.ID != "" && index < 0 {
			return nil, ErrNotFound
		}
		if index >= 0 && c.Addresses[index].Version != in.Version {
			return nil, ErrConflict
		}
		if in.Delete {
			if index < 0 {
				return nil, ErrNotFound
			}
			c.Addresses = append(c.Addresses[:index], c.Addresses[index+1:]...)
			return map[string]bool{"deleted": true}, nil
		}
		if index < 0 && count >= 20 {
			return nil, ErrInvalid
		}
		if index >= 0 {
			id = in.ID
		} else if in.Version != 0 {
			return nil, ErrConflict
		}
		address, e := addressFromInput(addressIdentity(*w, a), in, id)
		if e != nil {
			return nil, e
		}
		province, city, valid := canonicalDestination(c.Settings.Geography, address.Province, address.City)
		if !valid {
			return nil, ErrAddressUnavailable
		}
		address.Province, address.City = province, city
		if index < 0 {
			c.Addresses = append(c.Addresses, address)
		} else {
			c.Addresses[index] = address
		}
		return address, nil
	})
}
