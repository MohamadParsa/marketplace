package backoffice

import (
	"encoding/base64"
	"strings"
	"unicode/utf8"
)

// PublicReference is an injective, version-stable alias. It is not an access token.
// Encoding the whole ID (rather than a source suffix/hash) needs no mutable registry.
func PublicReference(kind, id string) string {
	if id == "" {
		return ""
	}
	return "chr-" + kind + "-" + base64.RawURLEncoding.EncodeToString([]byte(id))
}

// ResolvePublicReference accepts legacy IDs too. Existing authorization and
// product/variant/offer membership checks must still run after decoding.
func ResolvePublicReference(kind, token string) (string, error) {
	prefix := "chr-" + kind + "-"
	if !strings.HasPrefix(token, prefix) {
		for _, k := range []string{"v", "o", "m", "p", "c", "a", "b", "s"} {
			if strings.HasPrefix(token, "chr-"+k+"-") {
				return "", ErrInvalid
			}
		}
		return token, nil
	}
	b, err := base64.RawURLEncoding.DecodeString(strings.TrimPrefix(token, prefix))
	if err != nil || len(b) == 0 || len(b) > 128 || !utf8.Valid(b) || strings.TrimSpace(string(b)) != string(b) || strings.ContainsAny(string(b), "\x00/\\") || PublicReference(kind, string(b)) != token {
		return "", ErrInvalid
	}
	return string(b), nil
}
