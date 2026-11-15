package authmemory

import (
	"context"
	"github.com/cherriva/marketplace/apps/api/internal/domain/identity"
	"sync"
)

type Notifier struct {
	mu    sync.Mutex
	codes map[string]string
}

func NewNotifier() *Notifier { return &Notifier{codes: map[string]string{}} }
func (n *Notifier) Send(_ context.Context, surface identity.Surface, phone, code string) error {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.codes[string(surface)+":"+phone] = code
	return nil
}
func (n *Notifier) Code(surface identity.Surface, phone string) (string, bool) {
	n.mu.Lock()
	defer n.mu.Unlock()
	v, ok := n.codes[string(surface)+":"+phone]
	return v, ok
}
