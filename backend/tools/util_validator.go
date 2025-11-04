package tools

import (
	"fmt"
	"net/url"
	"reflect"
	"regexp"
	"strings"
)

const (
	USERNAME_LENGTH_MIN                 = 3
	USERNAME_LENGTH_MAX                 = 32
	PASSWORD_LENGTH_MIN                 = 8
	PASSWORD_LENGTH_MAX                 = 64
	DESCRIPTION_LENGTH_MIN              = 1
	DESCRIPTION_LENGTH_MAX              = 320
	EMAIL_LENGTH_MAX                    = 256
	DISPLAYNAME_LENGTH_MIN              = 1
	DISPLAYNAME_LENGTH_MAX              = 32
	COLOR_VALUE_MAX                     = 16777215
	COLOR_VALUE_MIN                     = 0
	REDIRECT_URI_SLICE_LEN_MAX          = 10
	REDIRECT_URI_STRING_LEN_MAX         = 128
	VALIDATOR_REQUIRED                  = "REQUIRED"
	VALIDATOR_URI_INVALID               = "VALIDATOR_URI_INVALID"
	VALIDATOR_SLICE_TOO_FEW_ITEMS       = "SLICE_TOO_FEW_ITEMS"
	VALIDATOR_SLICE_TOO_MANY_ITEMS      = "SLICE_TOO_MANY_ITEMS"
	VALIDATOR_INTEGER_TOO_SMALL         = "INTEGER_TOO_SMALL"
	VALIDATOR_INTEGER_TOO_LARGE         = "INTEGER_TOO_LARGE"
	VALIDATOR_STRING_INVALID            = "STRING_INVALID"
	VALIDATOR_STRING_NOT_MATCH          = "STRING_NOT_MATCH"
	VALIDATOR_STRING_TOO_SHORT          = "STRING_TOO_SHORT"
	VALIDATOR_STRING_TOO_LONG           = "STRING_TOO_LONG"
	VALIDATOR_STRING_REQUIRES_SPECIAL   = "STRING_REQUIRES_SPECIAL"
	VALIDATOR_STRING_REQUIRES_LOWERCASE = "STRING_REQUIRES_LOWERCASE"
	VALIDATOR_STRING_REQUIRES_UPPERCASE = "STRING_REQUIRES_UPPERCASE"
	VALIDATOR_STRING_REQUIRES_NUMBER    = "STRING_REQUIRES_NUMBER"
	VALIDATOR_TOKEN_INVALID             = "TOKEN_INVALID"
)

var (
	REGEX_USERNAME    = regexp.MustCompile("^[a-zA-Z0-9_]{3,32}$")
	REGEX_PASSCODE    = regexp.MustCompile("^([0-9]{6}|[0-9ABCDEF]{8})$")
	REGEX_EMAIL       = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	REGEX_HAS_SPECIAL = regexp.MustCompile(`\P{L}`)  // non-letter Unicode
	REGEX_HAS_UPPER   = regexp.MustCompile(`\p{Lu}`) // uppercase letter (any script)
	REGEX_HAS_LOWER   = regexp.MustCompile(`\p{Ll}`) // lowercase letter (any script)
	REGEX_HAS_NUMBER  = regexp.MustCompile(`[0-9]`)  // numbers
)

type ValidateFunc func(value any, param string) *ValidationError

type ValidationError struct {
	Field    string `json:"field"`
	Error    string `json:"id"`
	Literals []any  `json:"literals,omitempty"`
}

