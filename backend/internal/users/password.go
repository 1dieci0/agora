package users

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"

	"golang.org/x/crypto/argon2"
)

const (
	argonTime    uint32 = 3
	argonMemory  uint32 = 64 * 1024
	argonThreads uint8  = 4
	argonKeyLen  uint32 = 32
	saltLen      uint32 = 16
)

func HashPassword(password string) (string, error) {
	salt := make([]byte, saltLen)

	if _, err := rand.Read(salt); err != nil {
		return "", err
	}

	hash := argon2.IDKey(
		[]byte(password),
		salt,
		argonTime,
		argonMemory,
		argonThreads,
		argonKeyLen,
	)

	return fmt.Sprintf(
		"$argon2id$v=19$m=%d,t=%d,p=%d$%s$%s",
		argonMemory,
		argonTime,
		argonThreads,
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(hash),
	), nil
}

func VerifyPassword(password string, encoded string) (bool, error) {
	parts := strings.Split(encoded, "$")

	if len(parts) != 6 {
		return false, fmt.Errorf("invalid password hash format")
	}

	if parts[1] != "argon2id" {
		return false, fmt.Errorf("unsupported password hash")
	}

	if parts[2] != "v=19" {
		return false, fmt.Errorf("unsupported Argon2 version")
	}

	parameters := strings.Split(parts[3], ",")

	if len(parameters) != 3 {
		return false, fmt.Errorf("invalid Argon2 parameters")
	}

	var memory uint32
	var time uint32
	var threads uint8

	for _, parameter := range parameters {
		parts := strings.SplitN(parameter, "=", 2)

		if len(parts) != 2 {
			return false, fmt.Errorf("invalid Argon2 parameter")
		}

		value, err := strconv.ParseUint(parts[1], 10, 32)
		if err != nil {
			return false, err
		}

		switch parts[0] {
		case "m":
			memory = uint32(value)

		case "t":
			time = uint32(value)

		case "p":
			if value > 255 {
				return false, fmt.Errorf("invalid thread count")
			}

			threads = uint8(value)

		default:
			return false, fmt.Errorf("unknown Argon2 parameter")
		}
	}

	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return false, err
	}

	expectedHash, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return false, err
	}

	actualHash := argon2.IDKey(
		[]byte(password),
		salt,
		time,
		memory,
		threads,
		uint32(len(expectedHash)),
	)

	return subtle.ConstantTimeCompare(actualHash, expectedHash) == 1, nil
}
