package httpapi

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/netip"
	"strings"
	"testing"
	"time"

	"salesagent.local/backend/internal/platform/auth"
)

const testApplicationOrigin = "http://127.0.0.1:3000"

type fakeAuthenticationService struct {
	loginSession  auth.AuthenticatedSession
	loginToken    string
	loginErr      error
	loginCalls    int
	loginEmail    string
	loginPassword string
	loginIP       string
	loginDeadline time.Time
	loginHasLimit bool

	resolvedSession auth.AuthenticatedSession
	resolveErr      error
	resolveToken    string
	resolveDeadline time.Time
	resolveHasLimit bool

	logoutErr      error
	logoutToken    string
	logoutCSRF     string
	logoutDeadline time.Time
	logoutHasLimit bool
}

func (service *fakeAuthenticationService) Login(
	ctx context.Context,
	email string,
	password string,
	ip string,
) (auth.AuthenticatedSession, string, error) {
	service.loginDeadline, service.loginHasLimit = ctx.Deadline()
	service.loginCalls++
	service.loginEmail = email
	service.loginPassword = password
	service.loginIP = ip
	return service.loginSession, service.loginToken, service.loginErr
}

func (service *fakeAuthenticationService) ResolveSession(
	ctx context.Context,
	token string,
) (auth.AuthenticatedSession, error) {
	service.resolveDeadline, service.resolveHasLimit = ctx.Deadline()
	service.resolveToken = token
	return service.resolvedSession, service.resolveErr
}

func (service *fakeAuthenticationService) Logout(
	ctx context.Context,
	token string,
	csrf string,
) error {
	service.logoutDeadline, service.logoutHasLimit = ctx.Deadline()
	service.logoutToken = token
	service.logoutCSRF = csrf
	return service.logoutErr
}

func TestLoginSuccessSetsSecurelyScopedHttpOnlyCookie(t *testing.T) {
	t.Parallel()

	expiresAt := time.Now().UTC().Add(8 * time.Hour)
	rawToken := testRawToken(3)
	service := &fakeAuthenticationService{
		loginToken: rawToken,
		loginSession: auth.AuthenticatedSession{
			SuperAdmin: auth.SuperAdmin{
				Email:       "admin@example.com",
				DisplayName: "Super Admin",
			},
			CSRFToken: "csrf-token",
			ExpiresAt: expiresAt,
		},
	}
	router := testAuthRouter(t, service, false)
	request := newJSONRequest(
		http.MethodPost,
		"/api/v1/auth/login",
		`{"email":"admin@example.com","password":"correct password"}`,
	)
	request.RemoteAddr = "192.0.2.25:4321"
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", response.Code, response.Body.String())
	}
	if service.loginEmail != "admin@example.com" || service.loginIP != "192.0.2.25" {
		t.Fatalf("Login arguments = email %q, IP %q", service.loginEmail, service.loginIP)
	}
	assertBoundedAuthDeadline(t, service.loginDeadline, service.loginHasLimit)
	if strings.Contains(response.Body.String(), rawToken) {
		t.Fatal("raw session token must not appear in JSON")
	}

	cookies := response.Result().Cookies()
	if len(cookies) != 1 {
		t.Fatalf("cookies = %d, want 1", len(cookies))
	}
	cookie := cookies[0]
	if cookie.Name != sessionCookieName || cookie.Value != rawToken {
		t.Fatalf("cookie = %#v", cookie)
	}
	if !cookie.HttpOnly || cookie.Secure || cookie.SameSite != http.SameSiteStrictMode || cookie.Path != "/" {
		t.Fatalf("local cookie attributes = %#v", cookie)
	}
	if cookie.Domain != "" {
		t.Fatalf("cookie Domain = %q, want empty", cookie.Domain)
	}
	assertAuthNoStore(t, response)
	if response.Header().Get("Access-Control-Allow-Origin") != "" {
		t.Fatal("auth API must not add broad CORS headers")
	}
}

func TestLoginUsesForwardedClientIPOnlyFromTrustedProxy(t *testing.T) {
	t.Parallel()

	service := &fakeAuthenticationService{loginErr: auth.ErrInvalidCredentials}
	handler, err := NewAuthHandler(service, AuthHandlerOptions{
		ApplicationOrigin: testApplicationOrigin,
		SessionTTL:        8 * time.Hour,
		TrustedProxyCIDRs: []netip.Prefix{netip.MustParsePrefix("127.0.0.1/32")},
	})
	if err != nil {
		t.Fatalf("NewAuthHandler() error = %v", err)
	}
	router := NewRouter(nil, handler)
	request := newJSONRequest(
		http.MethodPost,
		"/api/v1/auth/login",
		`{"email":"admin@example.com","password":"wrong password"}`,
	)
	request.RemoteAddr = "127.0.0.1:4321"
	request.Header.Set("X-Forwarded-For", "203.0.113.25")
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	if response.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, body = %s", response.Code, response.Body.String())
	}
	if service.loginIP != "203.0.113.25" {
		t.Fatalf("Login IP = %q, want forwarded client IP", service.loginIP)
	}
}

