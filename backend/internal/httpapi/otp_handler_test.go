package httpapi

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"salesagent.local/backend/internal/platform/auth"
)

func TestPasswordLoginReturnsPendingOTPChallengeWithoutSession(t *testing.T) {
	t.Parallel()

	challenge := testPendingChallenge()
	service := &fakeAuthenticationService{loginChallenge: &challenge}
	router := testAuthRouter(t, service, false)
	request := newJSONRequest(
		http.MethodPost,
		"/api/v1/auth/login",
		`{"email":"admin@example.com","password":"correct password"}`,
	)
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	if response.Code != http.StatusAccepted {
		t.Fatalf("status = %d, body = %s", response.Code, response.Body.String())
	}
	if cookies := response.Result().Cookies(); len(cookies) != 0 {
		t.Fatalf("password-only login set cookies: %#v", cookies)
	}
	if strings.Contains(response.Body.String(), "correct password") ||
		strings.Contains(response.Body.String(), "otp_hash") {
		t.Fatalf("sensitive login material leaked: %s", response.Body.String())
	}

	var envelope otpRequiredEnvelope
	if err := json.Unmarshal(response.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("decode challenge response: %v", err)
	}
	if envelope.Data.AuthenticationState != "OTP_REQUIRED" ||
		envelope.Data.Challenge.ChallengeID != challenge.ID ||
		envelope.Data.Challenge.DestinationHint != challenge.DestinationHint ||
		envelope.Data.Challenge.State != auth.OTPChallengePending {
		t.Fatalf("challenge response = %#v", envelope)
	}
	assertAuthNoStore(t, response)
}

func TestVerifyOTPSetsExistingSecureSessionCookie(t *testing.T) {
	t.Parallel()

	rawToken := testRawToken(31)
	expiresAt := time.Now().UTC().Add(8 * time.Hour)
	service := &fakeAuthenticationService{verifyLogin: auth.AuthenticatedLogin{
		Session: auth.AuthenticatedSession{
			SuperAdmin: auth.SuperAdmin{
				Email:       "admin@example.com",
				DisplayName: "Super Admin",
			},
			CSRFToken: "csrf-token",
			ExpiresAt: expiresAt,
		},
		RawSessionToken: rawToken,
	}}
	router := testAuthRouter(t, service, true)
	request := newJSONRequest(
		http.MethodPost,
		"/api/v1/auth/otp/verify",
		`{"challenge_id":"`+testRawToken(30)+`","otp":"001284"}`,
	)
	request.RemoteAddr = "192.0.2.50:4321"
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", response.Code, response.Body.String())
	}
	if service.verifyChallengeID != testRawToken(30) ||
		service.verifyOTP != "001284" ||
		service.verifyIP != "192.0.2.50" {
		t.Fatalf(
			"VerifyOTP arguments = (%q, %q, %q)",
			service.verifyChallengeID,
			service.verifyOTP,
			service.verifyIP,
		)
	}
	assertBoundedAuthDeadline(t, service.verifyDeadline, service.verifyHasLimit)
	if strings.Contains(response.Body.String(), rawToken) || strings.Contains(response.Body.String(), "001284") {
		t.Fatalf("OTP or session token leaked into JSON: %s", response.Body.String())
	}

	cookies := response.Result().Cookies()
	if len(cookies) != 1 {
		t.Fatalf("cookies = %#v, want one", cookies)
	}
	cookie := cookies[0]
	if cookie.Name != sessionCookieName || cookie.Value != rawToken ||
		!cookie.HttpOnly || !cookie.Secure || cookie.SameSite != http.SameSiteStrictMode ||
		cookie.Path != "/" || cookie.Domain != "" {
		t.Fatalf("session cookie = %#v", cookie)
	}
	assertAuthNoStore(t, response)
}