var rules = map[string]ValidateFunc{
	"email": func(value any, _ string) *ValidationError {
		s := indirectString(value)
		if len(s) > EMAIL_LENGTH_MAX {
			return &ValidationError{Error: VALIDATOR_STRING_TOO_LONG, Literals: []any{EMAIL_LENGTH_MAX}}
		}
		if !REGEX_EMAIL.MatchString(s) {
			return &ValidationError{Error: VALIDATOR_STRING_INVALID}
		}
		return nil
	},
	"username": func(value any, _ string) *ValidationError {
		s := indirectString(value)
		if len(s) < USERNAME_LENGTH_MIN {
			return &ValidationError{Error: VALIDATOR_STRING_TOO_SHORT, Literals: []any{USERNAME_LENGTH_MIN}}
		}
		if len(s) > USERNAME_LENGTH_MAX {
			return &ValidationError{Error: VALIDATOR_STRING_TOO_LONG, Literals: []any{USERNAME_LENGTH_MAX}}
		}
		if !REGEX_USERNAME.MatchString(s) {
			return &ValidationError{Error: VALIDATOR_STRING_INVALID}
		}
		return nil
	},
	"displayname": func(value any, _ string) *ValidationError {
		s := indirectString(value)
		if len(s) < DISPLAYNAME_LENGTH_MIN {
			return &ValidationError{Error: VALIDATOR_STRING_TOO_SHORT, Literals: []any{DISPLAYNAME_LENGTH_MIN}}
		}
		if len(s) > DISPLAYNAME_LENGTH_MAX {
			return &ValidationError{Error: VALIDATOR_STRING_TOO_LONG, Literals: []any{DISPLAYNAME_LENGTH_MAX}}
		}
		return nil
	},
	"password": func(value any, _ string) *ValidationError {
		s := indirectString(value)
		if len(s) < PASSWORD_LENGTH_MIN {
			return &ValidationError{Error: VALIDATOR_STRING_TOO_SHORT, Literals: []any{PASSWORD_LENGTH_MIN}}
		}
		if len(s) > PASSWORD_LENGTH_MAX {
			return &ValidationError{Error: VALIDATOR_STRING_TOO_LONG, Literals: []any{PASSWORD_LENGTH_MAX}}
		}
		if !REGEX_HAS_SPECIAL.MatchString(s) {
			return &ValidationError{Error: VALIDATOR_STRING_REQUIRES_SPECIAL}
		}
		if !REGEX_HAS_UPPER.MatchString(s) {
			return &ValidationError{Error: VALIDATOR_STRING_REQUIRES_UPPERCASE}
		}
		if !REGEX_HAS_LOWER.MatchString(s) {
			return &ValidationError{Error: VALIDATOR_STRING_REQUIRES_LOWERCASE}
		}
		if !REGEX_HAS_NUMBER.MatchString(s) {
			return &ValidationError{Error: VALIDATOR_STRING_REQUIRES_NUMBER}
		}
		return nil
	},
	"description": func(value any, _ string) *ValidationError {
		s := indirectString(value)
		if len(s) < DESCRIPTION_LENGTH_MIN {
			return &ValidationError{Error: VALIDATOR_STRING_TOO_SHORT, Literals: []any{DESCRIPTION_LENGTH_MIN}}
		}
		if len(s) > DESCRIPTION_LENGTH_MAX {
			return &ValidationError{Error: VALIDATOR_STRING_TOO_LONG, Literals: []any{DESCRIPTION_LENGTH_MAX}}
		}
		return nil
	},
	"passcode": func(value any, _ string) *ValidationError {
		s := indirectString(value)
		if !REGEX_PASSCODE.MatchString(s) {
			return &ValidationError{Error: VALIDATOR_STRING_INVALID}
		}
		return nil
	},
	"color": func(value any, _ string) *ValidationError {
		i := indirectInt(value)
		if i < COLOR_VALUE_MIN {
			return &ValidationError{Error: VALIDATOR_INTEGER_TOO_SMALL, Literals: []any{COLOR_VALUE_MIN}}
		}
		if i > COLOR_VALUE_MAX {
			return &ValidationError{Error: VALIDATOR_INTEGER_TOO_LARGE, Literals: []any{COLOR_VALUE_MAX}}
		}
		return nil
	},
	"token": func(value any, _ string) *ValidationError {
		s := indirectString(value)
		if !CompareSignedString(s) {
			return &ValidationError{Error: VALIDATOR_TOKEN_INVALID}
		}
		return nil
	},
	"uri": func(value any, _ string) *ValidationError {
		s := indirectString(value)
		if _, err := url.Parse(s); err != nil {
			return &ValidationError{Error: VALIDATOR_URI_INVALID}
		}
		return nil
	},
	"required": func(value any, _ string) *ValidationError {
		if value == nil {
			return &ValidationError{Error: VALIDATOR_REQUIRED}
		}

		// Retrieve Type
		v := reflect.ValueOf(value)
		for v.Kind() == reflect.Ptr {
			if v.IsNil() {
				return &ValidationError{Error: VALIDATOR_REQUIRED}
			}
			v = v.Elem() // dereference pointer
		}

		switch v.Kind() {
		case reflect.String, reflect.Slice, reflect.Map:
			if v.Len() == 0 {
				return &ValidationError{Error: VALIDATOR_REQUIRED}
			}
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			if v.Int() == 0 {
				return &ValidationError{Error: VALIDATOR_REQUIRED}
			}
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			if v.Uint() == 0 {
				return &ValidationError{Error: VALIDATOR_REQUIRED}
			}
		case reflect.Bool:
			if !v.Bool() {
				return &ValidationError{Error: VALIDATOR_REQUIRED}
			}
		default:
			panic("validator: unsupported type " + v.Kind().String())
		}

		return nil
	},
}

