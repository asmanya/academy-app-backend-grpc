package utils

import (
	"context"
	"errors"
)

type ContextKey string

func AuthorizeUser(ctx context.Context, allowedRoles ...string) error {
	userRole, ok := ctx.Value(ContextKey("role")).(string)
	if !ok {
		return errors.New("user not authorized for access: role not found")
	}

	for _, allowallowedRole := range allowedRoles {
		if allowallowedRole == userRole {
			return nil
		}
	}

	return errors.New("user not authorized for access")
}
