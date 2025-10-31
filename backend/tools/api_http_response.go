package tools

import (
	"compress/gzip"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

type contextKey string

var SESSION_KEY contextKey = "gloopert"

// Cancel Request and Respond with an API Error
func SendClientError(w http.ResponseWriter, r *http.Request, e APIError) {
	w.WriteHeader(e.Status)
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"code":%d,"message":%q}`, e.Code, e.Message)
}

// Cancel Request and Respond with a Generic Server Error
func SendServerError(w http.ResponseWriter, r *http.Request, err error) {
	LoggerHttp.Error(err.Error(), map[string]any{
		"url":     r.URL.String(),
		"headers": r.Header,
		"session": r.Context().Value(SESSION_KEY),
		"error":   err,
	})
	SendClientError(w, r, ERROR_GENERIC_SERVER)
}

// Cancel Request and Respond with a Validation Error
func SendFormError(w http.ResponseWriter, r *http.Request, verrs ...ValidationError) {
	w.WriteHeader(ERROR_BODY_INVALID.Status)
	SendJSON(w, r, map[string]any{
		"code":    ERROR_BODY_INVALID.Code,
		"message": ERROR_BODY_INVALID.Message,
		"errors":  verrs,
	})
}

// Encode and Compress Outgoing Body
func SendJSON(w http.ResponseWriter, r *http.Request, b any) error {

	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	if strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
		w.Header().Set("Content-Encoding", "gzip")
		z := gzip.NewWriter(w)
		defer z.Close()

		e := json.NewEncoder(z)
		e.SetEscapeHTML(false)
		return e.Encode(b)
	} else {
		e := json.NewEncoder(w)
		e.SetEscapeHTML(false)
		return e.Encode(b)
	}
}
