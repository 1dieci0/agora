package users

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"time"
)

type contextKey string

const userIDKey contextKey = "userID"

func UserIDFromContext(ctx context.Context) (int, bool) {
	userID, ok := ctx.Value(userIDKey).(int)
	return userID, ok
}

func setUserIDContext(ctx context.Context, userID int) context.Context {
	return context.WithValue(ctx, userIDKey, userID)
}

func createSession(db *sql.DB, userID int) (string, error) {
	bytes := make([]byte, 32)

	_, err := rand.Read(bytes)
	if err != nil {
		return "", err
	}

	sessionID := hex.EncodeToString(bytes)

	_, err = db.Exec(
		`INSERT INTO sessions (id, user_id, expires_at)
		 VALUES (?, ?, ?)`,
		sessionID,
		userID,
		time.Now().Add(30*24*time.Hour),
	)

	if err != nil {
		return "", err
	}

	return sessionID, nil
}

func deleteSession(db *sql.DB, sessionID string) error {
	_, err := db.Exec(
		"DELETE FROM sessions WHERE id = ?",
		sessionID,
	)

	return err
}
