package backoffice

import (
	"strings"
	"testing"
)

func TestPublicReferenceRoundTripAndNamespaceIsolation(t *testing.T) {
	for _, kind := range []string{"v", "o", "m", "p", "c", "a", "b", "s"} {
		for _, id := range []string{"gsp-variant-12034-54a6054d1cc4", "native-uuid", "chr-v-bmF0aXZl", "شناسه", strings.Repeat("a", 128)} {
			token := PublicReference(kind, id)
			actual, err := ResolvePublicReference(kind, token)
			if err != nil || actual != id || strings.Contains(token, "gsp-") {
				t.Fatalf("%s %s: %q %v", kind, id, actual, err)
			}
			if token != PublicReference(kind, id) {
				t.Fatal("unstable token")
			}
			if kind != "m" {
				if _, err = ResolvePublicReference("m", token); err == nil {
					t.Fatal("cross-kind alias accepted")
				}
			}
		}
	}
	for _, token := range []string{"chr-v-", "chr-v-!", "chr-v-YQ==", PublicReference("v", "x/y"), PublicReference("v", strings.Repeat("a", 129)), PublicReference("v", "\x00x")} {
		if _, err := ResolvePublicReference("v", token); err == nil {
			t.Fatal("invalid reference accepted", token)
		}
	}
	if a, _ := ResolvePublicReference("v", "gsp-variant-1"); a != "gsp-variant-1" {
		t.Fatal("legacy link lost")
	}
	if PublicReference("v", "ab") == PublicReference("v", "a-b") || PublicReference("v", "same") == PublicReference("o", "same") {
		t.Fatal("alias collision")
	}
}
