package identity

import "strings"

func NormalizePhone(phone string) string {
	phone = strings.Join(strings.Fields(phone), "")
	if strings.HasPrefix(phone, "+98") {
		return "0" + strings.TrimPrefix(phone, "+98")
	}
	return phone
}
