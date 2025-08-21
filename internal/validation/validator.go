package validation

import (
	"fmt"
	"regexp"
	"strings"
)

var (
	phoneRegex = regexp.MustCompile(`^\+[1-9]\d{1,14}$`)
)

type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

type ValidationErrors []ValidationError

func (v ValidationErrors) Error() string {
	var messages []string
	for _, err := range v {
		messages = append(messages, fmt.Sprintf("%s: %s", err.Field, err.Message))
	}
	return strings.Join(messages, "; ")
}

func ValidatePhoneNumber(phone string) error {
	if phone == "" {
		return fmt.Errorf("phone number is required")
	}

	if !phoneRegex.MatchString(phone) {
		return fmt.Errorf("phone number must be in E.164 format (e.g., +905551234567)")
	}

	return nil
}

func ValidateMessageContent(content string) error {
	if content == "" {
		return fmt.Errorf("message content is required")
	}

	if len(content) > 160 {
		return fmt.Errorf("message content must be 160 characters or less")
	}

	return nil
}

func ValidateWebhookRequest(to, content string) ValidationErrors {
	var errors ValidationErrors

	if err := ValidatePhoneNumber(to); err != nil {
		errors = append(errors, ValidationError{
			Field:   "to",
			Message: err.Error(),
		})
	}

	if err := ValidateMessageContent(content); err != nil {
		errors = append(errors, ValidationError{
			Field:   "content",
			Message: err.Error(),
		})
	}

	return errors
}
