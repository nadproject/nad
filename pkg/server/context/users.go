package context

import (
	"context"

	"github.com/nadproject/nad/pkg/server/models"
)

const (
	userKey privateKey = "user"
)

type privateKey string

// WithUser creates a new context with the given user
func WithUser(ctx context.Context, user *models.User) context.Context {
	return context.WithValue(ctx, userKey, user)
}

// User retrieves a user from the given context. It returns a pointer to
// a user. If the context does not contain a user, it returns nil.
func User(ctx context.Context) *models.User {
	if temp := ctx.Value(userKey); temp != nil {
		if user, ok := temp.(*models.User); ok {
			return user
		}
	}

	return nil
}
