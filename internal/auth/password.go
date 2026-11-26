package auth

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"

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

func hashPassword(password string) string {
	salt := make([]byte, passwordSaltLen)
	rand.Read(salt)
	digest := argon2.IDKey([]byte(password), salt, passwordTime, passwordMemory, passwordThreads, passwordHashLen)
	return encodePasswordHash(salt, digest)
}

func encodePasswordHash(salt, digest []byte) string {
	return fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version, passwordMemory, passwordTime, passwordThreads,
		base64.RawStdEncoding.EncodeToString(salt), base64.RawStdEncoding.EncodeToString(digest))
}
