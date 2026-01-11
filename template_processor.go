package main

import (
	"strings"
)

// ProcessTemplate replaces placeholders in the template with actual values
func ProcessTemplate(template string, replacements map[string]string) string {
	result := template
	for key, value := range replacements {
		placeholder := "{{" + key + "}}"
		result = strings.ReplaceAll(result, placeholder, value)
	}
	return result
}
