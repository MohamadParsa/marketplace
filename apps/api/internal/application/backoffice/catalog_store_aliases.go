package backoffice

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"sort"
	"strings"
	"time"
)

// AnonymousCatalogStoreName is stable across imports and never includes a source business name.
func AnonymousCatalogStoreName(id string) string {
	suffix := strings.TrimPrefix(id, "gsp-store-")
	numeric := suffix != "" && suffix != id
	for _, r := range suffix {
		if r < '0' || r > '9' {
			numeric = false
		}
	}
	if !numeric {
		sum := sha256.Sum256([]byte(id))
		suffix = hex.EncodeToString(sum[:4])
	}
	suffix = strings.Map(func(r rune) rune {
		if r >= '0' && r <= '9' {
			return '۰' + r - '0'
		}
		return r
	}, suffix)
	return "فروشگاه چریوا " + suffix
}

// AnonymizeImportedStores only changes names of the explicitly supplied imported stores.
// It leaves IDs, ownership, stock, orders, money and source import fingerprints untouched.
func AnonymizeImportedStores(ctx context.Context, repo Repository, ids []string, now time.Time) (int, error) {
	ids = append([]string{}, ids...)
	sort.Strings(ids)
	for i, id := range ids {
		if !strings.HasPrefix(id, "gsp-store-") || (i > 0 && ids[i-1] == id) {
			return 0, ErrInvalid
		}
	}
	if len(ids) == 0 {
		return 0, nil
	}
	encoded, _ := json.Marshal(ids)
	sum := sha256.Sum256(encoded)
	key := "catalog-store-aliases-" + hex.EncodeToString(sum[:])
	unchanged := errors.New("store_names_already_anonymous")
	count := 0
	e := repo.Update(ctx, func(w *Workspace) error {
		byID := map[string]int{}
		for i, st := range w.Stores {
			byID[st.ID] = i
		}
		for _, id := range ids {
			index, ok := byID[id]
			if !ok {
				return ErrNotFound
			}
			name := AnonymousCatalogStoreName(id)
			if w.Stores[index].Name != name {
				w.Stores[index].Name = name
				count++
			}
		}
		if count == 0 {
			return unchanged
		}
		w.Audit = append(w.Audit, Audit{ID: key, Actor: "catalog-maintenance-cli", Action: "catalog.store-names.anonymize", Target: "imported-test-catalog", Reason: "owner requested fictional seller names; store IDs and financial history retained", At: now})
		return nil
	})
	if errors.Is(e, unchanged) {
		return 0, nil
	}
	if e != nil {
		return 0, e
	}
	return count, nil
}
