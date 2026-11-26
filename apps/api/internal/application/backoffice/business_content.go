package backoffice

import (
	"context"
	"github.com/cherriva/marketplace/apps/api/internal/domain/commerce"
	"net/mail"
	"net/url"
	"sort"
	"strings"
	"time"
	"unicode"
)

func businessPages(b *commerce.Business) map[string]*string {
	return map[string]*string{"about": &b.About, "contact": &b.Contact, "shipping": &b.ShippingPolicy, "returns": &b.ReturnPolicy, "privacy": &b.Privacy, "terms": &b.Terms}
}
func validSocialURL(raw string) bool {
	if len(raw) > 2048 || strings.IndexFunc(raw, func(r rune) bool { return unicode.IsSpace(r) || unicode.IsControl(r) }) >= 0 {
		return false
	}
	u, e := url.Parse(raw)
	return e == nil && u.Scheme == "https" && u.Hostname() != "" && u.User == nil && u.Opaque == "" && strings.Contains(u.Hostname(), ".")
}
func normalizeBusiness(b commerce.Business) (commerce.Business, error) {
	b.Name = strings.TrimSpace(b.Name)
	b.SupportPhone = strings.TrimSpace(b.SupportPhone)
	b.SupportEmail = strings.TrimSpace(b.SupportEmail)
	if b.SupportEmail != "" {
		a, e := mail.ParseAddress(b.SupportEmail)
		if e != nil || a.Address != b.SupportEmail {
			return b, ErrInvalid
		}
	}
	for _, r := range b.SupportPhone {
		if !strings.ContainsRune("0123456789۰۱۲۳۴۵۶۷۸۹+ -()", r) {
			return b, ErrInvalid
		}
	}
	pages := businessPages(&b)
	for key, v := range pages {
		*v = strings.TrimSpace(*v)
		if !textBound(*v, 20000) || strings.ContainsRune(*v, '\x00') {
			return b, ErrInvalid
		}
		if b.PublishedPages != nil && b.PublishedPages[key] && *v == "" {
			return b, ErrInvalid
		}
	}
	for key := range b.PublishedPages {
		if pages[key] == nil {
			return b, ErrInvalid
		}
	}
	if len(b.SocialLinks) > 20 {
		return b, ErrInvalid
	}
	kinds := map[string]bool{"instagram": true, "telegram": true, "whatsapp": true, "linkedin": true, "x": true, "youtube": true, "aparat": true, "other": true}
	seen := map[string]bool{}
	for i := range b.SocialLinks {
		l := &b.SocialLinks[i]
		l.Title = strings.TrimSpace(l.Title)
		l.URL = strings.TrimSpace(l.URL)
		if l.ID == "" || len(l.ID) > 80 || seen[l.ID] || !textBound(l.Title, 80) || !kinds[l.Kind] || l.Position < 0 || l.Position > 1000000 || l.URL != "" && !validSocialURL(l.URL) || l.Active && (l.Title == "" || l.URL == "") {
			return b, ErrInvalid
		}
		seen[l.ID] = true
	}
	return b, nil
}
func publicBusiness(b commerce.Business) commerce.Business {
	for key, v := range businessPages(&b) {
		if b.PublishedPages != nil && !b.PublishedPages[key] {
			*v = ""
		}
	}
	b.PublishedPages = nil
	links := []commerce.SocialLink{}
	for _, l := range b.SocialLinks {
		if l.Active && l.Title != "" && validSocialURL(l.URL) {
			links = append(links, l)
		}
	}
	sort.Slice(links, func(i, j int) bool {
		if links[i].Position != links[j].Position {
			return links[i].Position < links[j].Position
		}
		return links[i].ID < links[j].ID
	})
	b.SocialLinks = links
	return b
}
func (s *Service) PublicBusinessContent(ctx context.Context) (commerce.Business, error) {
	return cachedPublic(ctx, s, "business", "", nil, 5*time.Minute, func(ctx context.Context) (commerce.Business, error) {
		if r, ok := s.repository.(interface {
			ReadPublicBusiness(context.Context) (commerce.Business, error)
		}); ok {
			return r.ReadPublicBusiness(ctx)
		}
		w, e := s.repository.Snapshot(ctx)
		if e != nil {
			return commerce.Business{}, e
		}
		if w.Commerce == nil {
			return commerce.Business{}, nil
		}
		return publicBusiness(w.Commerce.Settings.Business), nil
	})
}
