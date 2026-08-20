package email

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/smtp"
	"time"

	"salesagent.local/backend/internal/config"
	"salesagent.local/backend/internal/platform/auth"
)

type smtpDelivery struct {
	fromAddress string
	fromName    string
	host        string
	port        int
	tlsMode     config.SMTPTLSMode
	username    string
	password    string
	timeout     time.Duration
}

func (d smtpDelivery) sendOTP(ctx context.Context, message auth.OTPEmail) error {
	if ctx == nil {
		return ErrInvalidOTPEmail
	}

	content, recipient, err := buildOTPMessage(d.fromAddress, d.fromName, message)
	if err != nil {
		return err
	}

	return safeDeliveryError(d.deliver(ctx, recipient, content))
}

func (d smtpDelivery) deliver(ctx context.Context, recipient string, content []byte) error {
	operationCtx, cancel := context.WithTimeout(ctx, d.timeout)
	defer cancel()

	address := net.JoinHostPort(d.host, fmt.Sprintf("%d", d.port))
	rawConnection, err := (&net.Dialer{Timeout: d.timeout}).DialContext(operationCtx, "tcp", address)
	if err != nil {
		return err
	}
	defer rawConnection.Close()

	deadline, ok := operationCtx.Deadline()
	if ok {
		if err := rawConnection.SetDeadline(deadline); err != nil {
			return err
		}
	}
	// Capture the immutable transport connection. The protocol wrapper below
	// may become a TLS connection, while cancellation must remain race-free and
	// can always interrupt it by closing the underlying socket.
	stopCancellationWatch := context.AfterFunc(operationCtx, func() {
		_ = rawConnection.Close()
	})
	defer stopCancellationWatch()

	var connection net.Conn = rawConnection
	if d.tlsMode == config.SMTPTLSDirect {
		tlsConnection := tls.Client(connection, d.tlsConfig())
		if err := tlsConnection.HandshakeContext(operationCtx); err != nil {
			return err
		}
		connection = tlsConnection
	}

	client, err := smtp.NewClient(connection, d.host)
	if err != nil {
		return err
	}
	defer client.Close()

	if d.tlsMode == config.SMTPTLSSTARTTLS {
		if supported, _ := client.Extension("STARTTLS"); !supported {
			return fmt.Errorf("SMTP server does not advertise STARTTLS")
		}
		if err := client.StartTLS(d.tlsConfig()); err != nil {
			return err
		}
	}

	if d.username != "" {
		if err := client.Auth(smtp.PlainAuth("", d.username, d.password, d.host)); err != nil {
			return err
		}
	}
	if err := client.Mail(d.fromAddress); err != nil {
		return err
	}
	if err := client.Rcpt(recipient); err != nil {
		return err
	}

	messageWriter, err := client.Data()
	if err != nil {
		return err
	}
	if _, err := messageWriter.Write(content); err != nil {
		_ = messageWriter.Close()
		return err
	}
	if err := messageWriter.Close(); err != nil {
		return err
	}

	return client.Quit()
}

func (d smtpDelivery) tlsConfig() *tls.Config {
	return &tls.Config{
		MinVersion: tls.VersionTLS12,
		ServerName: d.host,
	}
}
