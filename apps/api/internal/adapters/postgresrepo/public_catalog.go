package postgresrepo

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"github.com/cherriva/marketplace/apps/api/internal/application/backoffice"
	"github.com/cherriva/marketplace/apps/api/internal/domain/catalog"
	"github.com/cherriva/marketplace/apps/api/internal/domain/content"
	"github.com/jackc/pgx/v5"
	"sort"
	"time"
)

func catalogHash(data []byte) string { sum := sha256.Sum256(data); return hex.EncodeToString(sum[:]) }

// Runs inside the same workspace transaction for every writer, including the importer.
func savePublicCatalog(ctx context.Context, tx pgx.Tx, state backoffice.Workspace) error {
	w := backoffice.CatalogReadWorkspace(state)
	parts := map[string]*backoffice.Workspace{}
	for _, p := range w.Products {
		parts[p.ID] = &backoffice.Workspace{Products: []catalog.Product{p}}
	}
	for _, v := range w.Variants {
		if part := parts[v.ProductID]; part != nil {
			part.Variants = append(part.Variants, v)
		}
	}
	for _, o := range w.Offers {
		if part := parts[o.ProductID]; part != nil {
			part.Offers = append(part.Offers, o)
		}
	}
	shared := w
	shared.Products = nil
	shared.Variants = nil
	shared.Offers = nil
	payload, e := json.Marshal(shared)
	if e != nil {
		return e
	}
	var expiry *time.Time
	if state.Commerce != nil {
		for _, q := range state.Commerce.Quotes {
			if q.Status == "reserved" && (expiry == nil || q.ReservedUntil.Before(*expiry)) {
				t := q.ReservedUntil
				expiry = &t
			}
		}
	}
	existing := map[string]string{}
	rows, e := tx.Query(ctx, "SELECT id,fingerprint FROM catalog_public_products")
	if e != nil {
		return e
	}
	for rows.Next() {
		var id, hash string
		if e = rows.Scan(&id, &hash); e != nil {
			rows.Close()
			return e
		}
		existing[id] = hash
	}
	e = rows.Err()
	rows.Close()
	if e != nil {
		return e
	}
	changed := false
	ids := make([]string, 0, len(parts))
	for id := range parts {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	batch := &pgx.Batch{}
	for _, id := range ids {
		part := parts[id]
		sort.Slice(part.Variants, func(i, j int) bool { return part.Variants[i].ID < part.Variants[j].ID })
		sort.Slice(part.Offers, func(i, j int) bool { return part.Offers[i].ID < part.Offers[j].ID })
		b, e := json.Marshal(struct {
			Products []catalog.Product `json:"products"`
			Variants []catalog.Variant `json:"variants"`
			Offers   []catalog.Offer   `json:"offers"`
		}{part.Products, append([]catalog.Variant{}, part.Variants...), append([]catalog.Offer{}, part.Offers...)})
		if e != nil {
			return e
		}
		hash := catalogHash(b)
		if existing[id] != hash {
			changed = true
			batch.Queue("INSERT INTO catalog_public_products(id,fingerprint,payload) VALUES($1,$2,$3::jsonb) ON CONFLICT(id) DO UPDATE SET revision=nextval('catalog_public_revision_seq'),fingerprint=excluded.fingerprint,payload=excluded.payload", id, hash, b)
		}
		delete(existing, id)
	}
	for id := range existing {
		changed = true
		batch.Queue("DELETE FROM catalog_public_products WHERE id=$1", id)
	}
	if e = tx.SendBatch(ctx, batch).Close(); e != nil {
		return e
	}
	var oldHash string
	e = tx.QueryRow(ctx, "SELECT dependency_hash FROM catalog_public_state WHERE singleton").Scan(&oldHash)
	if e != nil && e != pgx.ErrNoRows {
		return e
	}
	hash := catalogHash(payload)
	sharedChanged := oldHash != hash
	if !changed && !sharedChanged {
		_, e = tx.Exec(ctx, "UPDATE catalog_public_state SET next_expiry=$1 WHERE singleton AND next_expiry IS DISTINCT FROM $1", expiry)
		return e
	}
	fullPayload, e := json.Marshal(w)
	if e != nil {
		return e
	}
	_, e = tx.Exec(ctx, `INSERT INTO catalog_public_state(singleton,dependency_hash,payload,next_expiry,full_payload) VALUES(true,$1,$2::jsonb,$3,$5::jsonb)
 ON CONFLICT(singleton) DO UPDATE SET
 generation=catalog_public_state.generation+CASE WHEN $4 OR catalog_public_state.dependency_hash<>excluded.dependency_hash THEN 1 ELSE 0 END,
 dependency_revision=catalog_public_state.dependency_revision+CASE WHEN catalog_public_state.dependency_hash<>excluded.dependency_hash THEN 1 ELSE 0 END,
 dependency_hash=excluded.dependency_hash,payload=excluded.payload,next_expiry=excluded.next_expiry,full_payload=excluded.full_payload`, hash, payload, expiry, changed, fullPayload)
	if e != nil {
		return e
	}
	if !changed && !sharedChanged {
		return nil
	}
	publicIDs := backoffice.PublicCatalogMediaIDs(w)
	_, e = tx.Exec(ctx, "DELETE FROM catalog_public_media WHERE NOT(id=ANY($1::text[]))", publicIDs)
	if e != nil {
		return e
	}
	_, e = tx.Exec(ctx, "INSERT INTO catalog_public_media SELECT unnest($1::text[]) ON CONFLICT DO NOTHING", publicIDs)
	if e != nil {
		return e
	}
	publicBannerIDs := []string{}
	bannerBatch := &pgx.Batch{}
	for _, b := range w.Banners {
		if b.Published == nil || !b.Published.Config.Enabled {
			continue
		}
		for _, id := range []string{b.Published.Config.DesktopID, b.Published.Config.MobileID} {
			for _, a := range state.BannerAssets {
				if a.ID == id {
					publicBannerIDs = append(publicBannerIDs, id)
					bannerBatch.Queue("INSERT INTO catalog_public_banner_media VALUES($1,$2,$3,$4) ON CONFLICT(id) DO UPDATE SET data=excluded.data,width=excluded.width,height=excluded.height", a.ID, a.Data, a.Width, a.Height)
				}
			}
		}
	}
	bannerBatch.Queue("DELETE FROM catalog_public_banner_media WHERE NOT(id=ANY($1::text[]))", publicBannerIDs)
	return tx.SendBatch(ctx, bannerBatch).Close()
}

// Initial backfill only writes rebuildable read models and copies legacy inline media.
// It never replays financial saves or changes the original workspace/identity.
func (r *Repository) ensurePublicCatalog(ctx context.Context) error {
	var exists bool
	if e := r.db.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM catalog_public_state)").Scan(&exists); e != nil {
		return e
	}
	if exists {
		return nil
	}
	return pgx.BeginFunc(ctx, r.db, func(tx pgx.Tx) error {
		w, e := load(ctx, tx, true)
		if e != nil {
			return e
		}
		for _, m := range w.IconMedia {
			if len(m.Data) > 0 {
				if _, e = tx.Exec(ctx, "INSERT INTO catalog_media(id,data) VALUES($1,$2) ON CONFLICT DO NOTHING", m.ID, m.Data); e != nil {
					return e
				}
			}
		}
		return savePublicCatalog(ctx, tx, w)
	})
}
func (r *Repository) CatalogStamp(ctx context.Context, id string) (backoffice.CatalogStamp, error) {
	r.stampQueries.Add(1)
	var s backoffice.CatalogStamp
	e := r.db.QueryRow(ctx, `SELECT s.namespace,s.generation,s.dependency_revision,coalesce(p.revision,0),s.next_expiry
 FROM catalog_public_state s LEFT JOIN catalog_public_products p ON p.id=$1 WHERE s.singleton`, id).Scan(&s.Namespace, &s.Generation, &s.Dependencies, &s.Product, &s.NextExpiry)
	// Namespace is unique per persistent database/schema, unrelated to its credentials.
	return s, e
}
func (r *Repository) PublicCatalog(ctx context.Context, id string) (backoffice.Workspace, error) {
	r.projectionQueries.Add(1)
	queryStart := time.Now()
	var data []byte
	var e error
	if id != "" {
		e = r.db.QueryRow(ctx, `SELECT s.payload || coalesce(p.payload,'{}'::jsonb) FROM catalog_public_state s LEFT JOIN catalog_public_products p ON p.id=$1 WHERE s.singleton`, id).Scan(&data)
	} else {
		e = r.db.QueryRow(ctx, "SELECT full_payload FROM catalog_public_state WHERE singleton").Scan(&data)
	}
	r.projectionSQLMicros.Add(uint64(time.Since(queryStart).Microseconds()))
	var w backoffice.Workspace
	if e == nil {
		decodeStart := time.Now()
		e = json.Unmarshal(data, &w)
		r.decodeMicros.Add(uint64(time.Since(decodeStart).Microseconds()))
	}
	return w, e
}
func (r *Repository) CommerceExpiryDue(ctx context.Context, now time.Time) (bool, error) {
	var due bool
	e := r.db.QueryRow(ctx, "SELECT coalesce(next_expiry<=$1,false) FROM catalog_public_state WHERE singleton", now).Scan(&due)
	return due, e
}
func (r *Repository) ReadPublicCatalogMedia(ctx context.Context, id string) ([]byte, error) {
	r.mediaQueries.Add(1)
	var data []byte
	e := r.db.QueryRow(ctx, "SELECT m.data FROM catalog_media m JOIN catalog_public_media p USING(id) WHERE m.id=$1", id).Scan(&data)
	if e == pgx.ErrNoRows {
		return nil, backoffice.ErrNotFound
	}
	return data, e
}
func (r *Repository) ReadPublicBannerMedia(ctx context.Context, id string) (content.BannerAsset, error) {
	var a content.BannerAsset
	e := r.db.QueryRow(ctx, "SELECT id,data,width,height FROM catalog_public_banner_media WHERE id=$1", id).Scan(&a.ID, &a.Data, &a.Width, &a.Height)
	if e == pgx.ErrNoRows {
		e = backoffice.ErrNotFound
	}
	return a, e
}

