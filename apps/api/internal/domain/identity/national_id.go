package identity

import "strings"

// ValidNationalID checks format/check digit only, not the holder's legal identity.
func ValidNationalID(s string) bool {
	if len(s) != 10 {
		return false
	}
	same := true
	sum := 0
	for i := 0; i < 10; i++ {
		if s[i] < '0' || s[i] > '9' {
			return false
		}
		if s[i] != s[0] {
			same = false
		}
		if i < 9 {
			sum += int(s[i]-'0') * (10 - i)
		}
	}
	if same {
		return false
	}
	rem := sum % 11
	check := int(s[9] - '0')
	if rem < 2 {
		return check == rem
	}
	return check == 11-rem
}

func NormalizeNationalID(s string) string {
	return strings.Map(func(r rune) rune {
		if r >= '۰' && r <= '۹' {
			return '0' + r - '۰'
		}
		if r >= '٠' && r <= '٩' {
			return '0' + r - '٠'
		}
		return r
	}, strings.TrimSpace(s))
}
