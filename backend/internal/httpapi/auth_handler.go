package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"mime"
	"net"
	"net/http"
	"net/netip"
	"strings"
	"time"

	"salesagent.local/backend/internal/platform/auth"
)

const (
	sessionCookieName           = "sales_agent_session"
	maxAuthBodyBytes            = 8 * 1024
	authServiceOperationTimeout = 10 * time.Second
)

type AuthenticationService interface {
	Login(
		ctx context.Context,
		email string,
		password string,
		requestingIP string,
	) (auth.AuthenticatedSession, string, error)
	ResolveSession(ctx context.Context, rawSessionToken string) (auth.AuthenticatedSession, error)
	Logout(ctx context.Context, rawSessionToken string, csrfToken string) error
}

type AuthHandlerOptions struct {
	ApplicationOrigin string
	CookieSecure      bool
	SessionTTL        time.Duration
	LocalDevelopment  bool
	// TrustedProxyCIDRs is the explicit network boundary within which a peer
	// may append trustworthy client-address information to X-Forwarded-For.
	TrustedProxyCIDRs []netip.Prefix
}

type AuthHandler struct {
	service           AuthenticationService
	applicationOrigin string
	cookieSecure      bool
	sessionTTL        time.Duration
	localDevelopment  bool
	trustedProxyCIDRs []netip.Prefix
}

