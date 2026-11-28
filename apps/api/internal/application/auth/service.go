package auth

import (
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"errors"
	"github.com/cherriva/marketplace/apps/api/internal/domain/identity"
	"golang.org/x/crypto/bcrypt"
	"net/mail"
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
	passwordFailures map[string][]time.Time
	mu               sync.Mutex
	users            Users
	challenges       Challenges
	sessions         Sessions
	notifier         OTPNotifier
	clock            Clock
	random           Random
}

func NewService(users Users, challenges Challenges, sessions Sessions, notifier OTPNotifier, clock Clock, random Random) *Service {
	return &Service{passwordFailures: map[string][]time.Time{}, users: users, challenges: challenges, sessions: sessions, notifier: notifier, clock: clock, random: random}
}

func (s *Service) RequestOTP(ctx context.Context, surface identity.Surface, phone string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	phone = identity.NormalizePhone(phone)
	if len(phone) != 11 || !strings.HasPrefix(phone, "09") || strings.IndexFunc(phone, func(r rune) bool { return r < '0' || r > '9' }) >= 0 {
		return "", identity.ErrInvalidCredentials
	}
	user, err := s.users.FindByPhone(ctx, surface, phone)
	_, registration := s.users.(CustomerRegistration)
	newCustomer := errors.Is(err, identity.ErrUnknown) && surface == identity.SurfaceCustomer && registration
	if err != nil && !newCustomer {
		return "", err
	}
	if !newCustomer && !user.Active {
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
	if errors.Is(err, identity.ErrUnknown) && surface == identity.SurfaceCustomer {
		if registrar, ok := s.users.(CustomerRegistration); ok {
			id, e := s.random.Token()
			if e != nil {
				return identity.Session{}, e
			}
			user, err = registrar.RegisterVerifiedCustomer(ctx, challenge.Phone, id)
		}
	}
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
	now := s.clock.Now()
	keySource := strings.ToLower(identifier)
	if err == nil {
		keySource = user.ID
	}
	digest := sha256.Sum256([]byte(keySource))
	key := hex.EncodeToString(digest[:])
	if s.passwordFailures == nil {
		s.passwordFailures = map[string][]time.Time{}
	}
	for k, times := range s.passwordFailures {
		kept := times[:0]
		for _, at := range times {
			if now.Sub(at) < requestWindow {
				kept = append(kept, at)
			}
		}
		if len(kept) == 0 {
			delete(s.passwordFailures, k)
		} else {
			s.passwordFailures[k] = kept
		}
	}
	if len(s.passwordFailures[key]) >= maxAttempts || len(s.passwordFailures) >= 10000 {
		return identity.Session{}, identity.ErrRateLimit
	}
	if len(identifier) > 254 || len(password) > 72 || err != nil || !user.Active || user.PasswordHash == "" || bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)) != nil {
		s.passwordFailures[key] = append(s.passwordFailures[key], now)
		return identity.Session{}, identity.ErrInvalidCredentials
	}
	delete(s.passwordFailures, key)
	return s.createSession(ctx, user)
}

