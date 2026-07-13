package validation

import (
	"fmt"
	"shopping-cart/internal/utils"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

func InitValidator() error {
	v, ok := binding.Validator.Engine().(*validator.Validate)

	if !ok {
		return fmt.Errorf("validator engine not found")
	}

	RegisterCustomValidations(v)

	return nil
}

func HandleValidationErrors(err error) gin.H {
	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		errors := make(map[string]string)

		for _, e := range validationErrors {
			// e.Namespace() like "root.field"
			rootPath := strings.Split(e.Namespace(), ".")[0]
			rawPath := strings.TrimPrefix(e.Namespace(), rootPath+".")
			parts := strings.Split(rawPath, ".")

			for i, part := range parts {
				if strings.Contains(part, "[") {
					idx := strings.Index(part, "[")
					basePart := part[:idx]
					index := part[idx:]
					parts[i] = utils.CamelCaseToSnakeCase(basePart) + index
				} else {
					parts[i] = utils.CamelCaseToSnakeCase(part)
				}
			}

			fieldPath := strings.Join(parts, ".")

			switch e.Tag() {
			case "required":
				errors[fieldPath] = fmt.Sprintf("%s is required", fieldPath)
			case "min":
				errors[fieldPath] = fmt.Sprintf("%s must be greater than %s characters long", fieldPath, e.Param())
			case "min_int":
				errors[fieldPath] = fmt.Sprintf("%s must be greater than or equal to %s", fieldPath, e.Param())
			case "max":
				errors[fieldPath] = fmt.Sprintf("%s must be less than %s characters long", fieldPath, e.Param())
			case "max_int":
				errors[fieldPath] = fmt.Sprintf("%s must be less than or equal to %s", fieldPath, e.Param())
			case "regex":
				errors[fieldPath] = fmt.Sprintf("%s must match the regex %s", fieldPath, e.Param())
			case "gt":
				errors[fieldPath] = fmt.Sprintf("%s must be greater than %s", fieldPath, e.Param())
			case "gte":
				errors[fieldPath] = fmt.Sprintf("%s must be greater than or equal to %s", fieldPath, e.Param())
			case "lt":
				errors[fieldPath] = fmt.Sprintf("%s must be less than %s", fieldPath, e.Param())
			case "lte":
				errors[fieldPath] = fmt.Sprintf("%s must be less than or equal to %s", fieldPath, e.Param())
			case "uuid":
				errors[fieldPath] = fmt.Sprintf("%s must be a valid UUID", fieldPath)
			case "email":
				errors[fieldPath] = fmt.Sprintf("%s must be a valid email address", fieldPath)
			case "email_advanced":
				errors[fieldPath] = fmt.Sprintf("%s must be a valid email address and not in the blacklist", fieldPath)
			case "url":
				errors[fieldPath] = fmt.Sprintf("%s must be a valid URL", fieldPath)
			case "json":
				errors[fieldPath] = fmt.Sprintf("%s must be a valid JSON", fieldPath)
			case "slug":
				errors[fieldPath] = fmt.Sprintf("%s must be a format of slug", fieldPath)
			case "search":
				errors[fieldPath] = fmt.Sprintf("%s must be a format of search", fieldPath)
			case "oneof":
				errors[fieldPath] = fmt.Sprintf("%s must be one of the following: %s", fieldPath, strings.Join(strings.Split(e.Param(), " "), ", "))
			case "dive":
				errors[fieldPath] = fmt.Sprintf("%s must be a valid array", fieldPath)
			case "file_extension":
				errors[fieldPath] = fmt.Sprintf("%s must be a valid file extension: %s", fieldPath, strings.Join(strings.Split(e.Param(), " "), ", "))
			case "password_advanced":
				errors[fieldPath] = fmt.Sprintf("%s must be at least 8 characters long and at most 50 characters long and must contain at least one uppercase letter, one lowercase letter, one number, and one special character", fieldPath)
			default:
				errors[fieldPath] = fmt.Sprintf("%s is not valid", fieldPath)
			}
		}

		return gin.H{"errors": errors}
	}

	return gin.H{"error": "Validation error: " + err.Error()}
}
