package commerce

import (
	"math/big"
	"sort"
	"time"
)

// Basis points and minor-unit integers keep all checkout arithmetic exact.
func Percent(amount Amount, bps int) (Amount, error) {
	if amount < 0 || amount > MaxAmount || bps < 0 || bps > 10000 {
		return 0, ErrAmount
	}
	n := new(big.Int).Mul(big.NewInt(int64(amount)), big.NewInt(int64(bps)))
	n.Add(n, big.NewInt(5000)).Quo(n, big.NewInt(10000))
	return Amount(n.Int64()), nil
}

type WeightedAmount struct {
	Key    string
	Amount Amount
}

// Allocate uses largest remainders; keys break ties independently of input order.
func Allocate(total Amount, weights []WeightedAmount) (map[string]Amount, error) {
	result := map[string]Amount{}
	var sum Amount
	for _, w := range weights {
		if w.Key == "" || w.Amount < 0 {
			return nil, ErrAmount
		}
		if _, exists := result[w.Key]; exists {
			return nil, ErrAmount
		}
		result[w.Key] = 0
		var err error
		sum, err = Add(sum, w.Amount)
		if err != nil {
			return nil, err
		}
	}
	if total < 0 || total > sum {
		return nil, ErrAmount
	}
	if total == 0 {
		return result, nil
	}
	type remainder struct {
		key   string
		value *big.Int
	}
	remainders := []remainder{}
	var assigned Amount
	for _, w := range weights {
		n := new(big.Int).Mul(big.NewInt(int64(total)), big.NewInt(int64(w.Amount)))
		q, r := new(big.Int), new(big.Int)
		q.QuoRem(n, big.NewInt(int64(sum)), r)
		result[w.Key] = Amount(q.Int64())
		assigned += result[w.Key]
		remainders = append(remainders, remainder{w.Key, r})
	}
	sort.Slice(remainders, func(i, j int) bool {
		c := remainders[i].value.Cmp(remainders[j].value)
		if c == 0 {
			return remainders[i].key < remainders[j].key
		}
		return c > 0
	})
	for i := Amount(0); i < total-assigned; i++ {
		result[remainders[int(i)].key]++
	}
	return result, nil
}

type Coupon struct {
	ID            string     `json:"id"`
	Code          string     `json:"code"`
	Name          string     `json:"name"`
	Version       int        `json:"version"`
	Active        bool       `json:"active"`
	Kind          string     `json:"kind"`
	Value         Amount     `json:"value"` // fixed: minor units; percent: basis points
	Target        string     `json:"target"`
	AllCategories bool       `json:"all_categories"`
	Categories    []string   `json:"categories"`
	Min           Amount     `json:"min"`
	Max           Amount     `json:"max"`
	Cap           Amount     `json:"cap"`
	StartsAt      *time.Time `json:"starts_at,omitempty"`
	EndsAt        *time.Time `json:"ends_at,omitempty"`
	GlobalLimit   int        `json:"global_limit"` // 0 explicitly means unlimited
	UserLimit     int        `json:"user_limit"`
	PlatformBPS   int        `json:"platform_bps"`
}

// Unit allocations are immutable purchase snapshots, suitable for later partial returns.
type DiscountAllocation struct {
	Key      string `json:"key"`
	StoreID  string `json:"store_id"`
	Amount   Amount `json:"amount"`
	Platform Amount `json:"platform"`
	Seller   Amount `json:"seller"`
}
type DiscountUsage struct {
	Coupon      Coupon               `json:"coupon"`
	Amount      Amount               `json:"amount"`
	Allocations []DiscountAllocation `json:"allocations"`
}

type CreditLot struct {
	ID     string    `json:"id"`
	UserID string    `json:"user_id"`
	Source string    `json:"source"`
	Amount Amount    `json:"amount"`
	Spent  Amount    `json:"spent"`
	At     time.Time `json:"at"`
}
type CreditAccount struct {
	UserID string    `json:"user_id"`
	Anchor string    `json:"anchor"`
	At     time.Time `json:"at"`
}
type CreditAllocation struct {
	LotID  string `json:"lot_id"`
	Amount Amount `json:"amount"`
}
type CreditUsage struct {
	Requested Amount             `json:"requested"`
	Existing  []CreditAllocation `json:"existing"`
	Funding   []CreditAllocation `json:"funding"`
	Status    string             `json:"status"` // held / posted / released; no second financial debit
}