func (s *Service) UpdateProfile(ctx context.Context, surface identity.Surface, token, name, email, password string, nationalID ...string) (identity.User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	session, err := s.Authenticate(ctx, token, surface)
	if err != nil {
		return identity.User{}, err
	}
	name = strings.TrimSpace(name)
	email = identity.NormalizeEmail(email)
	if name == "" || len(name) > 640 || len(email) > 254 {
		return identity.User{}, identity.ErrInvalidProfile
	}
	if email != "" {
		address, err := mail.ParseAddress(email)
		if err != nil || address.Address != email {
			return identity.User{}, identity.ErrInvalidProfile
		}
	}
	var nid *string
	if len(nationalID) > 0 {
		v := identity.NormalizeNationalID(nationalID[0])
		if !identity.ValidNationalID(v) {
			return identity.User{}, identity.ErrInvalidProfile
		}
		nid = &v
	}
	hash := ""
	if password != "" {
		if len(password) < 8 || len(password) > 72 {
			return identity.User{}, identity.ErrInvalidProfile
		}
		encoded, e := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if e != nil {
			return identity.User{}, e
		}
		hash = string(encoded)
	}
	if err := s.updateProfileVersion(ctx, surface, session.User, name, email, hash, nid); err != nil {
		return identity.User{}, err
	}
	updated, e := s.users.FindByPhone(ctx, surface, session.User.Phone)
	if e != nil {
		return identity.User{}, e
	}
	session.User = updated
	if e = s.sessions.Put(ctx, session); e != nil {
		return identity.User{}, e
	}
	return updated, nil
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
	if len(password) < 8 || len(password) > 72 {
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
	return s.updateProfileVersion(ctx, surface, user, user.Name, user.Email, string(hash), nil)
}

// AdminResetPassword lets an authenticated super-admin reset a user's password.
func (s *Service) AdminResetPassword(ctx context.Context, token string, surface identity.Surface, targetPhone, password string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	admin, err := s.Authenticate(ctx, token, identity.SurfaceBackoffice)
	if err != nil || admin.User.Role != identity.RoleSuperAdmin {
		return identity.ErrNotAdmin
	}
	if len(password) < 8 || len(password) > 72 {
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
	setter, ok := s.users.(interface {
		SetTemporaryPassword(context.Context, identity.Surface, string, string, identity.User, string, time.Time) error
	})
	if !ok {
		return identity.ErrInvalidProfile
	}
	auditID, e := s.random.Token()
	if e != nil {
		return e
	}
	return setter.SetTemporaryPassword(ctx, surface, user.ID, string(hash), admin.User, auditID, s.clock.Now())
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
	return s.authenticate(ctx, token, surface, false)
}

// ResumeSession restores identity only; restricted sessions still cannot access protected operations.
func (s *Service) ResumeSession(ctx context.Context, token string, surface identity.Surface) (identity.Session, error) {
	return s.authenticate(ctx, token, surface, true)
}
func (s *Service) authenticate(ctx context.Context, token string, surface identity.Surface, restricted bool) (identity.Session, error) {
	session, err := s.sessions.Get(ctx, token)
	if err != nil || session.User.Surface != surface || !s.clock.Now().Before(session.Expires) {
		return identity.Session{}, identity.ErrInvalidOTP
	}
	user, err := s.users.FindByPhone(ctx, surface, session.User.Phone)
	if err != nil || !user.Active || user.ID != session.User.ID || user.AccessVersion != session.User.AccessVersion {
		return identity.Session{}, identity.ErrInactive
	}
	if user.PasswordChangeRequired && !restricted {
		return identity.Session{}, identity.ErrPasswordChangeRequired
	}
	session.User = user
	return session, nil
}
func challengeKey(surface identity.Surface, phone string) string {
	return string(surface) + ":" + phone
}

// Restricted tokens can only finish this operation or log out; they cannot read profiles/workspaces/carts.
func (s *Service) CompleteTemporaryPassword(ctx context.Context, token string, surface identity.Surface, password string) (identity.Session, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	session, e := s.authenticate(ctx, token, surface, true)
	if e != nil {
		return identity.Session{}, e
	}
	if !session.User.PasswordChangeRequired || len(password) < 8 || len(password) > 72 || bcrypt.CompareHashAndPassword([]byte(session.User.PasswordHash), []byte(password)) == nil {
		return identity.Session{}, identity.ErrInvalidProfile
	}
	setter, ok := s.users.(interface {
		ChangeTemporaryPassword(context.Context, identity.Surface, string, int, string, string, time.Time) error
	})
	if !ok {
		return identity.Session{}, identity.ErrInvalidProfile
	}
	hash, e := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if e != nil {
		return identity.Session{}, e
	}
	auditID, e := s.random.Token()
	if e != nil {
		return identity.Session{}, e
	}
	if e = setter.ChangeTemporaryPassword(ctx, surface, session.User.ID, session.User.AccessVersion, string(hash), auditID, s.clock.Now()); e != nil {
		return identity.Session{}, e
	}
	user, e := s.users.FindByPhone(ctx, surface, session.User.Phone)
	if e != nil {
		return identity.Session{}, e
	}
	_ = s.sessions.Delete(ctx, token)
	return s.createSession(ctx, user)
}

func (s *Service) updateProfileVersion(ctx context.Context, surface identity.Surface, user identity.User, name, email, hash string, nid *string) error {
	if r, ok := s.users.(interface {
		UpdateProfileVersion(context.Context, identity.Surface, string, string, string, string, *string, *int) error
	}); ok {
		return r.UpdateProfileVersion(ctx, surface, user.ID, name, email, hash, nid, &user.AccessVersion)
	}
	return s.users.UpdateProfile(ctx, surface, user.ID, name, email, hash, nid)
}