func NewAuthHandler(service AuthenticationService, options AuthHandlerOptions) (*AuthHandler, error) {
	if service == nil {
		return nil, errors.New("authentication service is required")
	}
	if strings.TrimSpace(options.ApplicationOrigin) == "" {
		return nil, errors.New("application origin is required")
	}
	if options.SessionTTL <= 0 {
		return nil, errors.New("session TTL must be positive")
	}
	trustedProxyCIDRs := make([]netip.Prefix, len(options.TrustedProxyCIDRs))
	for index, prefix := range options.TrustedProxyCIDRs {
		if !prefix.IsValid() || prefix.Bits() == 0 || prefix.Addr().Is4In6() {
			return nil, errors.New("trusted proxy CIDRs must be valid and bounded")
		}
		trustedProxyCIDRs[index] = prefix.Masked()
	}

	return &AuthHandler{
		service:           service,
		applicationOrigin: options.ApplicationOrigin,
		cookieSecure:      options.CookieSecure,
		sessionTTL:        options.SessionTTL,
		localDevelopment:  options.LocalDevelopment,
		trustedProxyCIDRs: trustedProxyCIDRs,
	}, nil
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type superAdminResponse struct {
	Email       string `json:"email"`
	DisplayName string `json:"display_name"`
}

type authData struct {
	SuperAdmin       superAdminResponse `json:"super_admin"`
	CSRFToken        string             `json:"csrf_token"`
	LocalDevelopment bool               `json:"local_development"`
}

type authEnvelope struct {
	Data authData `json:"data"`
}

func (handler *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	setAuthNoStore(w)

	// Login can create an authenticated cookie, so it receives the same exact
	// Origin protection as authenticated mutations to prevent login CSRF.
	if !handler.validOrigin(r) {
		writeAPIError(
			w,
			r,
			http.StatusForbidden,
			"ORIGIN_VALIDATION_FAILED",
			"The request origin could not be verified.",
			nil,
		)
		return
	}

	var request loginRequest
	if err := decodeStrictJSON(w, r, &request); err != nil {
		writeAPIError(w, r, http.StatusBadRequest, "INVALID_REQUEST", "The sign-in request is invalid.", nil)
		return
	}

	fields := validateLoginRequest(request)
	if len(fields) > 0 {
		writeAPIError(w, r, http.StatusBadRequest, "INVALID_REQUEST", "Check the highlighted fields.", fields)
		return
	}

	operationCtx, cancelOperation := handler.serviceOperationContext(r.Context())
	defer cancelOperation()
	session, rawSessionToken, err := handler.service.Login(
		operationCtx,
		request.Email,
		request.Password,
		requestingIP(r, handler.trustedProxyCIDRs),
	)
	if err != nil {
		handler.writeAuthError(w, r, err)
		return
	}

	// The raw token crosses the application boundary only through an HttpOnly
	// cookie. It is never serialized in JSON or made available to JavaScript.
	handler.setSessionCookie(w, rawSessionToken, session.ExpiresAt)
	writeJSON(w, http.StatusOK, handler.response(session))
}

func (handler *AuthHandler) Session(w http.ResponseWriter, r *http.Request) {
	setAuthNoStore(w)

	cookie, err := r.Cookie(sessionCookieName)
	if err != nil {
		handler.writeAuthError(w, r, auth.ErrUnauthenticated)
		return
	}

	operationCtx, cancelOperation := handler.serviceOperationContext(r.Context())
	defer cancelOperation()
	session, err := handler.service.ResolveSession(operationCtx, cookie.Value)
	if err != nil {
		handler.writeAuthError(w, r, err)
		return
	}

	writeJSON(w, http.StatusOK, handler.response(session))
}

func (handler *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	setAuthNoStore(w)

	if !handler.validOrigin(r) {
		writeAPIError(
			w,
			r,
			http.StatusForbidden,
			"CSRF_VALIDATION_FAILED",
			"The security token could not be verified.",
			nil,
		)
		return
	}

	cookie, err := r.Cookie(sessionCookieName)
	if err != nil {
		handler.writeAuthError(w, r, auth.ErrUnauthenticated)
		return
	}

	csrfToken := r.Header.Get("X-CSRF-Token")
	if csrfToken == "" {
		handler.writeAuthError(w, r, auth.ErrInvalidCSRFToken)
		return
	}

	operationCtx, cancelOperation := handler.serviceOperationContext(r.Context())
	defer cancelOperation()
	if err := handler.service.Logout(operationCtx, cookie.Value, csrfToken); err != nil {
		handler.writeAuthError(w, r, err)
		return
	}

	// Logout revokes PostgreSQL state first. Clearing the browser cookie before
	// a successful revoke could leave a live server session the user can no
	// longer explicitly terminate.
	handler.clearSessionCookie(w)
	w.WriteHeader(http.StatusNoContent)
}

func (handler *AuthHandler) response(session auth.AuthenticatedSession) authEnvelope {
	return authEnvelope{Data: authData{
		SuperAdmin: superAdminResponse{
			Email:       session.SuperAdmin.Email,
			DisplayName: session.SuperAdmin.DisplayName,
		},
		CSRFToken:        session.CSRFToken,
		LocalDevelopment: handler.localDevelopment,
	}}
}

func (handler *AuthHandler) writeAuthError(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, auth.ErrInvalidCredentials):
		writeAPIError(w, r, http.StatusUnauthorized, "INVALID_CREDENTIALS", "Invalid email or password.", nil)
	case errors.Is(err, auth.ErrRateLimited):
		w.Header().Set("Retry-After", "900")
		writeAPIError(w, r, http.StatusTooManyRequests, "AUTHENTICATION_RATE_LIMITED", "Too many sign-in attempts. Try again later.", nil)
	case errors.Is(err, auth.ErrOTPRequired):
		writeAPIError(w, r, http.StatusPreconditionRequired, "OTP_REQUIRED", "Email verification is required to complete sign in.", nil)
	case errors.Is(err, auth.ErrUnauthenticated):
		writeAPIError(w, r, http.StatusUnauthorized, "UNAUTHENTICATED", "A valid Super Admin session is required.", nil)
	case errors.Is(err, auth.ErrInvalidCSRFToken):
		writeAPIError(w, r, http.StatusForbidden, "CSRF_VALIDATION_FAILED", "The security token could not be verified.", nil)
	default:
		// SQL, Redis, crypto, and topology details are collapsed into one safe
		// response. The raw error must never be serialized or logged here.
		writeAPIError(w, r, http.StatusServiceUnavailable, "AUTHENTICATION_UNAVAILABLE", "Authentication is temporarily unavailable.", nil)
	}
}

func (handler *AuthHandler) validOrigin(r *http.Request) bool {
	return r.Header.Get("Origin") == handler.applicationOrigin
}

func (handler *AuthHandler) serviceOperationContext(parent context.Context) (context.Context, context.CancelFunc) {
	// Authentication combines cryptographic work with Redis and PostgreSQL I/O.
	// A handler-owned deadline prevents a stalled dependency from retaining an
	// HTTP connection indefinitely while preserving caller cancellation.
	return context.WithTimeout(parent, authServiceOperationTimeout)
}