// No cache: authentication always resolves current roles and credential versions.
func (r *Repository) IdentitySnapshot(ctx context.Context) (backoffice.Workspace, error) {
	var w backoffice.Workspace
	e := pgx.BeginTxFunc(ctx, r.db, pgx.TxOptions{IsoLevel: pgx.RepeatableRead, AccessMode: pgx.ReadOnly}, func(tx pgx.Tx) error { return loadIdentity(ctx, tx, &w) })
	return w, e
}

type CatalogReadStatistics struct {
	StampQueries, ProjectionQueries, ProjectionSQLMicros, DecodeMicros, MediaQueries, FullSnapshots uint64
}

func (r *Repository) CatalogReadStatistics() CatalogReadStatistics {
	return CatalogReadStatistics{r.stampQueries.Load(), r.projectionQueries.Load(), r.projectionSQLMicros.Load(), r.decodeMicros.Load(), r.mediaQueries.Load(), r.fullSnapshots.Load()}
}

func (r *Repository) ReadManagedCatalogMedia(ctx context.Context, id string) ([]byte, error) {
	var data []byte
	reference, _ := json.Marshal([]map[string]string{{"id": id}})
	e := r.db.QueryRow(ctx, `SELECT m.data FROM catalog_media m JOIN catalog_public_state s ON s.singleton WHERE m.id=$1 AND s.payload->'icon_media' @> $2::jsonb`, id, reference).Scan(&data)
	if e == pgx.ErrNoRows {
		e = backoffice.ErrNotFound
	}
	return data, e
}
