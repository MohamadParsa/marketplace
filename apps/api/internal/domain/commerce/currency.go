package commerce

import (
	"errors"
	"math/big"
	"regexp"
	"strings"
)

type Currency string

const (
	CurrencyToman Currency = "IRT"
	CurrencyRial  Currency = "IRR"
	SiteCurrency           = CurrencyToman
)

func (c Currency) Valid() bool { return c == CurrencyToman || c == CurrencyRial }

// DefaultGatewayRate expresses gateway units per one toman. Amount and ledger
// storage remain integer rial; a gateway setting never rewrites stored money.
func (c Currency) DefaultGatewayRate() string {
	if c == CurrencyRial {
		return "10"
	}
	return "1"
}

var ErrGatewayConversion = errors.New("invalid_gateway_conversion")
var gatewayRatePattern = regexp.MustCompile(`^[0-9]{1,7}(\.[0-9]{1,6})?$`)

// NormalizeGatewayRate validates exact decimal input, without float rounding.
func NormalizeGatewayRate(value string) (string, error) {
	if !gatewayRatePattern.MatchString(value) {
		return "", ErrGatewayConversion
	}
	rate, ok := new(big.Rat).SetString(value)
	if !ok || rate.Sign() <= 0 || rate.Cmp(big.NewRat(1000000, 1)) > 0 {
		return "", ErrGatewayConversion
	}
	return strings.TrimRight(strings.TrimRight(rate.FloatString(6), "0"), "."), nil
}

// GatewayAmount prepares an integer provider amount. Fractional provider units
// and overflow fail instead of silently changing the amount owed by a customer.
// The future adapter must snapshot this result, currency and rate per attempt.
func GatewayAmount(amount Amount, currency Currency, rate string) (int64, error) {
	if amount <= 0 || amount > MaxAmount || !currency.Valid() {
		return 0, ErrGatewayConversion
	}
	normalized, err := NormalizeGatewayRate(rate)
	if err != nil {
		return 0, err
	}
	factor, _ := new(big.Rat).SetString(normalized)
	converted := new(big.Rat).Mul(big.NewRat(int64(amount), 10), factor)
	if !converted.IsInt() || !converted.Num().IsInt64() || converted.Sign() <= 0 {
		return 0, ErrGatewayConversion
	}
	return converted.Num().Int64(), nil
}