func TestRequestingIPTrustedProxyBoundary(t *testing.T) {
	t.Parallel()

	trusted := []netip.Prefix{
		netip.MustParsePrefix("10.0.0.0/8"),
		netip.MustParsePrefix("127.0.0.1/32"),
	}
	tests := []struct {
		name       string
		remoteAddr string
		forwarded  []string
		want       string
	}{
		{
			name:       "untrusted peer cannot spoof forwarded address",
			remoteAddr: "192.0.2.20:4000",
			forwarded:  []string{"203.0.113.99"},
			want:       "192.0.2.20",
		},
		{
			name:       "trusted peer supplies client address",
			remoteAddr: "127.0.0.1:4000",
			forwarded:  []string{"203.0.113.25"},
			want:       "203.0.113.25",
		},
		{
			name:       "rightmost untrusted hop defeats left-side spoof",
			remoteAddr: "10.0.0.3:4000",
			forwarded:  []string{"198.51.100.99, 203.0.113.25, 10.0.0.2"},
			want:       "203.0.113.25",
		},
		{
			name:       "multiple forwarded header lines are one chain",
			remoteAddr: "10.0.0.3:4000",
			forwarded:  []string{"198.51.100.99, 203.0.113.25", "10.0.0.2"},
			want:       "203.0.113.25",
		},
		{
			name:       "malformed chain falls back to known peer",
			remoteAddr: "127.0.0.1:4000",
			forwarded:  []string{"not-an-ip"},
			want:       "127.0.0.1",
		},
		{
			name:       "missing chain falls back to known peer",
			remoteAddr: "127.0.0.1:4000",
			want:       "127.0.0.1",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			request := httptest.NewRequest(http.MethodPost, "/", nil)
			request.RemoteAddr = test.remoteAddr
			for _, value := range test.forwarded {
				request.Header.Add("X-Forwarded-For", value)
			}

			if got := requestingIP(request, trusted); got != test.want {
				t.Fatalf("requestingIP() = %q, want %q", got, test.want)
			}
		})
	}
}

func TestProductionCookieUsesSecure(t *testing.T) {
	t.Parallel()

	service := &fakeAuthenticationService{
		loginToken: testRawToken(4),
		loginSession: auth.AuthenticatedSession{
			SuperAdmin: auth.SuperAdmin{Email: "admin@example.com", DisplayName: "Super Admin"},
			CSRFToken:  "csrf-token",
			ExpiresAt:  time.Now().Add(time.Hour),
		},
	}
	router := testAuthRouter(t, service, true)
	request := newJSONRequest(http.MethodPost, "/api/v1/auth/login", `{"email":"admin@example.com","password":"correct password"}`)
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", response.Code, response.Body.String())
	}
	if cookies := response.Result().Cookies(); len(cookies) != 1 || !cookies[0].Secure {
		t.Fatalf("production cookies = %#v", cookies)
	}
}

func TestLoginRejectsUnknownOversizedAndInvalidInput(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		body string
	}{
		{name: "unknown privileged field", body: `{"email":"admin@example.com","password":"password","is_active":true}`},
		{name: "unknown ID field", body: `{"email":"admin@example.com","password":"password","admin_id":"00000000-0000-0000-0000-000000000001"}`},
		{name: "malformed JSON", body: `{"email":`},
		{name: "multiple values", body: `{"email":"a@example.com","password":"password"}{}`},
		{name: "invalid email", body: `{"email":"not-an-email","password":"password"}`},
		{name: "empty password", body: `{"email":"admin@example.com","password":""}`},
		{name: "oversized", body: `{"email":"admin@example.com","password":"` + strings.Repeat("a", maxAuthBodyBytes) + `"}`},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			service := &fakeAuthenticationService{}
			router := testAuthRouter(t, service, false)
			request := newJSONRequest(http.MethodPost, "/api/v1/auth/login", test.body)
			response := httptest.NewRecorder()

			router.ServeHTTP(response, request)

			if response.Code != http.StatusBadRequest {
				t.Fatalf("status = %d, body = %s", response.Code, response.Body.String())
			}
			if service.loginCalls != 0 {
				t.Fatal("invalid DTO must not reach authentication service")
			}
			assertErrorCode(t, response, "INVALID_REQUEST")
			assertAuthNoStore(t, response)
		})
	}
}

