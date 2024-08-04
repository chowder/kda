package tokensource

import (
	"context"
	"time"
)

type Token struct {
	Value     string
	ExpiresAt time.Time
}

type TokenSource interface {
	Create(ctx context.Context) (*Token, error)
}