func TestOTPRequestsRejectInvalidBodiesBeforeService(t *testing.T) {
	t.Parallel()

	validID := testRawToken(32)
	tests := []struct {
		name   string
		target string
		body   string
	}{
		{name: "verify malformed JSON", target: "/api/v1/auth/otp/verify", body: `{"challenge_id":`},
		{name: "verify unknown field", target: "/api/v1/auth/otp/verify", body: `{"challenge_id":"` + validID + `","otp":"123456","admin_id":"other"}`},
		{name: "verify malformed challenge", target: "/api/v1/auth/otp/verify", body: `{"challenge_id":"predictable","otp":"123456"}`},
		{name: "verify short OTP", target: "/api/v1/auth/otp/verify", body: `{"challenge_id":"` + validID + `","otp":"12345"}`},
		{name: "verify nonnumeric OTP", target: "/api/v1/auth/otp/verify", body: `{"challenge_id":"` + validID + `","otp":"12a456"}`},
		{name: "verify oversized body", target: "/api/v1/auth/otp/verify", body: `{"challenge_id":"` + validID + `","otp":"123456","padding":"` + strings.Repeat("a", maxAuthBodyBytes) + `"}`},
		{name: "resend unknown field", target: "/api/v1/auth/otp/resend", body: `{"challenge_id":"` + validID + `","super_admin_id":"other"}`},
		{name: "status malformed challenge", target: "/api/v1/auth/otp/status", body: `{"challenge_id":"not-random"}`},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			service := &fakeAuthenticationService{}
			request := newJSONRequest(http.MethodPost, test.target, test.body)
			response := httptest.NewRecorder()

			testAuthRouter(t, service, false).ServeHTTP(response, request)

			if response.Code != http.StatusBadRequest {
				t.Fatalf("status = %d, body = %s", response.Code, response.Body.String())
			}
			if service.verifyCalls != 0 || service.resendCalls != 0 || service.statusCalls != 0 {
				t.Fatal("invalid OTP DTO reached authentication service")
			}
			assertErrorCode(t, response, "INVALID_REQUEST")
			assertAuthNoStore(t, response)
		})
	}
}

func TestOTPEndpointsRequireExactOrigin(t *testing.T) {
	t.Parallel()

	validID := testRawToken(33)
	tests := []struct {
		target string
		body   string
	}{
		{target: "/api/v1/auth/otp/verify", body: `{"challenge_id":"` + validID + `","otp":"123456"}`},
		{target: "/api/v1/auth/otp/resend", body: `{"challenge_id":"` + validID + `"}`},
		{target: "/api/v1/auth/otp/status", body: `{"challenge_id":"` + validID + `"}`},
	}

	for _, test := range tests {
		service := &fakeAuthenticationService{}
		request := newJSONRequest(http.MethodPost, test.target, test.body)
		request.Header.Set("Origin", "https://attacker.example")
		response := httptest.NewRecorder()

		testAuthRouter(t, service, false).ServeHTTP(response, request)

		if response.Code != http.StatusForbidden {
			t.Fatalf("%s status = %d, body = %s", test.target, response.Code, response.Body.String())
		}
		if service.verifyCalls != 0 || service.resendCalls != 0 || service.statusCalls != 0 {
			t.Fatalf("%s reached authentication service", test.target)
		}
		assertErrorCode(t, response, "ORIGIN_VALIDATION_FAILED")
	}
}

func TestResendAndStatusReturnOnlySafeChallengeContext(t *testing.T) {
	t.Parallel()

	pending := testPendingChallenge()
	service := &fakeAuthenticationService{
		resendChallenge: pending,
		statusChallenge: auth.ChallengeState{
			PendingChallenge: pending,
			State:            auth.OTPChallengeExpired,
		},
	}
	router := testAuthRouter(t, service, false)

	resendRequest := newJSONRequest(
		http.MethodPost,
		"/api/v1/auth/otp/resend",
		`{"challenge_id":"`+pending.ID+`"}`,
	)
	resendRequest.RemoteAddr = "198.51.100.8:1234"
	resendResponse := httptest.NewRecorder()
	router.ServeHTTP(resendResponse, resendRequest)
	if resendResponse.Code != http.StatusOK {
		t.Fatalf("resend status = %d, body = %s", resendResponse.Code, resendResponse.Body.String())
	}
	if service.resendChallengeID != pending.ID || service.resendIP != "198.51.100.8" {
		t.Fatalf("resend arguments = (%q, %q)", service.resendChallengeID, service.resendIP)
	}
	assertBoundedAuthDeadline(t, service.resendDeadline, service.resendHasLimit)
	assertSafeChallengeEnvelope(t, resendResponse, auth.OTPChallengePending, pending)

	statusRequest := newJSONRequest(
		http.MethodPost,
		"/api/v1/auth/otp/status",
		`{"challenge_id":"`+pending.ID+`"}`,
	)
	statusResponse := httptest.NewRecorder()
	router.ServeHTTP(statusResponse, statusRequest)
	if statusResponse.Code != http.StatusOK {
		t.Fatalf("status endpoint = %d, body = %s", statusResponse.Code, statusResponse.Body.String())
	}
	if service.statusChallengeID != pending.ID {
		t.Fatalf("status challenge ID = %q", service.statusChallengeID)
	}
	assertBoundedAuthDeadline(t, service.statusDeadline, service.statusHasLimit)
	assertSafeChallengeEnvelope(t, statusResponse, auth.OTPChallengeExpired, pending)
}

