package auth

import (
	"context"
	"github.com/cherriva/marketplace/apps/api/internal/domain/identity"
	"time"
)

type Clock interface{ Now() time.Time }
type Random interface {
	Token() (string, error)
	Code() (string, error)
}
type Users interface {
	FindByPhone(context.Context, identity.Surface, string) (identity.User, error)
	FindByEmail(context.Context, identity.Surface, string) (identity.User, error)
	UpdateProfile(context.Context, identity.Surface, string, string, string, string) error
}
type Challenges interface {
	PutChallenge(context.Context, string, identity.Challenge) error
	GetChallenge(context.Context, string) (identity.Challenge, error)
	DeleteChallenge(context.Context, string) error
	RecordRequest(context.Context, string, time.Time) ([]time.Time, error)
	LastRequest(context.Context, string) (time.Time, bool, error)
}
type Sessions interface {
	Put(context.Context, identity.Session) error
	Get(context.Context, string) (identity.Session, error)
	Delete(context.Context, string) error
}
type OTPNotifier interface {
	Send(context.Context, identity.Surface, string, string) error
}
