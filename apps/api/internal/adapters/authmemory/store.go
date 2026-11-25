package authmemory

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/cherriva/marketplace/apps/api/internal/domain/identity"
	"math/big"
	"sync"
	"time"
)

var ErrNotFound = errors.New("not found")

type Clock struct{}

func (Clock) Now() time.Time { return time.Now() }

type Random struct{}

func (Random) Token() (string, error) { return randomHex(24) }
func (Random) Code() (string, error) {
	value, err := rand.Int(rand.Reader, big.NewInt(1000000))
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%06d", value.Int64()), nil
}
func randomBytes(size int) ([]byte, error) {
	b := make([]byte, size)
	_, err := rand.Read(b)
	return b, err
}
func randomHex(size int) (string, error) {
	b, err := randomBytes(size)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

type Store struct {
	mu         sync.RWMutex
	users      map[string]identity.User
	challenges map[string]identity.Challenge
	sessions   map[string]identity.Session
	requests   map[string][]time.Time
	last       map[string]time.Time
}

func NewStore(users []identity.User) *Store {
	byKey := make(map[string]identity.User, len(users))
	for _, u := range users {
		u.Phone = identity.NormalizePhone(u.Phone)
		byKey[key(u.Surface, u.Phone)] = u
	}
	return &Store{users: byKey, challenges: map[string]identity.Challenge{}, sessions: map[string]identity.Session{}, requests: map[string][]time.Time{}, last: map[string]time.Time{}}
}
func (s *Store) FindByPhone(_ context.Context, surface identity.Surface, phone string) (identity.User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	u, ok := s.users[key(surface, phone)]
	if !ok {
		return identity.User{}, identity.ErrUnknown
	}
	return u, nil
}
func (s *Store) FindByEmail(_ context.Context, surface identity.Surface, email string) (identity.User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, u := range s.users {
		if u.Surface == surface && identity.NormalizeEmail(u.Email) == identity.NormalizeEmail(email) && u.Email != "" {
			return u, nil
		}
	}
	return identity.User{}, identity.ErrUnknown
}
func (s *Store) UpdateProfile(_ context.Context, surface identity.Surface, userID, name, email, passwordHash string, nationalID *string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	var key string
	for k, u := range s.users {
		if u.ID == userID && u.Surface == surface {
			key = k
			break
		}
	}
	if key == "" {
		return identity.ErrUnknown
	}
	for _, u := range s.users {
		if u.Surface == surface && u.ID != userID && email != "" && identity.NormalizeEmail(u.Email) == identity.NormalizeEmail(email) {
			return identity.ErrConflict
		}
	}
	u := s.users[key]
	if name != "" {
		u.Name = name
	}
	u.Email = identity.NormalizeEmail(email)
	if nationalID != nil {
		u.NationalID = *nationalID
	}
	if passwordHash != "" {
		u.PasswordHash = passwordHash
	}
	s.users[key] = u
	return nil
}
func (s *Store) PutChallenge(_ context.Context, id string, c identity.Challenge) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.challenges[id] = c
	key := c.CooldownKey()
	if c.IssuedAt.After(s.last[key]) {
		s.last[key] = c.IssuedAt
	}
	return nil
}
func (s *Store) GetChallenge(_ context.Context, id string) (identity.Challenge, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	c, ok := s.challenges[id]
	if !ok {
		return identity.Challenge{}, ErrNotFound
	}
	return c, nil
}
func (s *Store) DeleteChallenge(_ context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.challenges, id)
	return nil
}
func (s *Store) RecordRequest(_ context.Context, k string, at time.Time) ([]time.Time, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	recent := s.requests[k][:0]
	for _, previous := range s.requests[k] {
		if previous.After(at.Add(-time.Hour)) {
			recent = append(recent, previous)
		}
	}
	s.requests[k] = append(recent, at)
	return append([]time.Time(nil), s.requests[k]...), nil
}
func (s *Store) LastRequest(_ context.Context, k string) (time.Time, bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	t, ok := s.last[k]
	return t, ok, nil
}
func (s *Store) Put(_ context.Context, session identity.Session) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.sessions[session.Token] = session
	return nil
}
func (s *Store) Get(_ context.Context, token string) (identity.Session, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	v, ok := s.sessions[token]
	if !ok {
		return identity.Session{}, ErrNotFound
	}
	return v, nil
}
func (s *Store) Delete(_ context.Context, token string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.sessions, token)
	return nil
}
func key(surface identity.Surface, phone string) string { return string(surface) + ":" + phone }
