package commerce

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

var ErrAmount = errors.New("invalid toman amount")

func ParseToman(s string) (Amount, error) {
	if len(s) > 18 || s == "" || strings.TrimSpace(s) != s {
		return 0, ErrAmount
	}
	parts := strings.Split(s, ".")
	if len(parts) > 2 || parts[0] == "" {
		return 0, ErrAmount
	}
	for _, r := range parts[0] {
		if r < '0' || r > '9' {
			return 0, ErrAmount
		}
	}
	whole, e := strconv.ParseInt(parts[0], 10, 64)
	if e != nil || whole > int64(MaxAmount)/10 {
		return 0, ErrAmount
	}
	fraction := int64(0)
	if len(parts) == 2 {
		if len(parts[1]) != 1 || parts[1][0] < '0' || parts[1][0] > '9' {
			return 0, ErrAmount
		}
		fraction = int64(parts[1][0] - '0')
	}
	n := Amount(whole*10 + fraction)
	if n > MaxAmount {
		return 0, ErrAmount
	}
	return n, nil
}
func Toman(n Amount) string {
	negative := ""
	if n < 0 {
		negative = "-"
		n = -n
	}
	if n%10 == 0 {
		return negative + strconv.FormatInt(int64(n/10), 10)
	}
	return fmt.Sprintf("%s%d.%d", negative, n/10, n%10)
}
func Add(a, b Amount) (Amount, error) {
	if a < 0 || b < 0 || a > MaxAmount-b {
		return 0, ErrAmount
	}
	return a + b, nil
}
func Multiply(a Amount, n int) (Amount, error) {
	if a < 0 || n < 0 || n > 99 || n > 0 && a > MaxAmount/Amount(n) {
		return 0, ErrAmount
	}
	return a * Amount(n), nil
}
func DefaultSettings() Settings {
	return Settings{Version: 1, ClosedWeekdays: []int{4, 5}, Holidays: []string{}, Shipping: []ShippingMethod{}, Methods: []PaymentMethod{
		{ID: "wallet", Name: "کیف پول", Kind: "wallet", Active: true, AllCategories: true},
		{ID: "card-transfer", Name: "کارت‌به‌کارت", Kind: "card-transfer", Max: 150000000, AllCategories: true},
		{ID: "local", Name: "پرداخت محلی", Kind: "local", Max: 300000000, AllCategories: true},
	}}
}
