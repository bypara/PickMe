package auth

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
)

const (
	// Базовые параметры OWASP для Argon2id. Память в KiB.
	passwordMemory  = 19 * 1024
	passwordTime    = 2
	passwordThreads = 1
	passwordSaltLen = 16
	passwordHashLen = 32
)

var (
	errPasswordHash   = errors.New("invalid stored Argon2id hash")
	dummyPasswordHash = encodePasswordHash(make([]byte, passwordSaltLen), make([]byte, passwordHashLen))
)

func hashPassword(password string) string {
	salt := make([]byte, passwordSaltLen)
	rand.Read(salt)
	digest := argon2.IDKey([]byte(password), salt, passwordTime, passwordMemory, passwordThreads, passwordHashLen)
	return encodePasswordHash(salt, digest)
}

func verifyPassword(password, encoded string) (bool, error) {
	parts := strings.Split(encoded, "$")
	if len(parts) != 6 || parts[0] != "" || parts[1] != "argon2id" || parts[2] != fmt.Sprintf("v=%d", argon2.Version) {
		return false, errPasswordHash
	}

	var memory, iterations uint32
	var threads uint8
	if _, err := fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &memory, &iterations, &threads); err != nil {
		return false, errPasswordHash
	}
	// Проверяем границы нагрузки до запуска Argon2id с параметрами из БД.
	if threads < 1 || threads > 4 || memory < 8*uint32(threads) || memory > 64*1024 || iterations < 1 || iterations > 4 {
		return false, errPasswordHash
	}
	if parts[3] != fmt.Sprintf("m=%d,t=%d,p=%d", memory, iterations, threads) {
		return false, errPasswordHash
	}

	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil || len(salt) != passwordSaltLen {
		return false, errPasswordHash
	}
	expected, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil || len(expected) != passwordHashLen {
		return false, errPasswordHash
	}

	actual := argon2.IDKey([]byte(password), salt, iterations, memory, threads, passwordHashLen)
	return subtle.ConstantTimeCompare(actual, expected) == 1, nil
}

func encodePasswordHash(salt, digest []byte) string {
	return fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version, passwordMemory, passwordTime, passwordThreads,
		base64.RawStdEncoding.EncodeToString(salt), base64.RawStdEncoding.EncodeToString(digest))
}
