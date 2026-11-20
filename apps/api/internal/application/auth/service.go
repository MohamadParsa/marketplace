package auth

import (
	"context"
	"crypto/subtle"
	"github.com/cherriva/marketplace/apps/api/internal/domain/identity"
	"golang.org/x/crypto/bcrypt"
	"strings"
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
	requestID, _, err := s.issueOTP(ctx, identity.Challenge{Phone: phone, Surface: surface})
	return requestID, err
}

// PasswordAvailable reports whether the normalized phone has a credential.
// It is intended for the post-phone login choice; callers must not use it as
// an account enumeration API.
func (s *Service) PasswordAvailable(ctx context.Context, surface identity.Surface, phone string) (bool, error) {
	user, err := s.users.FindByPhone(ctx, surface, identity.NormalizePhone(phone))
	if err != nil {
		return false, err
	}
	if !user.Active {
		return false, identity.ErrInactive
	}
	return user.PasswordHash != "", nil
}

// RequestAdminOTP sends a challenge to the current administrator's own account.
func (s *Service) RequestAdminOTP(ctx context.Context, token string) (string, string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	session, err := s.Authenticate(ctx, token, identity.SurfaceBackoffice)
	if err != nil {
		return "", "", err
	}
	if session.User.Role != identity.RoleSuperAdmin {
		return "", "", identity.ErrNotAdmin
	}
	return s.issueOTP(ctx, identity.Challenge{Phone: session.User.Phone, Surface: identity.SurfaceBackoffice, ReauthSession: session.Token})
}

func (s *Service) issueOTP(ctx context.Context, challenge identity.Challenge) (string, string, error) {
	now := s.clock.Now()
	if last, ok, err := s.challenges.LastRequest(ctx, challenge.CooldownKey()); err != nil {
		return "", "", err
	} else if ok && now.Sub(last) < resendAfter {
		return "", "", identity.ErrResendSoon
	}
	recent, err := s.challenges.RecordRequest(ctx, challengeKey(challenge.Surface, challenge.Phone), now)
	if err != nil {
		return "", "", err
	}
	cutoff := now.Add(-requestWindow)
	count := 0
	for _, sent := range recent {
		if sent.After(cutoff) {
			count++
		}
	}
	if count > maxRequests {
		return "", "", identity.ErrRateLimit
	}
	code, err := s.random.Code()
	if err != nil {
		return "", "", err
	}
	requestID, err := s.random.Token()
	if err != nil {
		return "", "", err
	}
	challenge.Code, challenge.Expires, challenge.IssuedAt = code, now.Add(otpTTL), now
	if err := s.challenges.PutChallenge(ctx, requestID, challenge); err != nil {
		return "", "", err
	}
	if err := s.notifier.Send(ctx, challenge.Surface, challenge.Phone, code); err != nil {
		_ = s.challenges.DeleteChallenge(ctx, requestID)
		return "", "", err
	}
	return requestID, code, nil
}

func (s *Service) VerifyOTP(ctx context.Context, surface identity.Surface, requestID, code string) (identity.Session, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	challenge, err := s.consumeOTP(ctx, surface, "", requestID, code)
	if err != nil {
		return identity.Session{}, err
	}
	user, err := s.users.FindByPhone(ctx, surface, challenge.Phone)
	if err != nil {
		return identity.Session{}, err
	}
	if !user.Active {
		return identity.Session{}, identity.ErrInactive
	}
	return s.createSession(ctx, user)
}

// PasswordLogin authenticates a phone account without requiring a second OTP.
func (s *Service) PasswordLogin(ctx context.Context, surface identity.Surface, phone, password string) (identity.Session, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	identifier := strings.TrimSpace(phone)
	var user identity.User
	var err error
	if strings.Contains(identifier, "@") {
		user, err = s.users.FindByEmail(ctx, surface, identity.NormalizeEmail(identifier))
	} else {
		user, err = s.users.FindByPhone(ctx, surface, identity.NormalizePhone(identifier))
	}
	if err != nil || !user.Active || user.PasswordHash == "" || bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)) != nil {
		return identity.Session{}, identity.ErrInvalidCredentials
	}
	return s.createSession(ctx, user)
}

