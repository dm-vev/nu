package multitenancy

import (
	"context"
	"errors"
)

type contextKey string

const (
	// orgIDKey is the context key for the organization ID
	orgIDKey contextKey = "org_id"
)

var (
	// ErrNoOrgID is returned when no organization ID is found in the context
	ErrNoOrgID = errors.New("no organization ID found in context")
)

// WithOrgID returns a new context with the given organization ID
func WithOrgID(ctx context.Context, orgID string) context.Context {
	return context.WithValue(ctx, orgIDKey, orgID)
}

// GetOrgID returns the organization ID from the context
func GetOrgID(ctx context.Context) (string, error) {
	orgID, ok := ctx.Value(orgIDKey).(string)
	if !ok || orgID == "" {
		return "", ErrNoOrgID
	}
	return orgID, nil
}

// MustGetOrgID returns the organization ID from the context or panics
func MustGetOrgID(ctx context.Context) string {
	orgID, err := GetOrgID(ctx)
	if err != nil {
		panic(err)
	}
	return orgID
}

// HasOrgID returns true if the context has an organization ID
func HasOrgID(ctx context.Context) bool {
	_, err := GetOrgID(ctx)
	return err == nil
}
