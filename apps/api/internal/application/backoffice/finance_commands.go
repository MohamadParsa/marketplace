package backoffice

import "github.com/cherriva/marketplace/apps/api/internal/domain/finance"

func applyTransfer(state *Workspace, c Command, id string) error {
	if !finance.ValidOutcome(c.Value) || c.Reason == "" {
		return ErrInvalid
	}
	for i := range state.Transfers {
		transfer := &state.Transfers[i]
		if transfer.ID != c.Target {
			continue
		}
		if transfer.Version != c.Version || (transfer.Status != "pending" && transfer.Status != "unknown") {
			return ErrConflict
		}
		if transfer.Kind == "debt" {
			return ErrInvalid
		}
		transfer.Status = c.Value
		transfer.Reserved = c.Value == "unknown"
		transfer.Version++
		transfer.Reference = "TEST-" + id
		if c.Value == "success" {
			state.Ledger = append(state.Ledger, finance.Entry{ID: id, TransferID: transfer.ID, Amount: transfer.Amount, Kind: transfer.Kind, Source: transfer.Source})
		}
		return nil
	}
	return ErrNotFound
}
