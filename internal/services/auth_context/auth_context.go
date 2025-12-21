package authcontext

import "context"

type ContextKey string

const ContextUserIDKey ContextKey = "userID"

func CreateContextWithUserID(ctx context.Context, userID int64) context.Context {
	return context.WithValue(ctx, ContextUserIDKey, userID)
}

func GetUserIDFromContext(ctx context.Context) (userID int64) {
	userIDInterface := ctx.Value(ContextUserIDKey)

	if userIDInterface != nil {
		userID = userIDInterface.(int64)
	}

	return
}
