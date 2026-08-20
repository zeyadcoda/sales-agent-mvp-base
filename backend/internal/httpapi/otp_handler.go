package httpapi

import (
	"net/http"
	"time"

	"salesagent.local/backend/internal/platform/auth"
)

type otpChallengeRequest struct {
	ChallengeID string `json:"challenge_id"`
}

type otpVerifyRequest struct {
	ChallengeID string `json:"challenge_id"`
	OTP         string `json:"otp"`
}

type otpChallengeResponse struct {
	ChallengeID       string                 `json:"challenge_id"`
	ExpiresAt         time.Time              `json:"expires_at"`
	ResendAvailableAt time.Time              `json:"resend_available_at"`
	DestinationHint   string                 `json:"destination_hint"`
	State             auth.OTPChallengeState `json:"state"`
}

type otpRequiredData struct {
	AuthenticationState string               `json:"authentication_state"`
	Challenge           otpChallengeResponse `json:"challenge"`
}

type otpRequiredEnvelope struct {
	Data otpRequiredData `json:"data"`
}

type otpChallengeEnvelope struct {
	Data otpChallengeResponse `json:"data"`
}

func (handler *AuthHandler) VerifyOTP(w http.ResponseWriter, r *http.Request) {
	setAuthNoStore(w)
	if !handler.requireOTPOrigin(w, r) {
		return
	}

	var request otpVerifyRequest
	if err := decodeStrictJSON(w, r, &request); err != nil ||
		!auth.ValidOTPChallengeID(request.ChallengeID) ||
		!auth.ValidOTP(request.OTP) {
		writeAPIError(w, r, http.StatusBadRequest, "INVALID_REQUEST", "The verification request is invalid.", nil)
		return
	}

	operationCtx, cancelOperation := handler.serviceOperationContext(r.Context())
	defer cancelOperation()
	authenticated, err := handler.service.VerifyOTP(
		operationCtx,
		request.ChallengeID,
		request.OTP,
		requestingIP(r, handler.trustedProxyCIDRs),
	)
	if err != nil {
		handler.writeAuthError(w, r, err)
		return
	}

	// OTP verification creates a new session secret independent from the
	// challenge. Only the HttpOnly cookie receives its raw value.
	handler.setSessionCookie(w, authenticated.RawSessionToken, authenticated.Session.ExpiresAt)
	writeJSON(w, http.StatusOK, handler.response(authenticated.Session))
}

func (handler *AuthHandler) ResendOTP(w http.ResponseWriter, r *http.Request) {
	setAuthNoStore(w)
	if !handler.requireOTPOrigin(w, r) {
		return
	}

	request, ok := handler.decodeOTPChallengeRequest(w, r)
	if !ok {
		return
	}

	operationCtx, cancelOperation := handler.serviceOperationContext(r.Context())
	defer cancelOperation()
	challenge, err := handler.service.ResendOTP(
		operationCtx,
		request.ChallengeID,
		requestingIP(r, handler.trustedProxyCIDRs),
	)
	if err != nil {
		handler.writeAuthError(w, r, err)
		return
	}

	writeJSON(w, http.StatusOK, otpChallengeEnvelope{
		Data: pendingOTPChallengeResponse(challenge),
	})
}

func (handler *AuthHandler) OTPChallengeStatus(w http.ResponseWriter, r *http.Request) {
	setAuthNoStore(w)
	if !handler.requireOTPOrigin(w, r) {
		return
	}

	request, ok := handler.decodeOTPChallengeRequest(w, r)
	if !ok {
		return
	}

	operationCtx, cancelOperation := handler.serviceOperationContext(r.Context())
	defer cancelOperation()
	challenge, err := handler.service.GetOTPChallengeStatus(operationCtx, request.ChallengeID)
	if err != nil {
		handler.writeAuthError(w, r, err)
		return
	}

	response := pendingOTPChallengeResponse(challenge.PendingChallenge)
	response.State = challenge.State
	writeJSON(w, http.StatusOK, otpChallengeEnvelope{Data: response})
}

func (handler *AuthHandler) decodeOTPChallengeRequest(
	w http.ResponseWriter,
	r *http.Request,
) (otpChallengeRequest, bool) {
	var request otpChallengeRequest
	if err := decodeStrictJSON(w, r, &request); err != nil ||
		!auth.ValidOTPChallengeID(request.ChallengeID) {
		writeAPIError(w, r, http.StatusBadRequest, "INVALID_REQUEST", "The verification request is invalid.", nil)
		return otpChallengeRequest{}, false
	}

	return request, true
}

func (handler *AuthHandler) requireOTPOrigin(w http.ResponseWriter, r *http.Request) bool {
	if handler.validOrigin(r) {
		return true
	}

	writeAPIError(
		w,
		r,
		http.StatusForbidden,
		"ORIGIN_VALIDATION_FAILED",
		"The request origin could not be verified.",
		nil,
	)
	return false
}

func otpRequiredResponse(challenge auth.PendingChallenge) otpRequiredEnvelope {
	return otpRequiredEnvelope{Data: otpRequiredData{
		AuthenticationState: "OTP_REQUIRED",
		Challenge:           pendingOTPChallengeResponse(challenge),
	}}
}

func pendingOTPChallengeResponse(challenge auth.PendingChallenge) otpChallengeResponse {
	return otpChallengeResponse{
		ChallengeID:       challenge.ID,
		ExpiresAt:         challenge.ExpiresAt,
		ResendAvailableAt: challenge.ResendAvailableAt,
		DestinationHint:   challenge.DestinationHint,
		State:             auth.OTPChallengePending,
	}
}
