package auth

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"

	"golang.org/x/crypto/argon2"
)

const (
	argon2Algorithm = "argon2id"

	// These parameters meet the approved minimum and live in one place so a
	// future rehash policy can upgrade them without changing the PHC contract.
	defaultArgon2Memory      uint32 = 19 * 1024
	defaultArgon2Iterations  uint32 = 2
	defaultArgon2Parallelism uint8  = 1
	defaultSaltLength               = 16
	defaultKeyLength                = 32

	maximumEncodedHashLength = 512
	minimumDecodedSaltLength = 16
	maximumDecodedSaltLength = 64
	minimumDecodedKeyLength  = 16
	maximumDecodedKeyLength  = 64

	// Verification bounds prevent a corrupted or attacker-controlled database
	// value from forcing excessive Argon2 CPU or memory consumption.
	minimumVerificationMemory      uint32 = defaultArgon2Memory
	maximumVerificationMemory      uint32 = 64 * 1024
	minimumVerificationIterations  uint32 = defaultArgon2Iterations
	maximumVerificationIterations  uint32 = 10
	minimumVerificationParallelism uint8  = 1
	maximumVerificationParallelism uint8  = 8
)

var (
	ErrInvalidPasswordHash       = errors.New("invalid encoded password hash")
	ErrPasswordHasherUnavailable = errors.New("password hasher is not initialized")
)

type argon2Parameters struct {
	memory      uint32
	iterations  uint32
	parallelism uint8
	saltLength  int
	keyLength   uint32
}

var currentArgon2Parameters = argon2Parameters{
	memory:      defaultArgon2Memory,
	iterations:  defaultArgon2Iterations,
	parallelism: defaultArgon2Parallelism,
	saltLength:  defaultSaltLength,
	keyLength:   defaultKeyLength,
}

// PasswordHasher owns password hashing and verification policy. Keeping the
// random source private prevents callers from accidentally supplying salts.
type PasswordHasher struct {
	parameters argon2Parameters
	random     io.Reader
}

func NewPasswordHasher() *PasswordHasher {
	return &PasswordHasher{
		parameters: currentArgon2Parameters,
		random:     rand.Reader,
	}
}

// Hash derives an Argon2id password hash with a fresh cryptographic salt and
// encodes every verification parameter in the versioned PHC representation.
func (h *PasswordHasher) Hash(password string) (string, error) {
	if h == nil || h.random == nil || !validHashParameters(h.parameters) {
		return "", ErrPasswordHasherUnavailable
	}

	salt := make([]byte, h.parameters.saltLength)
	if _, err := io.ReadFull(h.random, salt); err != nil {
		return "", fmt.Errorf("generate password salt: %w", err)
	}

	derivedKey := argon2.IDKey(
		[]byte(password),
		salt,
		h.parameters.iterations,
		h.parameters.memory,
		h.parameters.parallelism,
		h.parameters.keyLength,
	)
	defer clear(derivedKey)

	return fmt.Sprintf(
		"$%s$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2Algorithm,
		argon2.Version,
		h.parameters.memory,
		h.parameters.iterations,
		h.parameters.parallelism,
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(derivedKey),
	), nil
}

// Verify parses a stored PHC value within strict resource bounds and compares
// the derived key in constant time. Malformed database values fail safely.
func (h *PasswordHasher) Verify(encodedHash string, password string) (bool, error) {
	if h == nil {
		return false, ErrPasswordHasherUnavailable
	}

	parsed, err := parsePasswordHash(encodedHash)
	if err != nil {
		return false, err
	}

	actualHash := argon2.IDKey(
		[]byte(password),
		parsed.salt,
		parsed.iterations,
		parsed.memory,
		parsed.parallelism,
		uint32(len(parsed.hash)),
	)
	defer clear(actualHash)

	return subtle.ConstantTimeCompare(actualHash, parsed.hash) == 1, nil
}

type parsedPasswordHash struct {
	memory      uint32
	iterations  uint32
	parallelism uint8
	salt        []byte
	hash        []byte
}

