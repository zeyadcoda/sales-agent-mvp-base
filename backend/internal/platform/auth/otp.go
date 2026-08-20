package auth

import (
	"crypto/hmac"
	cryptorand "crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"math/big"
	"strings"
	"time"
	"unicode/utf8"
)

const (
	challengeIDBytes            = 32
	otpDigits                   = 6
	otpHashSecretMinBytes       = 32
	otpValidity                 = 10 * time.Minute
	otpResendCooldown           = 60 * time.Second
	maxOTPFailedAttempts        = 5
	maxOTPRotationGenerateTries = 8
)

var million = big.NewInt(1_000_000)

type otpHasher struct {
	secret []byte
}

func newOTPHasher(secret []byte) (*otpHasher, error) {
	if len(secret) < otpHashSecretMinBytes {
		return nil, errors.New("OTP hash secret must contain at least 32 bytes")
	}

	return &otpHasher{secret: append([]byte(nil), secret...)}, nil
}

// hash binds the low-entropy code to its unguessable challenge identifier and
// a domain separator. A database-only compromise therefore cannot test the
// one-million-code space without also obtaining the server-held secret.
func (hasher *otpHasher) hash(challengeID string, otp string) ([sha256.Size]byte, error) {
	if hasher == nil || len(hasher.secret) < otpHashSecretMinBytes {
		return [sha256.Size]byte{}, errors.New("OTP hasher is unavailable")
	}
	if !validChallengeID(challengeID) || !validOTP(otp) {
		return [sha256.Size]byte{}, errors.New("invalid OTP hash input")
	}

	mac := hmac.New(sha256.New, hasher.secret)
	_, _ = mac.Write([]byte("sales-agent/super-admin-otp/v1\x00"))
	_, _ = mac.Write([]byte(challengeID))
	_, _ = mac.Write([]byte{0})
	_, _ = mac.Write([]byte(otp))

	var digest [sha256.Size]byte
	copy(digest[:], mac.Sum(nil))
	return digest, nil
}

// otpHashesEqual validates both operands before hmac.Equal. Malformed stored
// state can never be interpreted as a match.
func otpHashesEqual(stored, candidate []byte) bool {
	if len(stored) != sha256.Size || len(candidate) != sha256.Size {
		return false
	}

	return hmac.Equal(stored, candidate)
}

func generateChallengeID(random io.Reader) (string, error) {
	if random == nil {
		random = cryptorand.Reader
	}

	material := make([]byte, challengeIDBytes)
	if _, err := io.ReadFull(random, material); err != nil {
		return "", fmt.Errorf("generate OTP challenge identifier: %w", err)
	}

	return base64.RawURLEncoding.EncodeToString(material), nil
}

func validChallengeID(challengeID string) bool {
	material, err := base64.RawURLEncoding.DecodeString(challengeID)
	if err != nil || len(material) != challengeIDBytes {
		return false
	}

	return base64.RawURLEncoding.EncodeToString(material) == challengeID
}

// ValidOTPChallengeID exposes the canonical public-identifier check to the
// HTTP DTO boundary. Service and repository paths still revalidate it.
func ValidOTPChallengeID(challengeID string) bool {
	return validChallengeID(challengeID)
}

func generateSixDigitOTP(random io.Reader) (string, error) {
	if random == nil {
		random = cryptorand.Reader
	}

	value, err := cryptorand.Int(random, million)
	if err != nil {
		return "", fmt.Errorf("generate OTP: %w", err)
	}

	return fmt.Sprintf("%06d", value.Int64()), nil
}

func validOTP(otp string) bool {
	if len(otp) != otpDigits {
		return false
	}
	for index := 0; index < len(otp); index++ {
		if otp[index] < '0' || otp[index] > '9' {
			return false
		}
	}

	return true
}

// ValidOTP reports whether a value is exactly six ASCII decimal digits.
func ValidOTP(otp string) bool {
	return validOTP(otp)
}

func destinationHint(email string) string {
	local, domain, ok := strings.Cut(email, "@")
	if !ok || local == "" || domain == "" {
		return ""
	}

	first, _ := utf8.DecodeRuneInString(local)
	if first == utf8.RuneError {
		return "***@" + domain
	}

	return string(first) + "***@" + domain
}
