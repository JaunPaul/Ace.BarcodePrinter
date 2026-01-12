package main

import (
	"strings"
)

// ProcessTemplate replaces placeholders in the template with actual values
// simplified to just sanitization and replacement as layout is now handled by the printer template
func ProcessTemplate(template string, replacements map[string]string) string {
	// 1. Sanitize Inputs
	finalReplacements := make(map[string]string)
	
	for k, v := range replacements {
		finalReplacements[k] = SanitizeZPL(v)
	}
	
	// 2. Perform Replacement
	result := template
	for key, value := range finalReplacements {
		placeholder := "{{" + key + "}}"
		result = strings.ReplaceAll(result, placeholder, value)
	}
	return result
}

// SanitizeZPL removes characters that might break ZPL structure or layout
func SanitizeZPL(input string) string {
	// Remove ZPL command characters
	s := strings.ReplaceAll(input, "^", "")
	s = strings.ReplaceAll(s, "~", "")
	// Replace newlines with space to prevent injection
	s = strings.ReplaceAll(s, "\n", " ")
	s = strings.ReplaceAll(s, "\r", " ")
	return strings.TrimSpace(s)
}