func indirectString(value any) string {
	s := ""
	v := reflect.ValueOf(value)
	if v.Kind() == reflect.Ptr && !v.IsNil() {
		s = v.Elem().String()
	} else {
		s = v.String()
	}
	return s
}

func indirectInt(value any) int64 {
	n := int64(0)
	v := reflect.ValueOf(value)
	if v.Kind() == reflect.Ptr && !v.IsNil() {
		n = v.Elem().Int()
	} else {
		n = v.Int()
	}
	return n
}

func ValidateStruct(data any) ([]ValidationError, error) {
	return validateRecursive(data, "")
}

func validateRecursive(data any, prefix string) ([]ValidationError, error) {
	var errors []ValidationError

	v := reflect.ValueOf(data)
	if !v.IsValid() {
		return nil, nil
	}

	// Dereference pointer
	for v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return nil, nil
		}
		v = v.Elem()
	}

	switch v.Kind() {

	// Validate Struct
	case reflect.Struct:
		t := v.Type()
		for i := 0; i < v.NumField(); i++ {

			field := t.Field(i)
			fieldName := field.Name
			if prefix != "" {
				fieldName = prefix + "." + fieldName
			}
			fieldTag := field.Tag.Get("validate")
			if fieldTag == "" {
				continue
			}
			fieldValue := v.Field(i).Interface()
			fieldRules := strings.Split(fieldTag, ",")

			for j := 0; j < len(fieldRules); j++ {
				ruleParts := strings.SplitN(fieldRules[j], "=", 2)
				ruleName := strings.ToLower(ruleParts[0])
				ruleParam := ""
				if len(ruleParts) == 2 {
					ruleParam = ruleParts[1]
				}

				// Skip Empty Fields as Requested
				if ruleName == "omitempty" {
					if rules["required"](fieldValue, "") != nil {
						break
					}
					continue
				}

				// Validate Struct Value
				fn, ok := rules[ruleName]
				if !ok {
					panic("validator: unknown rule " + ruleName)
				}
				if err := fn(fieldValue, ruleParam); err != nil {
					err.Field = fieldName
					errors = append(errors, *err)
					break
				}

			}

		}

	// Validate Items in Slice/Array
	case reflect.Slice, reflect.Array:
		for i := 0; i < v.Len(); i++ {
			reasons, _ := validateRecursive(
				v.Index(i).Interface(),
				fmt.Sprintf("%s[%d]", prefix, i),
			)
			errors = append(errors, reasons...)
		}

	}

	return errors, nil
}