func (s *Service) UpdateProfile(ctx context.Context, surface identity.Surface, token, name, email, password string) (identity.User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	session, err := s.Authenticate(ctx, token, surface)
	if err != nil {
		return identity.User{}, err
	}
	name = strings.TrimSpace(name)
	email = identity.NormalizeEmail(email)
	if name == "" || (email != "" && !strings.Contains(email, "@")) {
		return identity.User{}, identity.ErrInvalidProfile
	}
	hash := ""
	if password != "" {
		if len(password) < 8 {
			return identity.User{}, identity.ErrInvalidProfile
		}
		encoded, e := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if e != nil {
			return identity.User{}, e
		}
		hash = string(encoded)
	}
	if err := s.users.UpdateProfile(ctx, surface, session.User.ID, name, email, hash); err != nil {
		return identity.User{}, err
	}
	return s.users.FindByPhone(ctx, surface, session.User.Phone)
}

func (s *Service) RequestRecoveryOTP(ctx context.Context, surface identity.Surface, phone, destination string) (string, string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	phone = identity.NormalizePhone(phone)
	user, err := s.users.FindByPhone(ctx, surface, phone)
	if err != nil || !user.Active {
		return "", "", identity.ErrInvalidCredentials
	}
	if destination != "" && destination != "mobile" {
		return "", "", identity.ErrEmailRecoveryUnavailable
	}
	return s.issueOTP(ctx, identity.Challenge{Phone: phone, Surface: surface, Purpose: "password-recovery"})
}

func (s *Service) RecoverPassword(ctx context.Context, surface identity.Surface, requestID, code, password string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if len(password) < 8 {
		return identity.ErrInvalidProfile
	}
	challenge, err := s.consumeOTP(ctx, surface, "", requestID, code)
	if err != nil || challenge.Purpose != "password-recovery" {
		return identity.ErrInvalidOTP
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	user, err := s.users.FindByPhone(ctx, surface, challenge.Phone)
	if err != nil {
		return err
	}
	return s.users.UpdateProfile(ctx, surface, user.ID, user.Name, user.Email, string(hash))
}

// AdminResetPassword lets an authenticated super-admin reset a user's password.
func (s *Service) AdminResetPassword(ctx context.Context, token string, surface identity.Surface, targetPhone, password string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	admin, err := s.Authenticate(ctx, token, identity.SurfaceBackoffice)
	if err != nil || admin.User.Role != identity.RoleSuperAdmin {
		return identity.ErrNotAdmin
	}
	if len(password) < 8 {
		return identity.ErrInvalidProfile
	}
	user, err := s.users.FindByPhone(ctx, surface, identity.NormalizePhone(targetPhone))
	if err != nil {
		return err
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	return s.users.UpdateProfile(ctx, surface, user.ID, user.Name, user.Email, string(hash))
}

func (s *Service) createSession(ctx context.Context, user identity.User) (identity.Session, error) {
	token, err := s.random.Token()
	if err != nil {
		return identity.Session{}, err
	}
	session := identity.Session{Token: token, User: user, Expires: s.clock.Now().Add(sessionTTL)}
	if err := s.sessions.Put(ctx, session); err != nil {
		return identity.Session{}, err
	}
	return session, nil
}

func (s *Service) ConfirmAdminOTP(ctx context.Context, token, requestID, code string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	session, err := s.Authenticate(ctx, token, identity.SurfaceBackoffice)
	if err != nil || session.User.Role != identity.RoleSuperAdmin {
		return identity.ErrInvalidOTP
	}
	_, err = s.consumeOTP(ctx, identity.SurfaceBackoffice, session.Token, requestID, code)
	return err
}

func (s *Service) consumeOTP(ctx context.Context, surface identity.Surface, reauthSession, requestID, code string) (identity.Challenge, error) {
	challenge, err := s.challenges.GetChallenge(ctx, requestID)
	if err != nil || challenge.Surface != surface || challenge.ReauthSession != reauthSession || !s.clock.Now().Before(challenge.Expires) {
		return identity.Challenge{}, identity.ErrInvalidOTP
	}
	if subtle.ConstantTimeCompare([]byte(challenge.Code), []byte(code)) != 1 {
		challenge.Attempts++
		if challenge.Attempts >= maxAttempts {
			_ = s.challenges.DeleteChallenge(ctx, requestID)
		} else {
			_ = s.challenges.PutChallenge(ctx, requestID, challenge)
		}
		return identity.Challenge{}, identity.ErrInvalidOTP
	}
	if err := s.challenges.DeleteChallenge(ctx, requestID); err != nil {
		return identity.Challenge{}, err
	}
	return challenge, nil
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
	if err != nil || !user.Active || user.ID != session.User.ID || user.AccessVersion != session.User.AccessVersion {
		return identity.Session{}, identity.ErrInactive
	}
	session.User = user
	return session, nil
}
func challengeKey(surface identity.Surface, phone string) string {
	return string(surface) + ":" + phone
}
