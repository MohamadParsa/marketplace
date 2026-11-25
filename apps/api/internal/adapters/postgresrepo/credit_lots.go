package postgresrepo

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/cherriva/marketplace/apps/api/internal/domain/commerce"
	"github.com/jackc/pgx/v5"
)

func saveCreditLots(ctx context.Context, tx pgx.Tx, c *commerce.State) error {
	accounts := map[string]bool{}
	for _, a := range c.CreditAccounts {
		if accounts[a.UserID] {
			return fmt.Errorf("duplicate credit account")
		}
		accounts[a.UserID] = true
	}
	ledger := map[string]commerce.WalletEntry{}
	balance := map[string]commerce.Amount{}
	for _, e := range c.Entries {
		ledger[e.ID] = e
		balance[e.UserID] += e.Amount
	}
	totals := map[string]commerce.Amount{}
	entryAmounts := map[string]commerce.Amount{}
	seen := map[string]bool{}
	for _, lot := range c.CreditLots {
		if lot.ID == "" || seen[lot.ID] || !accounts[lot.UserID] || lot.Amount <= 0 || lot.Amount > commerce.MaxAmount || lot.Spent < 0 || lot.Spent > lot.Amount {
			return fmt.Errorf("invalid credit source")
		}
		seen[lot.ID] = true
		var err error
		totals[lot.UserID], err = commerce.Add(totals[lot.UserID], lot.Amount-lot.Spent)
		if err != nil {
			return err
		}
		entryID := lot.LedgerEntryID()
		if entryID != "" {
			e, ok := ledger[entryID]
			if !ok || e.UserID != lot.UserID || e.Amount <= 0 {
				return fmt.Errorf("credit source entry mismatch")
			}
			entryAmounts[entryID], err = commerce.Add(entryAmounts[entryID], lot.Amount)
			if err != nil {
				return err
			}
		}
		canonical := lot
		canonical.Spent = 0
		flag := lot.CanWithdraw()
		canonical.Withdrawable = &flag
		canonical.EntryID = entryID
		b, err := json.Marshal(canonical)
		if err != nil {
			return err
		}
		var entry, origin any
		if entryID != "" {
			entry = entryID
		}
		if lot.OriginLotID != "" {
			origin = lot.OriginLotID
		}
		res, err := tx.Exec(ctx, `INSERT INTO commerce_credit_lots(id,user_id,entry_id,origin_lot_id,amount,spent,withdrawable,fingerprint)
 VALUES($1,$2,$3,$4,$5,$6,$7,$8) ON CONFLICT(id) DO UPDATE SET spent=excluded.spent
 WHERE commerce_credit_lots.fingerprint=excluded.fingerprint AND commerce_credit_lots.spent<=excluded.spent`, lot.ID, lot.UserID, entry, origin, int64(lot.Amount), int64(lot.Spent), flag, digest(b))
		if err != nil {
			return err
		}
		if res.RowsAffected() != 1 {
			return fmt.Errorf("credit origin is immutable")
		}
	}
	for user := range accounts {
		if totals[user] != balance[user] {
			return fmt.Errorf("credit sources do not reconcile with ledger")
		}
	}
	for id, amount := range entryAmounts {
		if amount != ledger[id].Amount {
			return fmt.Errorf("credit entry allocation mismatch")
		}
	}
	var count int
	if err := tx.QueryRow(ctx, "SELECT count(*) FROM commerce_credit_lots").Scan(&count); err != nil {
		return err
	}
	if count != len(c.CreditLots) {
		return fmt.Errorf("credit sources cannot shrink")
	}
	return nil
}
