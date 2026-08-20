package email

import (
	"bufio"
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"
	"testing"
	"time"

	"salesagent.local/backend/internal/config"
	"salesagent.local/backend/internal/platform/auth"
)

type capturedSMTPMessage struct {
	mailFrom string
	rcptTo   string
	content  string
}

func TestSMTPDeliverySendsOTPThroughLocalTransport(t *testing.T) {
	host, port, messages := startTestSMTPServer(t, false)
	delivery := smtpDelivery{
		fromAddress: "no-reply@sales-agent.local",
		fromName:    "Sales Agent",
		host:        host,
		port:        port,
		tlsMode:     config.SMTPTLSNone,
		timeout:     2 * time.Second,
	}

	err := delivery.sendOTP(context.Background(), auth.OTPEmail{
		RecipientEmail: "admin@example.com",
		DisplayName:    "Super Admin",
		OTP:            "001284",
		ExpiresAt:      time.Now().Add(10 * time.Minute),
	})
	if err != nil {
		t.Fatalf("sendOTP() returned unexpected error: %v", err)
	}

	select {
	case message := <-messages:
		if !strings.HasPrefix(message.mailFrom, "MAIL FROM:<no-reply@sales-agent.local>") {
			t.Errorf("MAIL FROM = %q", message.mailFrom)
		}
		if message.rcptTo != "RCPT TO:<admin@example.com>" {
			t.Errorf("RCPT TO = %q", message.rcptTo)
		}
		if !strings.Contains(message.content, "001284") ||
			!strings.Contains(message.content, "Super Admin authentication") {
			t.Errorf("captured message did not contain expected OTP content:\n%s", message.content)
		}
	case <-time.After(time.Second):
		t.Fatal("SMTP server did not capture a message")
	}
}

func TestSMTPDeliveryHonorsContextDeadline(t *testing.T) {
	host, port, _ := startTestSMTPServer(t, true)
	delivery := smtpDelivery{
		fromAddress: "no-reply@sales-agent.local",
		fromName:    "Sales Agent",
		host:        host,
		port:        port,
		tlsMode:     config.SMTPTLSNone,
		timeout:     2 * time.Second,
	}
	ctx, cancel := context.WithTimeout(context.Background(), 25*time.Millisecond)
	defer cancel()

	startedAt := time.Now()
	err := delivery.sendOTP(ctx, auth.OTPEmail{
		RecipientEmail: "admin@example.com",
		OTP:            "123456",
		ExpiresAt:      time.Now().Add(10 * time.Minute),
	})
	if !errors.Is(err, ErrDeliveryUnavailable) {
		t.Fatalf("sendOTP() error = %v, want %v", err, ErrDeliveryUnavailable)
	}
	if elapsed := time.Since(startedAt); elapsed > time.Second {
		t.Fatalf("sendOTP() ignored context deadline; elapsed %s", elapsed)
	}
}

func TestSMTPDeliveryDoesNotExposeTransportErrors(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	address := listener.Addr().(*net.TCPAddr)
	if err := listener.Close(); err != nil {
		t.Fatalf("close listener: %v", err)
	}

	delivery := smtpDelivery{
		fromAddress: "no-reply@sales-agent.local",
		fromName:    "Sales Agent",
		host:        "127.0.0.1",
		port:        address.Port,
		tlsMode:     config.SMTPTLSNone,
		timeout:     200 * time.Millisecond,
	}
	err = delivery.sendOTP(context.Background(), auth.OTPEmail{
		RecipientEmail: "admin@example.com",
		OTP:            "123456",
		ExpiresAt:      time.Now().Add(10 * time.Minute),
	})
	if !errors.Is(err, ErrDeliveryUnavailable) {
		t.Fatalf("sendOTP() error = %v, want %v", err, ErrDeliveryUnavailable)
	}
	if strings.Contains(err.Error(), strconv.Itoa(address.Port)) || strings.Contains(err.Error(), "connect") {
		t.Fatalf("sendOTP() exposed a raw transport error: %v", err)
	}
}

