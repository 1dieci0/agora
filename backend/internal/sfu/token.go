package sfu

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"os"
	"strings"
	"time"
)

type TokenClaims struct {
	UserID    int   `json:"user_id"`
	ChannelID int   `json:"channel_id"`
	ExpiresAt int64 `json:"expires_at"`
}

func signToken(claims TokenClaims) (string, error) {
	secret := os.Getenv("SFU_SECRET")
	if secret == "" {
		return "", errors.New("SFU_SECRET is not configured")
	}

	payload, err := json.Marshal(claims)
	if err != nil {
		return "", err
	}

	encodedPayload := base64.RawURLEncoding.EncodeToString(payload)

	mac := hmac.New(
		sha256.New,
		[]byte(secret),
	)

	mac.Write([]byte(encodedPayload))

	signature := base64.RawURLEncoding.EncodeToString(
		mac.Sum(nil),
	)

	return encodedPayload + "." + signature, nil
}

func verifyToken(token string) (TokenClaims, error) {
	secret := os.Getenv("SFU_SECRET")
	if secret == "" {
		return TokenClaims{}, errors.New("SFU_SECRET is not configured")
	}

	parts := strings.Split(token, ".")
	if len(parts) != 2 {
		return TokenClaims{}, errors.New("invalid token")
	}

	payload := parts[0]
	providedSignature := parts[1]

	mac := hmac.New(
		sha256.New,
		[]byte(secret),
	)

	mac.Write([]byte(payload))

	expectedSignature := base64.RawURLEncoding.EncodeToString(
		mac.Sum(nil),
	)

	if !hmac.Equal(
		[]byte(providedSignature),
		[]byte(expectedSignature),
	) {
		return TokenClaims{}, errors.New("invalid token signature")
	}

	data, err := base64.RawURLEncoding.DecodeString(payload)
	if err != nil {
		return TokenClaims{}, errors.New("invalid token payload")
	}

	var claims TokenClaims

	if err := json.Unmarshal(data, &claims); err != nil {
		return TokenClaims{}, errors.New("invalid token claims")
	}

	if claims.UserID <= 0 {
		return TokenClaims{}, errors.New("invalid user ID")
	}

	if claims.ChannelID <= 0 {
		return TokenClaims{}, errors.New("invalid channel ID")
	}

	if time.Now().Unix() >= claims.ExpiresAt {
		return TokenClaims{}, errors.New("token expired")
	}

	return claims, nil
}

func CreateSFUToken(
	userID int,
	channelID int,
) (string, error) {
	return signToken(TokenClaims{
		UserID:    userID,
		ChannelID: channelID,
		ExpiresAt: time.Now().Add(1 * time.Minute).Unix(),
	})
}

func ParseSFUToken(token string) (TokenClaims, error) {
	return verifyToken(token)
}
