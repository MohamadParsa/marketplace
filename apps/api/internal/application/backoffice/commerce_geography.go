package backoffice

import (
	"errors"
	"github.com/cherriva/marketplace/apps/api/internal/domain/commerce"
	"sort"
	"strings"
)

var ErrAddressUnavailable = errors.New("address_unavailable")

func publicGeography(provinces []commerce.Province) []commerce.Province {
	out := []commerce.Province{}
	for _, p := range provinces {
		if !p.Active {
			continue
		}
		cities := []commerce.City{}
		for _, c := range p.Cities {
			if c.Active {
				cities = append(cities, c)
			}
		}
		if len(cities) == 0 {
			continue
		}
		p.Cities = cities
		sort.Slice(p.Cities, func(i, j int) bool { return p.Cities[i].Name < p.Cities[j].Name })
		out = append(out, p)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}
func canonicalDestination(provinces []commerce.Province, province, city string) (string, string, bool) {
	for _, p := range provinces {
		if p.Active && NormalizeSearch(p.Name) == NormalizeSearch(province) {
			for _, c := range p.Cities {
				if c.Active && NormalizeSearch(c.Name) == NormalizeSearch(city) {
					return p.Name, c.Name, true
				}
			}
		}
	}
	return "", "", false
}
func parseGeography(previous, proposed []commerce.Province) ([]commerce.Province, error) {
	if len(proposed) > 100 {
		return nil, ErrInvalid
	}
	ids, names, parents := map[string]bool{}, map[string]bool{}, map[string]string{}
	provinceIDs := map[string]bool{}
	total := 0
	out := append([]commerce.Province{}, proposed...)
	for i := range out {
		p := &out[i]
		p.Name = strings.TrimSpace(p.Name)
		if p.ID == "" || len(p.ID) > 80 || ids[p.ID] || p.Name == "" || !textBound(p.Name, 100) || names[NormalizeSearch(p.Name)] {
			return nil, ErrInvalid
		}
		ids[p.ID] = true
		provinceIDs[p.ID] = true
		names[NormalizeSearch(p.Name)] = true
		p.Cities = append([]commerce.City{}, p.Cities...)
		cityNames := map[string]bool{}
		for j := range p.Cities {
			c := &p.Cities[j]
			c.Name = strings.TrimSpace(c.Name)
			total++
			if c.ID == "" || len(c.ID) > 80 || ids[c.ID] || c.Name == "" || !textBound(c.Name, 100) || cityNames[NormalizeSearch(c.Name)] || total > 10000 {
				return nil, ErrInvalid
			}
			ids[c.ID] = true
			cityNames[NormalizeSearch(c.Name)] = true
			parents[c.ID] = p.ID
		}
	}
	// Stable records are deactivated, not deleted or silently moved between provinces.
	for _, p := range previous {
		if !provinceIDs[p.ID] {
			return nil, ErrInvalid
		}
		for _, c := range p.Cities {
			if parents[c.ID] != p.ID {
				return nil, ErrInvalid
			}
		}
	}
	return out, nil
}