func TestSMTPDeliveryFailsClosedWhenSTARTTLSIsUnavailable(t *testing.T) {
	host, port, _ := startTestSMTPServer(t, false)
	delivery := smtpDelivery{
		fromAddress: "no-reply@example.com",
		fromName:    "Sales Agent",
		host:        host,
		port:        port,
		tlsMode:     config.SMTPTLSSTARTTLS,
		timeout:     time.Second,
	}

	err := delivery.sendOTP(context.Background(), auth.OTPEmail{
		RecipientEmail: "admin@example.com",
		OTP:            "123456",
		ExpiresAt:      time.Now().Add(10 * time.Minute),
	})
	if !errors.Is(err, ErrDeliveryUnavailable) {
		t.Fatalf("sendOTP() error = %v, want %v", err, ErrDeliveryUnavailable)
	}
}

func TestSMTPDeliveryTLSConfiguration(t *testing.T) {
	delivery := smtpDelivery{host: "smtp.example.com"}
	tlsConfig := delivery.tlsConfig()

	if tlsConfig.MinVersion != tls.VersionTLS12 {
		t.Fatalf("minimum TLS version = %d, want TLS 1.2", tlsConfig.MinVersion)
	}
	if tlsConfig.ServerName != "smtp.example.com" {
		t.Fatalf("TLS server name = %q, want smtp.example.com", tlsConfig.ServerName)
	}
}

func TestSMTPDeliveryRejectsMalformedOTPBeforeConnecting(t *testing.T) {
	delivery := smtpDelivery{
		fromAddress: "no-reply@sales-agent.local",
		fromName:    "Sales Agent",
		host:        "invalid.example",
		port:        25,
		tlsMode:     config.SMTPTLSNone,
		timeout:     time.Second,
	}

	err := delivery.sendOTP(context.Background(), auth.OTPEmail{
		RecipientEmail: "admin@example.com",
		OTP:            "12345",
		ExpiresAt:      time.Now().Add(10 * time.Minute),
	})
	if !errors.Is(err, ErrInvalidOTPEmail) {
		t.Fatalf("sendOTP() error = %v, want %v", err, ErrInvalidOTPEmail)
	}
}

func startTestSMTPServer(
	t *testing.T,
	stallAfterAccept bool,
) (string, int, <-chan capturedSMTPMessage) {
	t.Helper()

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	t.Cleanup(func() { _ = listener.Close() })

	messages := make(chan capturedSMTPMessage, 1)
	go func() {
		connection, acceptErr := listener.Accept()
		if acceptErr != nil {
			return
		}
		defer connection.Close()
		if stallAfterAccept {
			<-time.After(2 * time.Second)
			return
		}

		reader := bufio.NewReader(connection)
		_, _ = fmt.Fprint(connection, "220 localhost ESMTP ready\r\n")
		var captured capturedSMTPMessage
		for {
			line, readErr := reader.ReadString('\n')
			if readErr != nil {
				return
			}
			line = strings.TrimSuffix(strings.TrimSuffix(line, "\n"), "\r")

			switch {
			case strings.HasPrefix(line, "EHLO "):
				_, _ = fmt.Fprint(connection, "250-localhost\r\n250 8BITMIME\r\n")
			case strings.HasPrefix(line, "HELO "):
				_, _ = fmt.Fprint(connection, "250 localhost\r\n")
			case strings.HasPrefix(line, "MAIL FROM:"):
				captured.mailFrom = line
				_, _ = fmt.Fprint(connection, "250 OK\r\n")
			case strings.HasPrefix(line, "RCPT TO:"):
				captured.rcptTo = line
				_, _ = fmt.Fprint(connection, "250 OK\r\n")
			case line == "DATA":
				_, _ = fmt.Fprint(connection, "354 End data with <CR><LF>.<CR><LF>\r\n")
				var content strings.Builder
				for {
					dataLine, dataErr := reader.ReadString('\n')
					if dataErr != nil {
						return
					}
					if dataLine == ".\r\n" {
						break
					}
					content.WriteString(dataLine)
				}
				captured.content = content.String()
				_, _ = fmt.Fprint(connection, "250 queued\r\n")
			case line == "QUIT":
				_, _ = fmt.Fprint(connection, "221 bye\r\n")
				messages <- captured
				return
			default:
				_, _ = fmt.Fprint(connection, "500 unsupported\r\n")
			}
		}
	}()

	address := listener.Addr().(*net.TCPAddr)
	return address.IP.String(), address.Port, messages
}