func TestOTPApplicationErrorsHaveControlledResponses(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		err        error
		wantStatus int
		wantCode   string
	}{
		{name: "invalid", err: auth.ErrOTPInvalid, wantStatus: http.StatusUnauthorized, wantCode: "AUTH_OTP_INVALID"},
		{name: "unknown", err: auth.ErrOTPChallengeNotFound, wantStatus: http.StatusUnauthorized, wantCode: "AUTH_OTP_INVALID"},
		{name: "expired", err: auth.ErrOTPExpired, wantStatus: http.StatusGone, wantCode: "AUTH_OTP_EXPIRED"},
		{name: "attempts", err: auth.ErrOTPAttemptsExceeded, wantStatus: http.StatusLocked, wantCode: "AUTH_OTP_ATTEMPTS_EXCEEDED"},
		{name: "invalidated", err: auth.ErrOTPInvalidated, wantStatus: http.StatusConflict, wantCode: "AUTH_OTP_INVALIDATED"},
		{name: "consumed", err: auth.ErrOTPConsumed, wantStatus: http.StatusConflict, wantCode: "AUTH_OTP_CONSUMED"},
		{name: "too early", err: auth.ErrOTPResendTooEarly, wantStatus: http.StatusTooManyRequests, wantCode: "AUTH_OTP_RESEND_TOO_EARLY"},
		{name: "verify throttled", err: auth.ErrOTPVerifyRateLimited, wantStatus: http.StatusTooManyRequests, wantCode: "AUTH_OTP_RATE_LIMITED"},
		{name: "Redis unavailable", err: auth.ErrAuthenticationUnavailable, wantStatus: http.StatusServiceUnavailable, wantCode: "AUTHENTICATION_UNAVAILABLE"},
		{name: "raw internal error", err: errors.New("smtp://user:secret@internal.example"), wantStatus: http.StatusServiceUnavailable, wantCode: "AUTHENTICATION_UNAVAILABLE"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			service := &fakeAuthenticationService{verifyErr: test.err}
			request := newJSONRequest(
				http.MethodPost,
				"/api/v1/auth/otp/verify",
				`{"challenge_id":"`+testRawToken(34)+`","otp":"123456"}`,
			)
			response := httptest.NewRecorder()

			testAuthRouter(t, service, false).ServeHTTP(response, request)

			if response.Code != test.wantStatus {
				t.Fatalf("status = %d, want %d; body = %s", response.Code, test.wantStatus, response.Body.String())
			}
			if cookies := response.Result().Cookies(); len(cookies) != 0 {
				t.Fatalf("failed OTP verification set cookies: %#v", cookies)
			}
			assertErrorCode(t, response, test.wantCode)
			if strings.Contains(response.Body.String(), "smtp") || strings.Contains(response.Body.String(), "secret") {
				t.Fatalf("internal provider error leaked: %s", response.Body.String())
			}
			assertAuthNoStore(t, response)
		})
	}
}

func assertSafeChallengeEnvelope(
	t *testing.T,
	response *httptest.ResponseRecorder,
	wantState auth.OTPChallengeState,
	want auth.PendingChallenge,
) {
	t.Helper()

	var envelope otpChallengeEnvelope
	if err := json.Unmarshal(response.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("decode challenge envelope: %v", err)
	}
	if envelope.Data.ChallengeID != want.ID ||
		envelope.Data.State != wantState ||
		envelope.Data.DestinationHint != want.DestinationHint ||
		!envelope.Data.ExpiresAt.Equal(want.ExpiresAt) ||
		!envelope.Data.ResendAvailableAt.Equal(want.ResendAvailableAt) {
		t.Fatalf("challenge envelope = %#v", envelope)
	}
	for _, forbidden := range []string{"otp_hash", "super_admin_id", "failed_attempts", "delivery_version"} {
		if strings.Contains(response.Body.String(), forbidden) {
			t.Fatalf("challenge response exposed %q: %s", forbidden, response.Body.String())
		}
	}
	assertAuthNoStore(t, response)
}

func testPendingChallenge() auth.PendingChallenge {
	now := time.Now().UTC().Truncate(time.Second)
	return auth.PendingChallenge{
		ID:                testRawToken(29),
		ExpiresAt:         now.Add(10 * time.Minute),
		ResendAvailableAt: now.Add(time.Minute),
		DestinationHint:   "a***@example.com",
	}
}
