package core_auth

import "context"

type contextKey struct{}

var identityContextKey = contextKey{}

type Identity struct {
	UserID    string
	SessionID string
}

func ToContext(ctx context.Context, identity Identity) context.Context {
	return context.WithValue(ctx, identityContextKey, identity)
}

func FromContext(ctx context.Context) Identity {
	identity, _ := ctx.Value(identityContextKey).(Identity)
	return identity
}

func UserID(ctx context.Context) string {
	return FromContext(ctx).UserID
}

func SessionID(ctx context.Context) string {
	return FromContext(ctx).SessionID
}
