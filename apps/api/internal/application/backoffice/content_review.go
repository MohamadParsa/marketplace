package backoffice

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"strings"
	"time"

	"github.com/cherriva/marketplace/apps/api/internal/domain/catalog"
	"github.com/cherriva/marketplace/apps/api/internal/domain/identity"
)

const StoreContentWrite identity.StorePermission = "products.content.write"

func hasContentStoreAccess(w Workspace, a identity.User, storeID string) bool {
	return a.Surface == identity.SurfaceSeller && w.Access != nil && identity.AllowsStore(w.StoreMemberships, w.Access.Roles, w.Access.Grants, a.ID, storeID, StoreContentWrite)
}
func sellerCanEditContent(w Workspace, a identity.User, p catalog.Product) bool {
	return p.Model == "independent" && p.OwnerStoreID != "" && hasContentStoreAccess(w, a, p.OwnerStoreID)
}
func authorizeContentSubmission(w Workspace, a identity.User, id string) error {
	if a.Surface != identity.SurfaceSeller {
		return ErrForbidden
	}
	a, e := catalogActor(w, a)
	if e != nil {
		return e
	}
	p, e := findProduct(&w, id)
	if e != nil {
		return e
	}
	if !sellerCanEditContent(w, a, *p) {
		return ErrForbidden
	}
	for _, st := range w.Stores {
		if st.ID == p.OwnerStoreID && st.Active {
			return nil
		}
	}
	return ErrForbidden
}

// Unknown legacy policy is fail-closed: edits need review, public approved content is unaffected.
func setStoreContentPolicy(w *Workspace, c Command) error {
	if c.Value != "review" && c.Value != "direct" {
		return ErrInvalid
	}
	for i := range w.Stores {
		st := &w.Stores[i]
		if st.ID != c.Target {
			continue
		}
		if st.ContentPolicyVersion != c.Version {
			return ErrConflict
		}
		st.ContentPolicy = c.Value
		st.ContentPolicyVersion++
		return nil
	}
	return ErrNotFound
}
func approvedContentFingerprint(p catalog.Product) string {
	// Public addressing metadata is independent of seller content approval.
	p.PublicCode = ""
	p.Active = nil
	p.ContentRevision = nil
	p.Version = 0
	p.VariantVersion = 0
	p.CreatedAt = time.Time{}
	b, _ := json.Marshal(p)
	h := sha256.Sum256(b)
	return hex.EncodeToString(h[:])
}
func contentCandidate(p catalog.Product, r catalog.ProductContentRevision) catalog.Product {
	p.Name = r.Name
	p.EnglishName = r.EnglishName
	p.Description = r.Description
	p.BrandID = r.BrandID
	p.Content = r.Content
	p.Attributes = append([]catalog.AttributeValue{}, r.Attributes...)
	return p
}
func submitProductContent(w *Workspace, c Command, id, actor string, now time.Time) error {
	p, e := findProduct(w, c.Target)
	if e != nil {
		return e
	}
	if p.Version != c.Version {
		return ErrConflict
	}
	in := c.Catalog
	if in == nil || in.Active != nil || !validCatalogText(*in) || in.Published || in.CategoryID != p.CategoryID || in.Model != "" && in.Model != p.Model || in.OwnerStoreID != "" && in.OwnerStoreID != p.OwnerStoreID || !validBrand(*w, in.BrandID) {
		return ErrInvalid
	}
	if in.Content != nil && !validProductContent(*in.Content) {
		return ErrInvalid
	}
	attrs := p.Attributes
	if in.Attributes != nil {
		if !validContentAttributes(*w, p.CategoryID, p.Attributes, *in.Attributes) {
			return ErrInvalid
		}
		attrs = *in.Attributes
	}
	content := p.Content
	if in.Content != nil {
		copy := *in.Content
		copy.FullDescription = sanitizeProductHTML(copy.FullDescription)
		content = &copy
	}
	r := catalog.ProductContentRevision{ID: id, State: "pending", BaseFingerprint: approvedContentFingerprint(*p), SubmittedBy: actor, SubmittedAt: now, Name: strings.TrimSpace(in.Name), EnglishName: strings.TrimSpace(in.EnglishName), Description: in.Description, BrandID: in.BrandID, Content: content, Attributes: append([]catalog.AttributeValue{}, attrs...)}
	direct := false
	for _, st := range w.Stores {
		if st.ID == p.OwnerStoreID {
			direct = st.ContentPolicy == "direct"
		}
	}
	if direct {
		candidate := contentCandidate(*p, r)
		if p.Status != "published" {
			if !canPublishProduct(*w, candidate) {
				return ErrInvalid
			}
			candidate.Status = "published"
		}
		r.State = "applied-direct"
		r.ReviewedAt = &now
		r.ReviewedBy = actor
		*p = candidate
	} else if p.Status != "published" {
		p.Status = "pending"
		r.BaseFingerprint = approvedContentFingerprint(*p)
	}
	p.ContentRevision = &r
	p.Version++
	return nil
}
func reviewProductContent(w *Workspace, c Command, actor string, now time.Time) error {
	if !textBound(c.Reason, 1000) {
		return ErrInvalid
	}
	p, e := findProduct(w, c.Target)
	if e != nil {
		return e
	}
	if p.Version != c.Version || p.ContentRevision == nil || p.ContentRevision.State != "pending" {
		return ErrConflict
	}
	r := p.ContentRevision
	if c.Action == "product.content.reject" {
		if strings.TrimSpace(c.Reason) == "" || !textBound(c.Reason, 1000) {
			return ErrInvalid
		}
		r.State = "rejected"
		r.Reason = c.Reason
		if p.Status != "published" {
			p.Status = "rejected"
		}
	} else {
		if approvedContentFingerprint(*p) != r.BaseFingerprint {
			return ErrConflict
		}
		candidate := contentCandidate(*p, *r)
		if !validBrand(*w, candidate.BrandID) || !validProductAttributes(*w, candidate.CategoryID, candidate.Attributes) {
			return ErrInvalid
		}
		if p.Status != "published" && !canPublishProduct(*w, candidate) {
			return ErrInvalid
		}
		candidate.Status = "published"
		*p = candidate
		r.State = "approved"
		r.Reason = strings.TrimSpace(c.Reason)
	}
	r.ReviewedBy = actor
	r.ReviewedAt = &now
	p.ContentRevision = r
	p.Version++
	return nil
}
