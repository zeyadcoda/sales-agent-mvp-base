package email

import (
	"bytes"
	"fmt"
	"net/mail"
	"strings"

	"salesagent.local/backend/internal/platform/auth"
)

const otpEmailSubject = "Your Super Admin authentication code"

func buildOTPMessage(
	fromAddress string,
	fromName string,
	message auth.OTPEmail,
) ([]byte, string, error) {
	recipient, err := validateBareMailbox(message.RecipientEmail)
	if err != nil {
		return nil, "", ErrInvalidOTPEmail
	}
	if err := validateOTP(message.OTP); err != nil {
		return nil, "", err
	}
	if message.ExpiresAt.IsZero() {
		return nil, "", ErrInvalidOTPEmail
	}

	displayName := strings.TrimSpace(message.DisplayName)
	if displayName != "" {
		displayName, err = validateDisplayName(displayName)
		if err != nil {
			return nil, "", err
		}
	}

	from := (&mail.Address{Name: fromName, Address: fromAddress}).String()
	to := (&mail.Address{Address: recipient}).String()

	var content bytes.Buffer
	writeHeader(&content, "From", from)
	writeHeader(&content, "To", to)
	writeHeader(&content, "Subject", otpEmailSubject)
	writeHeader(&content, "MIME-Version", "1.0")
	writeHeader(&content, "Content-Type", "text/plain; charset=UTF-8")
	writeHeader(&content, "Content-Transfer-Encoding", "8bit")
	content.WriteString("\r\n")

	if displayName == "" {
		content.WriteString("Hello,\r\n\r\n")
	} else {
		fmt.Fprintf(&content, "Hello %s,\r\n\r\n", displayName)
	}
	content.WriteString("Your six-digit code for Super Admin authentication is:\r\n\r\n")
	content.WriteString(message.OTP)
	content.WriteString("\r\n\r\n")
	content.WriteString("This code expires in approximately 10 minutes")
	fmt.Fprintf(&content, " (at %s UTC).\r\n\r\n", message.ExpiresAt.UTC().Format("15:04"))
	content.WriteString("If you did not initiate this login, ignore this email.\r\n")

	return content.Bytes(), recipient, nil
}

func writeHeader(buffer *bytes.Buffer, name string, value string) {
	buffer.WriteString(name)
	buffer.WriteString(": ")
	buffer.WriteString(value)
	buffer.WriteString("\r\n")
}

func validateBareMailbox(value string) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" || strings.ContainsAny(value, "\r\n\x00") {
		return "", ErrInvalidOTPEmail
	}

	parsed, err := mail.ParseAddress(value)
	if err != nil || parsed.Name != "" || parsed.Address != value {
		return "", ErrInvalidOTPEmail
	}

	return value, nil
}
