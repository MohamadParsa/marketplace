package identity

import "strings"

func NormalizeEmail(email string) string { return strings.ToLower(strings.TrimSpace(email)) }

func NormalizePhone(phone string) string {
	phone = strings.Map(func(r rune) rune {
		switch r {
		case '۰', '٠':
			return '0'
		case '۱', '١':
			return '1'
		case '۲', '٢':
			return '2'
		case '۳', '٣':
			return '3'
		case '۴', '٤':
			return '4'
		case '۵', '٥':
			return '5'
		case '۶', '٦':
			return '6'
		case '۷', '٧':
			return '7'
		case '۸', '٨':
			return '8'
		case '۹', '٩':
			return '9'
		default:
			return r
		}
	}, phone)
	phone = strings.Join(strings.Fields(phone), "")
	if strings.HasPrefix(phone, "+98") {
		return "0" + strings.TrimPrefix(phone, "+98")
	}
	return phone
}
