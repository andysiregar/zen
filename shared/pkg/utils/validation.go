package utils

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/go-playground/validator/v10"
)

var validate *validator.Validate

func init() {
	validate = validator.New()
	
	// Register custom validators
	validate.RegisterValidation("subdomain", validateSubdomain)
	validate.RegisterValidation("role", validateRole)
	validate.RegisterValidation("status", validateStatus)
	validate.RegisterValidation("priority", validatePriority)
}

func ValidateStruct(s interface{}) error {
	return validate.Struct(s)
}

func GetValidationErrors(err error) map[string]string {
	errors := make(map[string]string)
	
	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		for _, e := range validationErrors {
			errors[strings.ToLower(e.Field())] = getErrorMessage(e)
		}
	}
	
	return errors
}

func getErrorMessage(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return fmt.Sprintf("%s is required", fe.Field())
	case "email":
		return "Invalid email format"
	case "min":
		return fmt.Sprintf("%s must be at least %s characters", fe.Field(), fe.Param())
	case "max":
		return fmt.Sprintf("%s must be at most %s characters", fe.Field(), fe.Param())
	case "subdomain":
		return "Invalid subdomain format"
	case "role":
		return "Invalid role"
	case "status":
		return "Invalid status"
	case "priority":
		return "Invalid priority"
	default:
		return fmt.Sprintf("%s is invalid", fe.Field())
	}
}

func validateSubdomain(fl validator.FieldLevel) bool {
	subdomain := fl.Field().String()
	
	// Subdomain rules:
	// - Only lowercase letters, numbers, and hyphens
	// - Must start with a letter
	// - Cannot end with a hyphen
	// - 3-63 characters
	pattern := `^[a-z][a-z0-9-]{1,61}[a-z0-9]$`
	matched, _ := regexp.MatchString(pattern, subdomain)
	
	return matched && !strings.Contains(subdomain, "--")
}

func validateRole(fl validator.FieldLevel) bool {
	role := fl.Field().String()
	validRoles := []string{"admin", "agent", "customer", "manager"}
	
	for _, validRole := range validRoles {
		if role == validRole {
			return true
		}
	}
	return false
}

func validateStatus(fl validator.FieldLevel) bool {
	status := fl.Field().String()
	validStatuses := []string{"active", "inactive", "pending", "suspended"}
	
	for _, validStatus := range validStatuses {
		if status == validStatus {
			return true
		}
	}
	return false
}

func validatePriority(fl validator.FieldLevel) bool {
	priority := fl.Field().String()
	validPriorities := []string{"low", "medium", "high", "urgent"}
	
	for _, validPriority := range validPriorities {
		if priority == validPriority {
			return true
		}
	}
	return false
}