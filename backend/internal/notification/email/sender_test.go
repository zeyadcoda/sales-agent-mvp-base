package email

import (
	"strings"
	"testing"
	"time"

	"salesagent.local/backend/internal/config"
)

func validMailpitSettings() config.AuthEmailConfig {
	return config.AuthEmailConfig{
		Mode:        config.AuthEmailMailpit,
		FromAddress: "no-reply@sales-agent.local",
		FromName:    "Sales Agent",
		SMTPHost:    "127.0.0.1",
		SMTPPort:    1025,
		SMTPTLSMode: config.SMTPTLSNone,
		SMTPTimeout: 5 * time.Second,
	}
}

func validSMTPSettings() config.AuthEmailConfig {
	return config.AuthEmailConfig{
		Mode:         config.AuthEmailSMTP,
		FromAddress:  "no-reply@example.com",
		FromName:     "Sales Agent",
		SMTPHost:     "smtp.example.com",
		SMTPPort:     587,
		SMTPTLSMode:  config.SMTPTLSSTARTTLS,
		SMTPUsername: "smtp-user",
		SMTPPassword: "smtp-password",
		SMTPTimeout:  5 * time.Second,
	}
}

func TestNewSenderSelectsMailpitAndSMTPAdapters(t *testing.T) {
	mailpitSender, err := NewSender(validMailpitSettings())
	if err != nil {
		t.Fatalf("NewSender(Mailpit) returned unexpected error: %v", err)
	}
	if _, ok := mailpitSender.(*MailpitSender); !ok {
		t.Fatalf("NewSender(Mailpit) returned %T, want *MailpitSender", mailpitSender)
	}

	smtpSender, err := NewSender(validSMTPSettings())
	if err != nil {
		t.Fatalf("NewSender(SMTP) returned unexpected error: %v", err)
	}
	if _, ok := smtpSender.(*SMTPSender); !ok {
		t.Fatalf("NewSender(SMTP) returned %T, want *SMTPSender", smtpSender)
	}
}

func TestNewMailpitSenderRejectsNonLocalOrAuthenticatedSMTP(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*config.AuthEmailConfig)
	}{
		{name: "host", mutate: func(settings *config.AuthEmailConfig) { settings.SMTPHost = "mailpit" }},
		{name: "port", mutate: func(settings *config.AuthEmailConfig) { settings.SMTPPort = 25 }},
		{name: "TLS", mutate: func(settings *config.AuthEmailConfig) { settings.SMTPTLSMode = config.SMTPTLSSTARTTLS }},
		{name: "credentials", mutate: func(settings *config.AuthEmailConfig) {
			settings.SMTPUsername = "user"
			settings.SMTPPassword = "password"
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			settings := validMailpitSettings()
			tt.mutate(&settings)

			if _, err := NewMailpitSender(settings); err == nil {
				t.Fatal("NewMailpitSender() should reject unsafe settings")
			}
		})
	}
}

func TestNewSMTPSenderRequiresTLS(t *testing.T) {
	settings := validSMTPSettings()
	settings.SMTPTLSMode = config.SMTPTLSNone

	if _, err := NewSMTPSender(settings); err == nil {
		t.Fatal("NewSMTPSender() should reject plaintext SMTP")
	}
}

func TestConstructorsRejectHeaderAndCredentialInjection(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*config.AuthEmailConfig)
	}{
		{name: "from address", mutate: func(settings *config.AuthEmailConfig) {
			settings.FromAddress = "sender@example.com\r\nBcc: victim@example.com"
		}},
		{name: "from name", mutate: func(settings *config.AuthEmailConfig) {
			settings.FromName = "Sales Agent\nBcc"
		}},
		{name: "username", mutate: func(settings *config.AuthEmailConfig) {
			settings.SMTPUsername = "user\n"
		}},
		{name: "password", mutate: func(settings *config.AuthEmailConfig) {
			settings.SMTPPassword = "password\r"
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			settings := validSMTPSettings()
			tt.mutate(&settings)

			_, err := NewSMTPSender(settings)
			if err == nil {
				t.Fatal("NewSMTPSender() should reject injected configuration")
			}
			if strings.Contains(err.Error(), "victim@example.com") || strings.Contains(err.Error(), "password\r") {
				t.Fatal("constructor error exposed configuration content")
			}
		})
	}
}