func parsePasswordHash(encodedHash string) (parsedPasswordHash, error) {
	if encodedHash == "" || len(encodedHash) > maximumEncodedHashLength {
		return parsedPasswordHash{}, invalidPasswordHash("invalid encoded length")
	}

	parts := strings.Split(encodedHash, "$")
	if len(parts) != 6 || parts[0] != "" {
		return parsedPasswordHash{}, invalidPasswordHash("invalid PHC fields")
	}

	if parts[1] != argon2Algorithm {
		return parsedPasswordHash{}, invalidPasswordHash("unsupported algorithm")
	}

	version, err := parsePHCUint(parts[2], "v=", 32)
	if err != nil || version != uint64(argon2.Version) {
		return parsedPasswordHash{}, invalidPasswordHash("unsupported version")
	}

	parameterParts := strings.Split(parts[3], ",")
	if len(parameterParts) != 3 {
		return parsedPasswordHash{}, invalidPasswordHash("invalid parameters")
	}

	memoryValue, err := parsePHCUint(parameterParts[0], "m=", 32)
	if err != nil {
		return parsedPasswordHash{}, invalidPasswordHash("invalid memory parameter")
	}
	iterationsValue, err := parsePHCUint(parameterParts[1], "t=", 32)
	if err != nil {
		return parsedPasswordHash{}, invalidPasswordHash("invalid iteration parameter")
	}
	parallelismValue, err := parsePHCUint(parameterParts[2], "p=", 8)
	if err != nil {
		return parsedPasswordHash{}, invalidPasswordHash("invalid parallelism parameter")
	}

	memory := uint32(memoryValue)
	iterations := uint32(iterationsValue)
	parallelism := uint8(parallelismValue)
	if memory < minimumVerificationMemory || memory > maximumVerificationMemory ||
		iterations < minimumVerificationIterations || iterations > maximumVerificationIterations ||
		parallelism < minimumVerificationParallelism || parallelism > maximumVerificationParallelism {
		return parsedPasswordHash{}, invalidPasswordHash("parameters outside verification bounds")
	}

	salt, err := decodePHCField(parts[4], maximumDecodedSaltLength)
	if err != nil || len(salt) < minimumDecodedSaltLength {
		return parsedPasswordHash{}, invalidPasswordHash("invalid salt")
	}

	derivedHash, err := decodePHCField(parts[5], maximumDecodedKeyLength)
	if err != nil || len(derivedHash) < minimumDecodedKeyLength {
		return parsedPasswordHash{}, invalidPasswordHash("invalid derived hash")
	}

	return parsedPasswordHash{
		memory:      memory,
		iterations:  iterations,
		parallelism: parallelism,
		salt:        salt,
		hash:        derivedHash,
	}, nil
}

func parsePHCUint(field string, prefix string, bitSize int) (uint64, error) {
	if !strings.HasPrefix(field, prefix) || len(field) == len(prefix) {
		return 0, ErrInvalidPasswordHash
	}

	value := strings.TrimPrefix(field, prefix)
	parsed, err := strconv.ParseUint(value, 10, bitSize)
	if err != nil || strconv.FormatUint(parsed, 10) != value {
		return 0, ErrInvalidPasswordHash
	}

	return parsed, nil
}

func decodePHCField(field string, maximumDecodedLength int) ([]byte, error) {
	if field == "" || strings.Contains(field, "=") || len(field) > base64.RawStdEncoding.EncodedLen(maximumDecodedLength) {
		return nil, ErrInvalidPasswordHash
	}

	decoded, err := base64.RawStdEncoding.DecodeString(field)
	if err != nil || len(decoded) > maximumDecodedLength {
		return nil, ErrInvalidPasswordHash
	}

	// Re-encoding rejects alternate/non-canonical representations that could
	// otherwise produce multiple strings for the same stored material.
	if base64.RawStdEncoding.EncodeToString(decoded) != field {
		return nil, ErrInvalidPasswordHash
	}

	return decoded, nil
}

func validHashParameters(parameters argon2Parameters) bool {
	return parameters.memory >= minimumVerificationMemory &&
		parameters.memory <= maximumVerificationMemory &&
		parameters.iterations >= minimumVerificationIterations &&
		parameters.iterations <= maximumVerificationIterations &&
		parameters.parallelism >= minimumVerificationParallelism &&
		parameters.parallelism <= maximumVerificationParallelism &&
		parameters.saltLength >= minimumDecodedSaltLength &&
		parameters.saltLength <= maximumDecodedSaltLength &&
		parameters.keyLength >= minimumDecodedKeyLength &&
		parameters.keyLength <= maximumDecodedKeyLength
}

func invalidPasswordHash(reason string) error {
	return fmt.Errorf("%w: %s", ErrInvalidPasswordHash, reason)
}
