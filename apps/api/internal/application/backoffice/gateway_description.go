package backoffice

import "regexp"

var gatewayOrderNumber = regexp.MustCompile(`^[1-9][0-9]{0,18}$`)

// The eventual gateway adapter gets a description from the operation context,
// never from gateway configuration or arbitrary customer-entered text.
func GatewayTransactionDescription(purpose, purchaseNumber string) (string, error) {
	switch purpose {
	case "wallet-topup":
		if purchaseNumber != "" {
			return "", ErrInvalid
		}
		return "شارژ کیف پول چریوا", nil
	case "order-payment":
		if !gatewayOrderNumber.MatchString(purchaseNumber) {
			return "", ErrInvalid
		}
		return "پرداخت سفارش شماره " + purchaseNumber, nil
	default:
		return "", ErrInvalid
	}
}
