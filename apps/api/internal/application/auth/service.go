package auth

import (
	"context"
	"github.com/cherriva/marketplace/apps/api/internal/domain/identity"
	"sync"
	"time"
)

const (
	otpTTL        = 90 * time.Second
	resendAfter   = 60 * time.Second
	requestWindow = time.Hour
	maxRequests   = 5
	maxAttempts   = 5
	sessionTTL    = 8 * time.Hour
)

type Service struct {
	mu         sync.Mutex
	users      Users
	challenges Challenges
	sessions   Sessions
	notifier   OTPNotifier
	clock      Clock
	random     Random
}

func NewService(users Users, challenges Challenges, sessions Sessions, notifier OTPNotifier, clock Clock, random Random) *Service {
	return &Service{users: users, challenges: challenges, sessions: sessions, notifier: notifier, clock: clock, random: random}
}

func (s *Service) RequestOTP(ctx context.Context, surface identity.Surface, phone string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	phone = identity.NormalizePhone(phone)
	user, err := s.users.FindByPhone(ctx, surface, phone)
	if err != nil {
		return "", err
	}
	if !user.Active {
		return "", identity.ErrInactive
	}
	key := challengeKey(surface, phone)
	now := s.clock.Now()
	if last, ok, err := s.challenges.LastRequest(ctx, key); err != nil {
		return "", err
	} else if ok && now.Sub(last) < resendAfter {
		return "", identity.ErrResendSoon
	}
	recent, err := s.challenges.RecordRequest(ctx, key, now)
	if err != nil {
		return "", err
	}
	cutoff := now.Add(-requestWindow)
	count := 0
	for _, sent := range recent {
		if sent.After(cutoff) {
			count++
		}
	}
	if count > maxRequests {
		return "", identity.ErrRateLimit
	}
	code, err := s.random.Code()
	if err != nil {
		return "", err
	}
	requestID, err := s.random.Token()
	if err != nil {
		return "", err
	}
	if err := s.challenges.PutChallenge(ctx, requestID, identity.Challenge{Phone: phone, Surface: surface, Code: code, Expires: now.Add(otpTTL), IssuedAt: now}); err != nil {
		return "", err
	}
	if err := s.notifier.Send(ctx, surface, phone, code); err != nil {
		_ = s.challenges.DeleteChallenge(ctx, requestID)
		return "", err
	}
	return requestID, nil
}

func (s *Service) VerifyOTP(ctx context.Context, surface identity.Surface, requestID, code string) (identity.Session, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	challenge, err := s.challenges.GetChallenge(ctx, requestID)
	if err != nil {
		return identity.Session{}, identity.ErrInvalidOTP
	}
	now := s.clock.Now()
	if challenge.Surface != surface || !now.Before(challenge.Expires) {
		return identity.Session{}, identity.ErrInvalidOTP
	}
	if challenge.Code != code {
		challenge.Attempts++
		if challenge.Attempts >= maxAttempts {
			_ = s.challenges.DeleteChallenge(ctx, requestID)
		} else {
			_ = s.challenges.PutChallenge(ctx, requestID, challenge)
		}
		return identity.Session{}, identity.ErrInvalidOTP
	}
	user, err := s.users.FindByPhone(ctx, surface, challenge.Phone)
	if err != nil {
		return identity.Session{}, err
	}
	if !user.Active {
		return identity.Session{}, identity.ErrInactive
	}
	_ = s.challenges.DeleteChallenge(ctx, requestID)
	token, err := s.random.Token()
	if err != nil {
		return identity.Session{}, err
	}
	session := identity.Session{Token: token, User: user, Expires: now.Add(sessionTTL)}
	if err := s.sessions.Put(ctx, session); err != nil {
		return identity.Session{}, err
	}
	return session, nil
}

func (s *Service) Logout(ctx context.Context, token string) error {
	return s.sessions.Delete(ctx, token)
}
func (s *Service) Authenticate(ctx context.Context, token string, surface identity.Surface) (identity.Session, error) {
	session, err := s.sessions.Get(ctx, token)
	if err != nil || session.User.Surface != surface || !s.clock.Now().Before(session.Expires) {
		return identity.Session{}, identity.ErrInvalidOTP
	}
	user, err := s.users.FindByPhone(ctx, surface, session.User.Phone)
	if err != nil || !user.Active {
		return identity.Session{}, identity.ErrInactive
	}
	session.User = user
	return session, nil
}
func challengeKey(surface identity.Surface, phone string) string {
	return string(surface) + ":" + phone
}
