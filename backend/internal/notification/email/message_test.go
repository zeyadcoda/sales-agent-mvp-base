package email

import (
	"errors"
	"strings"
	"testing"
	"time"

	"salesagent.local/backend/internal/platform/auth"
)

func TestBuildOTPMessageContainsOnlySafeAuthenticationContent(t *testing.T) {
	expiresAt := time.Date(2026, time.August, 19, 9, 30, 0, 0, time.UTC)
	content, recipient, err := buildOTPMessage(
		"no-reply@sales-agent.local",
		"Sales Agent",
		auth.OTPEmail{
			RecipientEmail: "admin@example.com",
			DisplayName:    "Super Admin",
			OTP:            "001284",
			ExpiresAt:      expiresAt,
		},
	)
	if err != nil {
		t.Fatalf("buildOTPMessage() returned unexpected error: %v", err)
	}
	if recipient != "admin@example.com" {
		t.Fatalf("recipient = %q, want admin@example.com", recipient)
	}

	message := string(content)
	for _, required := range []string{
		"From: \"Sales Agent\" <no-reply@sales-agent.local>\r\n",
		"To: <admin@example.com>\r\n",
		"Subject: Your Super Admin authentication code\r\n",
		"Content-Type: text/plain; charset=UTF-8\r\n",
		"Hello Super Admin,",
		"001284",
		"Super Admin authentication",
		"approximately 10 minutes",
		"09:30 UTC",
		"If you did not initiate this login, ignore this email.",
	} {
		if !strings.Contains(message, required) {
			t.Errorf("message does not contain %q\n%s", required, message)
		}
	}
}

func TestBuildOTPMessageSupportsAnEmptyDisplayName(t *testing.T) {
	content, _, err := buildOTPMessage(
		"no-reply@sales-agent.local",
		"Sales Agent",
		auth.OTPEmail{
			RecipientEmail: "admin@example.com",
			OTP:            "999999",
			ExpiresAt:      time.Now().Add(10 * time.Minute),
		},
	)
	if err != nil {
		t.Fatalf("buildOTPMessage() returned unexpected error: %v", err)
	}
	if !strings.Contains(string(content), "\r\n\r\nHello,\r\n") {
		t.Fatal("message should use a generic greeting when no display name is available")
	}
}

func TestBuildOTPMessageRejectsMalformedInput(t *testing.T) {
	valid := auth.OTPEmail{
		RecipientEmail: "admin@example.com",
		DisplayName:    "Super Admin",
		OTP:            "093221",
		ExpiresAt:      time.Now().Add(10 * time.Minute),
	}

	tests := []struct {
		name   string
		mutate func(*auth.OTPEmail)
	}{
		{name: "short OTP", mutate: func(message *auth.OTPEmail) { message.OTP = "12345" }},
		{name: "long OTP", mutate: func(message *auth.OTPEmail) { message.OTP = "1234567" }},
		{name: "non-numeric OTP", mutate: func(message *auth.OTPEmail) { message.OTP = "12a456" }},
		{name: "Unicode digits", mutate: func(message *auth.OTPEmail) { message.OTP = "١٢٣٤٥٦" }},
		{name: "missing expiry", mutate: func(message *auth.OTPEmail) { message.ExpiresAt = time.Time{} }},
		{name: "recipient display name", mutate: func(message *auth.OTPEmail) { message.RecipientEmail = "Admin <admin@example.com>" }},
		{name: "recipient injection", mutate: func(message *auth.OTPEmail) { message.RecipientEmail = "admin@example.com\r\nBcc: victim@example.com" }},
		{name: "display name injection", mutate: func(message *auth.OTPEmail) { message.DisplayName = "Admin\nBcc" }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			message := valid
			tt.mutate(&message)

			_, _, err := buildOTPMessage("no-reply@example.com", "Sales Agent", message)
			if !errors.Is(err, ErrInvalidOTPEmail) {
				t.Fatalf("buildOTPMessage() error = %v, want %v", err, ErrInvalidOTPEmail)
			}
		})
	}
}
