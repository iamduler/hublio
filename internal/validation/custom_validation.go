package validation

import (
	"path/filepath"
	"regexp"
	"shopping-cart/internal/utils"
	"strconv"
	"strings"

	"github.com/go-playground/validator/v10"
)

func RegisterCustomValidations(v *validator.Validate) error {
	// Email
	var blockedDomains = map[string]bool{
		"blacklist.com": true,
	}

	v.RegisterValidation("email_advanced", func(fl validator.FieldLevel) bool {
		email := fl.Field().String()
		domain := strings.Split(email, "@")[1]
		domain = utils.NormalizeString(domain)
		return !blockedDomains[domain]
	})

	// Password
	v.RegisterValidation("password_advanced", func(fl validator.FieldLevel) bool {
		password := fl.Field().String()

		if len(password) < 8 || len(password) > 50 {
			return false
		}

		hasUpper := regexp.MustCompile(`[A-Z]`).MatchString(password)
		hasLower := regexp.MustCompile(`[a-z]`).MatchString(password)
		hasNumber := regexp.MustCompile(`[0-9]`).MatchString(password)
		hasSpecial := regexp.MustCompile(`[!@#$%^&*()_+\-=\[\]{};':"\\|,.<>/?]`).MatchString(password)

		return hasUpper && hasLower && hasNumber && hasSpecial
	})

	// Slug
	var slugRegex = regexp.MustCompile(`^[a-z0-9]+(?:[-._][a-z0-9]+)*$`)
	v.RegisterValidation("slug", func(fl validator.FieldLevel) bool {
		return slugRegex.MatchString(fl.Field().String())
	})

	// Search
	var searchRegex = regexp.MustCompile(`^[a-zA-Z0-9\s]+$`)
	v.RegisterValidation("search", func(fl validator.FieldLevel) bool {
		return searchRegex.MatchString(fl.Field().String())
	})

	// Min integer
	v.RegisterValidation("min_int", func(fl validator.FieldLevel) bool {
		minStr := fl.Param()
		minVal, err := strconv.ParseInt(minStr, 10, 64)

		if err != nil {
			return false
		}

		return fl.Field().Int() >= minVal
	})

	// Max integer
	v.RegisterValidation("max_int", func(fl validator.FieldLevel) bool {
		maxStr := fl.Param()
		maxVal, err := strconv.ParseInt(maxStr, 10, 64)

		if err != nil {
			return false
		}

		return fl.Field().Int() <= maxVal
	})

	// File extension
	v.RegisterValidation("file_extension", func(fl validator.FieldLevel) bool {
		fileName := fl.Field().String() // Format: path/to/image.jpg
		allowedString := fl.Param()

		if allowedString == "" {
			return false
		}

		allowedExtensions := strings.Fields(allowedString) // convert string to array like ["jpg", "jpeg", "png", "gif", "bmp", "tiff", "ico", "webp"]
		fileName = strings.ToLower(filepath.Ext(fileName)) // get the file name like "image.jpg"
		extension := strings.TrimPrefix(fileName, ".")     // get the extension of the file like "jpg"

		for _, allowedExtension := range allowedExtensions {
			if strings.ToLower(allowedExtension) == extension {
				return true
			}
		}

		return false
	})

	return nil
}
