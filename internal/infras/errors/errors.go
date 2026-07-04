package errors

import "errors"

var (
	ErrMemberNotFound     = errors.New("member not found")
	ErrInsufficientPoints = errors.New("insufficient points")
	ErrInvalidRule        = errors.New("invalid rule")
	ErrDuplicateMember    = errors.New("member already exists")
	ErrInvalidTier        = errors.New("invalid tier")
	ErrPointsExpired      = errors.New("points expired")
	ErrRateLimit          = errors.New("rate limit exceeded")
)
