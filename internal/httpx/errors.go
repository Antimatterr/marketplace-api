package httpx

import (
	"encoding/json"
	"net/http"
)

type Code string

var (
	CodeInvalidId        Code = "invalid_id"
	CodeInternalError    Code = "internal_error"
	CodeNotFound         Code = "not_found"
	CodeRateLimit        Code = "rate_limited"
	CodeForbidden        Code = "forbidden"
	CodeUnauthenticated  Code = "unauthenticated"
	CodeFailedValidation Code = "validation_failed"
	CodeMalformedJSON    Code = "malformed_json"
)

type errorPayload struct {
	Code    Code   `json:"code"`
	Message string `json:"message"`
	Field   string `json:"field,omitempty"`
}

type errorEnvelope struct {
	Error errorPayload `json:"error"`
}

func Error(w http.ResponseWriter, status int, message string, code Code) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(errorEnvelope{
		Error: errorPayload{
			Code:    code,
			Message: message,
		},
	})
}

func ValidationError(w http.ResponseWriter, status int, message string, code Code, field string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(errorEnvelope{
		Error: errorPayload{
			Code:    code,
			Message: message,
			Field:   field,
		},
	})
}
