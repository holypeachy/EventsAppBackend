package helpers

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/holypeachy/EventsAppBackend/auth"
)

func ExtractUserId(ctx context.Context) (uuid.UUID, error) {
	value := ctx.Value(auth.UserIdCtxKey)

	userId, ok := value.(uuid.UUID)
	if !ok {
		return userId, errors.New("could not cast UUID")
	}

	return userId, nil
}
