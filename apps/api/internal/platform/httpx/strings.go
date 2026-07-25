package httpx

import (
	"regexp"
	"strings"
)

func CamelCaseToSnakeCase(str string) string {
	rs := strings.ToLower(regexp.MustCompile(`([A-Z])`).ReplaceAllString(str, "_$1"))
	rs = strings.TrimPrefix(rs, "_")
	rs = strings.TrimSuffix(rs, "_")
	return rs
}

func NormalizeString(str string) string {
	return strings.ToLower(strings.TrimSpace(str))
}

func ConvertToInt32Pointer(value int32) *int32 {
	if value == 0 {
		return nil
	}

	return &value
}

func CapitalizeFirstLetter(str string) string {
	if len(str) == 0 {
		return str
	}

	return strings.ToUpper(str[:1]) + str[1:]
}
