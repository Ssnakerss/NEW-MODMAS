package field

import "fmt"

var typeConversionMatrix = map[string][]string{
	"text":     {"email", "url", "phone"},
	"integer":  {"decimal", "text"},
	"decimal":  {"text"},
	"date":     {"datetime", "text"},
	"datetime": {"text"},
	"email":    {"text"},
	"url":      {"text"},
	"phone":    {"text"},
}

func IsConversionAllowed(from, to string) bool {
	if from == to {
		return true
	}
	allowed, ok := typeConversionMatrix[from]
	if !ok {
		return false
	}
	for _, t := range allowed {
		if t == to {
			return true
		}
	}
	return false
}

func ValidFieldType(t string) bool {
	valid := map[string]bool{
		"text": true, "integer": true, "decimal": true, "boolean": true,
		"date": true, "datetime": true, "select": true, "multi_select": true,
		"email": true, "url": true, "phone": true, "file": true, "relation": true,
	}
	return valid[t]
}

func ValidateOptions(fieldType string, options map[string]interface{}) error {
	if fieldType == "select" || fieldType == "multi_select" {
		choices, ok := options["choices"]
		if !ok || choices == nil {
			return fmt.Errorf("select field requires 'choices' option")
		}
	}
	return nil
}
