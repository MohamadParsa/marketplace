package backoffice

import (
	"context"
	"encoding/json"
	"github.com/cherriva/marketplace/apps/api/internal/domain/commerce"
	"github.com/cherriva/marketplace/apps/api/internal/domain/identity"
	"strings"
	"time"
)

type AddressInput struct {
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
func addressFromInput(a identity.User, in AddressInput, id string) (commerce.Address, error) {
	phone := normalizeDigits(in.Phone)
	postal := normalizeDigits(in.PostalCode)
	for _, v := range []struct {
		s   string
		max int
	}{{in.Recipient, 100}, {in.Province, 100}, {in.City, 100}, {in.Street, 1000}} {
		if strings.TrimSpace(v.s) == "" || !textBound(v.s, v.max) {
			return commerce.Address{}, ErrInvalid
		}
	}
	if !validPhone(phone) || len(postal) != 10 {
		return commerce.Address{}, ErrInvalid
	}
	for _, r := range postal {
		if r < '0' || r > '9' {
			return commerce.Address{}, ErrInvalid
		}
	}
	return commerce.Address{ID: id, UserID: a.ID, Version: in.Version + 1, Recipient: strings.TrimSpace(in.Recipient), Phone: phone, Province: strings.TrimSpace(in.Province), City: strings.TrimSpace(in.City), Street: strings.TrimSpace(in.Street), PostalCode: postal}, nil
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
		address, e := addressFromInput(a, in, id)
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
