package domain

import (
	"context"
)

type contextKey string

const UserIDKey     contextKey = "user_id"

func GetUserID(ctx context.Context) (int64, bool) {
    userID, ok := ctx.Value(UserIDKey).(int64)
    return userID, ok
}
