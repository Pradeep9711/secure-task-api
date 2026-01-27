package utils

import (
	"reflect"
	"regexp"
	"strings"
)

// Validator provides basic validation functionality
type Validator struct {
	Errors map[string]string
}

// NewValidator creates a new validator
func NewValidator() *Validator {
	return &Validator{
		Errors: make(map[string]string),
	}
}

// Required checks if a field is required
func (v *Validator) Required(field, value string) {
	if strings.TrimSpace(value) == "" {
		v.Errors[field] = field + " is required"
	}
}

// MinLength checks if a string has minimum length
func (v *Validator) MinLength(field, value string, min int) {
	if len(value) < min {
		v.Errors[field] = field + " must be at least " + string(rune(min)) + " characters"
	}
}

// MaxLength checks if a string has maximum length
func (v *Validator) MaxLength(field, value string, max int) {
	if len(value) > max {
		v.Errors[field] = field + " must be at most " + string(rune(max)) + " characters"
	}
}

// Email checks if a string is a valid email
func (v *Validator) Email(field, value string) {
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(value) {
		v.Errors[field] = field + " must be a valid email address"
	}
}

// IsValid checks if there are no validation errors
func (v *Validator) IsValid() bool {
	return len(v.Errors) == 0
}

// ValidateStruct performs basic validation on struct fields
func ValidateStruct(s interface{}) map[string]string {
	v := NewValidator()
	val := reflect.ValueOf(s)
	typeOfS := val.Type()

	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		fieldName := typeOfS.Field(i).Name
		tag := typeOfS.Field(i).Tag

		// Check for required tag
		if tag.Get("validate") == "required" {
			v.Required(fieldName, field.String())
		}
	}

	return v.Errors
}