func TestLoginPassesPasswordAsData(t *testing.T) {
	t.Parallel()

	service := &fakeAuthenticationService{loginErr: auth.ErrInvalidCredentials}
	router := testAuthRouter(t, service, false)
	injectionText := `x' OR '1'='1`
	request := newJSONRequest(
		http.MethodPost,
		"/api/v1/auth/login",
		`{"email":"admin@example.com","password":"x' OR '1'='1"}`,
	)
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	if response.Code != http.StatusUnauthorized || service.loginPassword != injectionText {
		t.Fatalf("status = %d, passed password = %q", response.Code, service.loginPassword)
	}
	assertErrorCode(t, response, "INVALID_CREDENTIALS")
}

func TestLoginMapsExpectedSafeErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		err        error
		wantStatus int
		wantCode   string
	}{
		{name: "invalid credentials", err: auth.ErrInvalidCredentials, wantStatus: http.StatusUnauthorized, wantCode: "INVALID_CREDENTIALS"},
		{name: "rate limited", err: auth.ErrRateLimited, wantStatus: http.StatusTooManyRequests, wantCode: "AUTHENTICATION_RATE_LIMITED"},
		{name: "OTP required", err: auth.ErrOTPRequired, wantStatus: http.StatusPreconditionRequired, wantCode: "OTP_REQUIRED"},
		{name: "internal unavailable", err: errors.New("SELECT password_hash FROM secret_internal_table"), wantStatus: http.StatusServiceUnavailable, wantCode: "AUTHENTICATION_UNAVAILABLE"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			service := &fakeAuthenticationService{loginErr: test.err}
			router := testAuthRouter(t, service, false)
			request := newJSONRequest(http.MethodPost, "/api/v1/auth/login", `{"email":"admin@example.com","password":"password"}`)
			response := httptest.NewRecorder()

			router.ServeHTTP(response, request)

			if response.Code != test.wantStatus {
				t.Fatalf("status = %d, want %d", response.Code, test.wantStatus)
			}
			if cookies := response.Result().Cookies(); len(cookies) != 0 {
				t.Fatalf("failed login set cookies: %#v", cookies)
			}
			assertErrorCode(t, response, test.wantCode)
			if strings.Contains(response.Body.String(), "SELECT") || strings.Contains(response.Body.String(), "secret_internal_table") {
				t.Fatalf("internal error leaked: %s", response.Body.String())
			}
			assertAuthNoStore(t, response)
		})
	}
}

func TestLoginRequiresExactApplicationOrigin(t *testing.T) {
	t.Parallel()

	service := &fakeAuthenticationService{}
	router := testAuthRouter(t, service, false)
	request := newJSONRequest(http.MethodPost, "/api/v1/auth/login", `{"email":"admin@example.com","password":"password"}`)
	request.Header.Set("Origin", "https://attacker.example")
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	if response.Code != http.StatusForbidden || service.loginCalls != 0 {
		t.Fatalf("status = %d, Login calls = %d", response.Code, service.loginCalls)
	}
	assertErrorCode(t, response, "ORIGIN_VALIDATION_FAILED")
}

func TestSessionRequiresCookieAndReturnsRealServiceIdentity(t *testing.T) {
	t.Parallel()

	t.Run("missing session", func(t *testing.T) {
		t.Parallel()
		service := &fakeAuthenticationService{}
		router := testAuthRouter(t, service, false)
		request := httptest.NewRequest(http.MethodGet, "/api/v1/auth/session", nil)
		response := httptest.NewRecorder()

		router.ServeHTTP(response, request)

		if response.Code != http.StatusUnauthorized {
			t.Fatalf("status = %d", response.Code)
		}
		assertErrorCode(t, response, "UNAUTHENTICATED")
		assertAuthNoStore(t, response)
	})

	t.Run("authenticated identity", func(t *testing.T) {
		t.Parallel()
		service := &fakeAuthenticationService{resolvedSession: auth.AuthenticatedSession{
			SuperAdmin: auth.SuperAdmin{Email: "real-admin@example.com", DisplayName: "Real Admin"},
			CSRFToken:  "csrf-token",
		}}
		router := testAuthRouter(t, service, false)
		request := httptest.NewRequest(http.MethodGet, "/api/v1/auth/session", nil)
		request.AddCookie(&http.Cookie{Name: sessionCookieName, Value: testRawToken(8)})
		response := httptest.NewRecorder()

		router.ServeHTTP(response, request)

		if response.Code != http.StatusOK {
			t.Fatalf("status = %d, body = %s", response.Code, response.Body.String())
		}
		if !strings.Contains(response.Body.String(), "real-admin@example.com") || !strings.Contains(response.Body.String(), "csrf-token") {
			t.Fatalf("body = %s", response.Body.String())
		}
		if service.resolveToken == "" {
			t.Fatal("session cookie was not resolved through auth service")
		}
		assertBoundedAuthDeadline(t, service.resolveDeadline, service.resolveHasLimit)
		assertAuthNoStore(t, response)
	})
}