func (handler *AuthHandler) setSessionCookie(w http.ResponseWriter, value string, expiresAt time.Time) {
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    value,
		Path:     "/",
		Expires:  expiresAt,
		MaxAge:   int(handler.sessionTTL.Seconds()),
		HttpOnly: true,
		Secure:   handler.cookieSecure,
		SameSite: http.SameSiteStrictMode,
	})
}

func (handler *AuthHandler) clearSessionCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    "",
		Path:     "/",
		Expires:  time.Unix(1, 0).UTC(),
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   handler.cookieSecure,
		SameSite: http.SameSiteStrictMode,
	})
}

func validateLoginRequest(request loginRequest) []fieldError {
	var fields []fieldError
	if _, err := auth.NormalizeEmail(request.Email); err != nil {
		fields = append(fields, fieldError{Field: "email", Message: "Enter a valid email address."})
	}
	if err := auth.ValidateLoginPassword(request.Password); err != nil {
		fields = append(fields, fieldError{Field: "password", Message: "Enter your password."})
	}

	return fields
}

func decodeStrictJSON(w http.ResponseWriter, r *http.Request, destination any) error {
	mediaType, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil || mediaType != "application/json" {
		return errors.New("content type must be application/json")
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxAuthBodyBytes)
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(destination); err != nil {
		return err
	}

	var trailing any
	if err := decoder.Decode(&trailing); !errors.Is(err, io.EOF) {
		return errors.New("request body must contain one JSON object")
	}

	return nil
}

func requestingIP(r *http.Request, trustedProxyCIDRs []netip.Prefix) string {
	peer, ok := parseConnectionIP(r.RemoteAddr)
	if !ok {
		// A malformed connection identity must not make throttling disappear.
		// Returning it intact sends requests into a shared fail-closed bucket.
		return strings.TrimSpace(r.RemoteAddr)
	}
	peerText := peer.String()

	// Forwarded headers are attacker-controlled unless the process accepted the
	// connection from an explicitly configured reverse proxy.
	if !addressInPrefixes(peer, trustedProxyCIDRs) {
		return peerText
	}

	forwarded := strings.Join(r.Header.Values("X-Forwarded-For"), ",")
	if strings.TrimSpace(forwarded) == "" {
		return peerText
	}

	hops := strings.Split(forwarded, ",")
	var leftmostValid netip.Addr
	for index := len(hops) - 1; index >= 0; index-- {
		hop, err := netip.ParseAddr(strings.TrimSpace(hops[index]))
		if err != nil {
			// A malformed chain cannot be partially trusted. Falling back to the
			// known proxy peer preserves a bounded limiter bucket.
			return peerText
		}
		hop = canonicalIP(hop)
		leftmostValid = hop
		if !addressInPrefixes(hop, trustedProxyCIDRs) {
			return hop.String()
		}
	}

	// Every listed hop was itself trusted. Use the furthest address supplied by
	// the trusted proxy chain rather than collapsing all traffic into one peer.
	if leftmostValid.IsValid() {
		return leftmostValid.String()
	}
	return peerText
}

func parseConnectionIP(remoteAddress string) (netip.Addr, bool) {
	remoteAddress = strings.TrimSpace(remoteAddress)
	if addressPort, err := netip.ParseAddrPort(remoteAddress); err == nil {
		return canonicalIP(addressPort.Addr()), true
	}

	host, _, err := net.SplitHostPort(remoteAddress)
	if err != nil {
		return netip.Addr{}, false
	}
	address, err := netip.ParseAddr(host)
	if err != nil {
		return netip.Addr{}, false
	}
	return canonicalIP(address), true
}

func canonicalIP(address netip.Addr) netip.Addr {
	return address.Unmap().WithZone("")
}

func addressInPrefixes(address netip.Addr, prefixes []netip.Prefix) bool {
	for _, prefix := range prefixes {
		if prefix.Contains(address) {
			return true
		}
	}
	return false
}

func setAuthNoStore(w http.ResponseWriter) {
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Pragma", "no-cache")
}
