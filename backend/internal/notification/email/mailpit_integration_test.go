package email

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"salesagent.local/backend/internal/platform/auth"
)

// TestMailpitIntegration is opt-in because the normal unit suite must not
// require local infrastructure. It verifies the same loopback SMTP and inbox
// endpoints documented for manual development.
func TestMailpitIntegration(t *testing.T) {
	if os.Getenv("TEST_MAILPIT") != "1" {
		t.Skip("set TEST_MAILPIT=1 with Mailpit running on 127.0.0.1:1025/8025")
	}

	sender, err := NewMailpitSender(validMailpitSettings())
	if err != nil {
		t.Fatalf("NewMailpitSender() returned unexpected error: %v", err)
	}
	recipient := "mailpit-integration-" + strconv.FormatInt(time.Now().UnixNano(), 10) + "@example.test"
	const otp = "042681"
	expiresAt := time.Now().UTC().Add(10 * time.Minute)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := sender.SendOTP(ctx, auth.OTPEmail{
		RecipientEmail: recipient,
		DisplayName:    "Mailpit Test Admin",
		OTP:            otp,
		ExpiresAt:      expiresAt,
	}); err != nil {
		t.Fatalf("SendOTP() returned unexpected error: %v", err)
	}

	if err := waitForMailpitMessage(ctx, recipient, otp); err != nil {
		t.Fatal(err)
	}
}

func waitForMailpitMessage(ctx context.Context, recipient string, otp string) error {
	client := &http.Client{Timeout: time.Second}
	ticker := time.NewTicker(50 * time.Millisecond)
	defer ticker.Stop()

	for {
		found, err := findMailpitMessage(ctx, client, recipient, otp)
		if err == nil && found {
			return nil
		}

		select {
		case <-ctx.Done():
			return fmt.Errorf("Mailpit inbox did not expose the delivered OTP email before the deadline")
		case <-ticker.C:
		}
	}
}

func findMailpitMessage(
	ctx context.Context,
	client *http.Client,
	recipient string,
	otp string,
) (bool, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://127.0.0.1:8025/api/v1/messages", nil)
	if err != nil {
		return false, err
	}
	response, err := client.Do(request)
	if err != nil {
		return false, err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return false, fmt.Errorf("Mailpit messages endpoint returned an unexpected status")
	}

	var inbox struct {
		Messages []struct {
			ID string `json:"ID"`
		} `json:"messages"`
	}
	if err := json.NewDecoder(response.Body).Decode(&inbox); err != nil {
		return false, err
	}

	for _, summary := range inbox.Messages {
		if summary.ID == "" {
			continue
		}
		found, err := mailpitMessageContains(ctx, client, summary.ID, recipient, otp)
		if err != nil {
			continue
		}
		if found {
			return true, nil
		}
	}

	return false, nil
}

func mailpitMessageContains(
	ctx context.Context,
	client *http.Client,
	messageID string,
	recipient string,
	otp string,
) (bool, error) {
	request, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		"http://127.0.0.1:8025/api/v1/message/"+messageID,
		nil,
	)
	if err != nil {
		return false, err
	}
	response, err := client.Do(request)
	if err != nil {
		return false, err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return false, fmt.Errorf("Mailpit message endpoint returned an unexpected status")
	}

	var message struct {
		Subject string `json:"Subject"`
		Text    string `json:"Text"`
		To      []struct {
			Address string `json:"Address"`
		} `json:"To"`
	}
	if err := json.NewDecoder(response.Body).Decode(&message); err != nil {
		return false, err
	}

	addressedToRecipient := false
	for _, address := range message.To {
		if address.Address == recipient {
			addressedToRecipient = true
			break
		}
	}

	return addressedToRecipient &&
		message.Subject == otpEmailSubject &&
		strings.Contains(message.Text, otp) &&
		strings.Contains(message.Text, "Super Admin authentication"), nil
}
