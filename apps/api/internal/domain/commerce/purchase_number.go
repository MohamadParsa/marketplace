package commerce

import "sort"

// EnsurePurchaseNumbers backfills pre-number orders deterministically. The whole
// aggregate, including this counter, is persisted under the repository write lock.
// Existing numbers are never reassigned; cancelled orders keep their number.
func (c *State) EnsurePurchaseNumbers() {
	if c.NextPurchaseNumber < 1000 {
		c.NextPurchaseNumber = 1000
	}
	pending := []int{}
	for i, p := range c.Purchases {
		if p.Number > 0 {
			if p.Number >= c.NextPurchaseNumber {
				c.NextPurchaseNumber = p.Number + 1
			}
		} else {
			pending = append(pending, i)
		}
	}
	sort.Slice(pending, func(i, j int) bool {
		a, b := c.Purchases[pending[i]], c.Purchases[pending[j]]
		if a.CreatedAt.Equal(b.CreatedAt) {
			return a.ID < b.ID
		}
		return a.CreatedAt.Before(b.CreatedAt)
	})
	for _, i := range pending {
		c.Purchases[i].Number = c.NextPurchaseNumber
		c.NextPurchaseNumber++
	}
}
