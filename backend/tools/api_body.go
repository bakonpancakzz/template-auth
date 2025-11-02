package tools

import (
	"compress/gzip"
	"encoding/json"
	"net/http"
	"reflect"
	"strconv"
	"strings"
)

// Use Built-in Validator against Request Body
func ValidateBody(w http.ResponseWriter, r *http.Request, b any) bool {
	verr, err := ValidateStruct(b)
	if err != nil {
		SendServerError(w, r, err)
		return false
	}
	if len(verr) > 0 {
		SendFormError(w, r, verr...)
		return false
	}
	return true
}

// Decode and Validate Incoming JSON Request
func ValidateJSON(w http.ResponseWriter, r *http.Request, b any) bool {
	defer r.Body.Close()
	if !strings.Contains(r.Header.Get("Content-Type"), "application/json") {
		SendClientError(w, r, ERROR_BODY_MALFORMED)
		return false
	}

	// Incoming Decoder
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(b); err != nil {
		SendClientError(w, r, ERROR_BODY_MALFORMED)
		return false
	}

	return ValidateBody(w, r, b)
}

// Decode and Validate Incoming Query Parameters
func ValidateQuery(w http.ResponseWriter, r *http.Request, b any) bool {
	defer r.Body.Close()
	if !strings.Contains(r.Header.Get("Content-Type"), "application/x-www-form-urlencoded") {
		SendClientError(w, r, ERROR_BODY_MALFORMED)
		return false
	}

	// Fill struct using Query values
	query := r.URL.Query()
	ptrValue := reflect.ValueOf(b)
	if ptrValue.Kind() != reflect.Ptr || ptrValue.IsNil() {
		panic("destination must be a non-nil pointer")
	}
	structValue := ptrValue.Elem()
	structType := structValue.Type()

	for i := 0; i < structType.NumField(); i++ {
		field := structType.Field(i)
		fieldTag := field.Tag.Get("query")
		if fieldTag == "" {
			continue
		}
		fieldValue := structValue.Field(i)

		if val := query.Get(fieldTag); val != "" && fieldValue.CanSet() {
			switch fieldValue.Kind() {
			case reflect.String:
				fieldValue.SetString(val)
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				if n, err := strconv.ParseInt(val, 10, 64); err == nil {
					fieldValue.SetInt(n)
				}
			case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
				if n, err := strconv.ParseUint(val, 10, 64); err == nil {
					fieldValue.SetUint(n)
				}
			case reflect.Bool:
				if b, err := strconv.ParseBool(val); err == nil {
					fieldValue.SetBool(b)
				}
			}
		}
	}

	return ValidateBody(w, r, b)
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
