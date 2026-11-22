package postgresrepo

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/cherriva/marketplace/apps/api/internal/domain/commerce"
	"github.com/jackc/pgx/v5"
	"time"
)

func digest(data []byte) string { b := sha256.Sum256(data); return hex.EncodeToString(b[:]) }
func saveCommerce(ctx context.Context, tx pgx.Tx, c *commerce.State) error {
	var existing int
	if err := tx.QueryRow(ctx, "SELECT count(*) FROM commerce_wallet_entries").Scan(&existing); err != nil {
		return err
	}
	if c == nil {
		var deposits int
		if err := tx.QueryRow(ctx, "SELECT count(*) FROM commerce_deposits").Scan(&deposits); err != nil {
			return err
		}
		if existing > 0 || deposits > 0 {
			return fmt.Errorf("financial state cannot be removed")
		}
		return nil
	}
	var depositCount int
	if err := tx.QueryRow(ctx, "SELECT count(*) FROM commerce_deposits").Scan(&depositCount); err != nil {
		return err
	}
	if depositCount > len(c.Deposits) {
		return fmt.Errorf("deposit history cannot shrink")
	}
	if existing > len(c.Entries) {
		return fmt.Errorf("wallet ledger cannot shrink")
	}
	for _, e := range c.Entries {
		canonical, _ := json.Marshal(struct {
			ID, User, Source, Kind, At string
			Amount                     commerce.Amount
		}{e.ID, e.UserID, e.Source, e.Kind, e.At.UTC().Format(time.RFC3339Nano), e.Amount})
		finger := digest(canonical)
		res, err := tx.Exec(ctx, "INSERT INTO commerce_wallet_entries(id,user_id,source,amount,kind,created_at,fingerprint) VALUES($1,$2,$3,$4,$5,$6,$7) ON CONFLICT(id) DO NOTHING", e.ID, e.UserID, e.Source, int64(e.Amount), e.Kind, e.At, finger)
		if err != nil {
			return err
		}
		if res.RowsAffected() == 0 {
			var saved string
			if err = tx.QueryRow(ctx, "SELECT fingerprint FROM commerce_wallet_entries WHERE id=$1", e.ID).Scan(&saved); err != nil {
				return err
			}
			if saved != finger {
				return fmt.Errorf("wallet ledger is immutable")
			}
		}
	}
	var savedEntries int
	if err := tx.QueryRow(ctx, "SELECT count(*) FROM commerce_wallet_entries").Scan(&savedEntries); err != nil {
		return err
	}
	if savedEntries != len(c.Entries) {
		return fmt.Errorf("wallet ledger entries cannot be replaced")
	}
	for _, d := range c.Deposits {
		raw, _ := json.Marshal(d)
		var verified any
		if d.VerifiedKey != "" {
			verified = d.VerifiedKey
		}
		result, err := tx.Exec(ctx, `INSERT INTO commerce_deposits(id,user_id,reference_key,verified_key,status,amount,fingerprint) VALUES($1,$2,$3,$4,$5,$6,$7)
  ON CONFLICT(id) DO UPDATE SET verified_key=excluded.verified_key,status=excluded.status,fingerprint=excluded.fingerprint
  WHERE (commerce_deposits.status='pending' AND commerce_deposits.user_id=excluded.user_id AND commerce_deposits.amount=excluded.amount AND commerce_deposits.reference_key=excluded.reference_key) OR commerce_deposits.fingerprint=excluded.fingerprint`, d.ID, d.UserID, d.ReferenceKey, verified, d.Status, int64(d.Amount), digest(raw))
		if err != nil {
			return err
		}
		if result.RowsAffected() != 1 {
			return fmt.Errorf("reviewed deposit is immutable")
		}
	}
	var savedDeposits int
	if err := tx.QueryRow(ctx, "SELECT count(*) FROM commerce_deposits").Scan(&savedDeposits); err != nil {
		return err
	}
	if savedDeposits != len(c.Deposits) {
		return fmt.Errorf("deposit history cannot be replaced")
	}
	ids := []string{}
	for _, r := range c.Receipts {
		ids = append(ids, r.ID)
		if len(r.Data) > 0 {
			finger := digest(r.Data)
			result, err := tx.Exec(ctx, `INSERT INTO commerce_receipts(id,deposit_id,user_id,data,fingerprint) VALUES($1,$2,$3,$4,$5) ON CONFLICT(id) DO NOTHING`, r.ID, r.DepositID, r.UserID, r.Data, finger)
			if err != nil {
				return err
			}
			if result.RowsAffected() == 0 {
				var saved string
				if err = tx.QueryRow(ctx, "SELECT fingerprint FROM commerce_receipts WHERE id=$1 AND user_id=$2 AND deposit_id=$3", r.ID, r.UserID, r.DepositID).Scan(&saved); err != nil {
					return err
				}
				if saved != finger {
					return fmt.Errorf("receipt is immutable")
				}
			}
		}
	}
	_, err := tx.Exec(ctx, "DELETE FROM commerce_receipts WHERE NOT(id=ANY($1::text[]))", ids)
	return err
}
func (r *Repository) ReadReceiptData(ctx context.Context, id, user, deposit string) ([]byte, error) {
	var data []byte
	e := r.db.QueryRow(ctx, "SELECT data FROM commerce_receipts WHERE id=$1 AND user_id=$2 AND deposit_id=$3", id, user, deposit).Scan(&data)
	return data, e
}
