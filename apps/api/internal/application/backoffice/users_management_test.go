package backoffice_test

import (
	"context"
	"github.com/cherriva/marketplace/apps/api/internal/application/backoffice"
	"sync"
	"testing"
)

func TestUserManagementConcurrentVersionUpdates(t *testing.T) {
	h := newCommerceHarness(t)
	ctx := context.Background()
	in := backoffice.UserInput{Phone: "09120000998", Active: true, FirstName: "هم‌زمان"}
	row, e := h.s.SaveUser(ctx, h.admin, in, "concurrent-create")
	if e != nil {
		t.Fatal(e)
	}
	in.ID = row.ID
	in.Version = row.Version
	var wg sync.WaitGroup
	result := make(chan error, 2)
	for _, key := range []string{"edit-a", "edit-b"} {
		wg.Add(1)
		go func(k string) { defer wg.Done(); _, e := h.s.SaveUser(ctx, h.admin, in, k); result <- e }(key)
	}
	wg.Wait()
	close(result)
	success, conflict := 0, 0
	for e := range result {
		if e == nil {
			success++
		} else if e == backoffice.ErrConflict {
			conflict++
		} else {
			t.Fatal(e)
		}
	}
	if success != 1 || conflict != 1 {
		t.Fatal("expected one committed write")
	}
}
