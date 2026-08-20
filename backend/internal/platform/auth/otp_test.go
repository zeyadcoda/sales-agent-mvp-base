package auth

import (
	"bytes"
	"crypto/sha256"
	"errors"
	"io"
	"strings"
	"testing"
)

func TestGenerateSixDigitOTPUsesNumericFixedWidthFormat(t *testing.T) {
	t.Parallel()

	for attempt := 0; attempt < 100; attempt++ {
		otp, err := generateSixDigitOTP(nil)
		if err != nil {
			t.Fatalf("generateSixDigitOTP() error = %v", err)
		}
		if !validOTP(otp) {
			t.Fatalf("generated OTP %q is not exactly six ASCII digits", otp)
		}
	}
}

func TestGenerateSixDigitOTPPreservesLeadingZeros(t *testing.T) {
	t.Parallel()

	otp, err := generateSixDigitOTP(bytes.NewReader([]byte{0, 0, 0}))
	if err != nil {
		t.Fatalf("generateSixDigitOTP() error = %v", err)
	}
	if otp != "000000" {
		t.Fatalf("generated OTP = %q, want 000000", otp)
	}
}

func TestGenerateSixDigitOTPPropagatesSecureRandomFailure(t *testing.T) {
	t.Parallel()

	if _, err := generateSixDigitOTP(errorReader{}); err == nil {
		t.Fatal("generateSixDigitOTP() error = nil, want random-source failure")
	}
}

func TestChallengeIdentifierIsCanonical256BitBase64URL(t *testing.T) {
	t.Parallel()

	id, err := generateChallengeID(bytes.NewReader(bytes.Repeat([]byte{11}, challengeIDBytes)))
	if err != nil {
		t.Fatalf("generateChallengeID() error = %v", err)
	}
	if len(id) != 43 || !validChallengeID(id) {
		t.Fatalf("challenge ID = %q, want canonical 32-byte base64url", id)
	}
	if strings.ContainsAny(id, "+/=") {
		t.Fatalf("challenge ID %q is not unpadded base64url", id)
	}
	for _, invalid := range []string{"", id + "=", id[:42], strings.Repeat("a", 43)} {
		if validChallengeID(invalid) {
			t.Fatalf("validChallengeID(%q) = true, want false", invalid)
		}
	}
}

func TestOTPHasherVerifiesOnlyMatchingChallengeAndCode(t *testing.T) {
	t.Parallel()

	hasher, err := newOTPHasher(bytes.Repeat([]byte{5}, otpHashSecretMinBytes))
	if err != nil {
		t.Fatalf("newOTPHasher() error = %v", err)
	}
	challengeA := tokenFromByte(1)
	challengeB := tokenFromByte(2)

	stored, err := hasher.hash(challengeA, "001284")
	if err != nil {
		t.Fatalf("hash() error = %v", err)
	}
	matching, _ := hasher.hash(challengeA, "001284")
	wrongCode, _ := hasher.hash(challengeA, "001285")
	wrongChallenge, _ := hasher.hash(challengeB, "001284")

	if !otpHashesEqual(stored[:], matching[:]) {
		t.Fatal("matching OTP hash did not verify")
	}
	if otpHashesEqual(stored[:], wrongCode[:]) {
		t.Fatal("wrong OTP verified")
	}
	if otpHashesEqual(stored[:], wrongChallenge[:]) {
		t.Fatal("OTP hash was not bound to its challenge")
	}
	if bytes.Equal(stored[:], []byte("001284")) || bytes.Contains(stored[:], []byte("001284")) {
		t.Fatal("OTP digest contains plaintext OTP")
	}
}

func TestOTPHashComparisonFailsClosedForMalformedStoredData(t *testing.T) {
	t.Parallel()

	candidate := bytes.Repeat([]byte{1}, sha256.Size)
	for _, malformed := range [][]byte{nil, {}, {1}, bytes.Repeat([]byte{1}, sha256.Size-1), bytes.Repeat([]byte{1}, sha256.Size+1)} {
		if otpHashesEqual(malformed, candidate) {
			t.Fatalf("malformed stored digest length %d verified", len(malformed))
		}
	}
}

func TestOTPHasherRequiresHighEntropySecretLength(t *testing.T) {
	t.Parallel()

	if _, err := newOTPHasher(bytes.Repeat([]byte{1}, otpHashSecretMinBytes-1)); err == nil {
		t.Fatal("newOTPHasher() accepted a secret shorter than 32 bytes")
	}
	if _, err := newOTPHasher(bytes.Repeat([]byte{1}, otpHashSecretMinBytes)); err != nil {
		t.Fatalf("newOTPHasher() rejected a 32-byte secret: %v", err)
	}
}

func TestValidOTPRejectsNonASCIIAndWrongLength(t *testing.T) {
	t.Parallel()

	for _, value := range []string{"12345", "1234567", "12345a", "１２３４５６", "123 56", "\n12345"} {
		if validOTP(value) {
			t.Fatalf("validOTP(%q) = true, want false", value)
		}
	}
}

func TestDestinationHintMinimizesAccountInformation(t *testing.T) {
	t.Parallel()

	if got := destinationHint("admin@example.com"); got != "a***@example.com" {
		t.Fatalf("destinationHint() = %q", got)
	}
	if got := destinationHint("malformed"); got != "" {
		t.Fatalf("destinationHint(malformed) = %q, want empty", got)
	}
}

type errorReader struct{}

func (errorReader) Read(_ []byte) (int, error) {
	return 0, errors.New("simulated secure random failure")
}

var _ io.Reader = errorReader{}