func TestLogoutRequiresOriginAndCSRFBeforeRevocation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		origin   string
		csrf     string
		wantCode string
	}{
		{name: "wrong origin", origin: "https://attacker.example", csrf: "csrf-token", wantCode: "CSRF_VALIDATION_FAILED"},
		{name: "missing CSRF", origin: testApplicationOrigin, wantCode: "CSRF_VALIDATION_FAILED"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			service := &fakeAuthenticationService{}
			router := testAuthRouter(t, service, false)
			request := httptest.NewRequest(http.MethodPost, "/api/v1/auth/logout", nil)
			request.Header.Set("Origin", test.origin)
			if test.csrf != "" {
				request.Header.Set("X-CSRF-Token", test.csrf)
			}
			request.AddCookie(&http.Cookie{Name: sessionCookieName, Value: testRawToken(5)})
			response := httptest.NewRecorder()

			router.ServeHTTP(response, request)

			if response.Code != http.StatusForbidden {
				t.Fatalf("status = %d, body = %s", response.Code, response.Body.String())
			}
			if service.logoutToken != "" {
				t.Fatal("request without both CSRF controls must not reach revocation")
			}
			assertErrorCode(t, response, test.wantCode)
		})
	}
}

func TestLogoutRevokesThenClearsCookie(t *testing.T) {
	t.Parallel()

	service := &fakeAuthenticationService{}
	router := testAuthRouter(t, service, false)
	rawToken := testRawToken(6)
	request := httptest.NewRequest(http.MethodPost, "/api/v1/auth/logout", nil)
	request.Header.Set("Origin", testApplicationOrigin)
	request.Header.Set("X-CSRF-Token", "csrf-token")
	request.AddCookie(&http.Cookie{Name: sessionCookieName, Value: rawToken})
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	if response.Code != http.StatusNoContent {
		t.Fatalf("status = %d, body = %s", response.Code, response.Body.String())
	}
	if service.logoutToken != rawToken || service.logoutCSRF != "csrf-token" {
		t.Fatalf("Logout arguments = token %q, CSRF %q", service.logoutToken, service.logoutCSRF)
	}
	assertBoundedAuthDeadline(t, service.logoutDeadline, service.logoutHasLimit)
	cookies := response.Result().Cookies()
	if len(cookies) != 1 || cookies[0].MaxAge != -1 || cookies[0].Value != "" {
		t.Fatalf("clearing cookies = %#v", cookies)
	}
	assertAuthNoStore(t, response)
}

func testAuthRouter(
	t *testing.T,
	service AuthenticationService,
	secure bool,
) http.Handler {
	t.Helper()

	handler, err := NewAuthHandler(service, AuthHandlerOptions{
		ApplicationOrigin: testApplicationOrigin,
		CookieSecure:      secure,
		SessionTTL:        8 * time.Hour,
		LocalDevelopment:  !secure,
	})
	if err != nil {
		t.Fatalf("NewAuthHandler() error = %v", err)
	}

	return NewRouter(nil, handler)
}

func newJSONRequest(method string, target string, body string) *http.Request {
	request := httptest.NewRequest(method, target, strings.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Origin", testApplicationOrigin)
	return request
}

func assertAuthNoStore(t *testing.T, response *httptest.ResponseRecorder) {
	t.Helper()
	if got := response.Header().Get("Cache-Control"); got != "no-store" {
		t.Fatalf("Cache-Control = %q, want no-store", got)
	}
}

func assertBoundedAuthDeadline(t *testing.T, deadline time.Time, hasDeadline bool) {
	t.Helper()
	if !hasDeadline {
		t.Fatal("authentication service context must have a deadline")
	}
	remaining := time.Until(deadline)
	if remaining <= 0 || remaining > authServiceOperationTimeout {
		t.Fatalf(
			"authentication service deadline has %s remaining, want within (0, %s]",
			remaining,
			authServiceOperationTimeout,
		)
	}
}

func assertErrorCode(t *testing.T, response *httptest.ResponseRecorder, want string) {
	t.Helper()
	var envelope errorEnvelope
	if err := json.Unmarshal(response.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("decode error response: %v; body = %s", err, response.Body.String())
	}
	if envelope.Error.Code != want {
		t.Fatalf("error code = %q, want %q; body = %s", envelope.Error.Code, want, response.Body.String())
	}
	if envelope.Error.CorrelationID == "" || envelope.Error.FieldErrors == nil {
		t.Fatalf("incomplete error envelope = %#v", envelope)
	}
}

func testRawToken(value byte) string {
	return base64.RawURLEncoding.EncodeToString([]byte(strings.Repeat(string([]byte{value}), 32)))
}
