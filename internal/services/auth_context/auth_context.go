package authcontext

import "context"

type ContextKey string

const contextUserIDKey ContextKey = "userID"

func CreateContextWithUserID(ctx context.Context, userID int64) context.Context {
	return context.WithValue(ctx, contextUserIDKey, userID)
}

func GetUserIDFromContext(ctx context.Context) (userID int64) {
	userIDInterface := ctx.Value(contextUserIDKey)

	if userIDInterface != nil {
		userID = userIDInterface.(int64)
	}

	return
}
