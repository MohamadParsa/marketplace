package identity

import "errors"

var (
	ErrInvalidOTP               = errors.New("invalid or expired otp")
	ErrInvalidCredentials       = errors.New("invalid credentials")
	ErrInactive                 = errors.New("account is inactive")
	ErrUnknown                  = errors.New("account is not registered")
	ErrResendSoon               = errors.New("otp resend is not available yet")
	ErrRateLimit                = errors.New("otp rate limit exceeded")
	ErrNotAdmin                 = errors.New("admin access required")
	ErrConflict                 = errors.New("conflict")
	ErrInvalidProfile           = errors.New("invalid profile")
	ErrEmailRecoveryUnavailable = errors.New("email recovery is not enabled")
)
