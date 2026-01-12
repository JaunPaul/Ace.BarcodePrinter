package main

import (
	"strings"
	"unicode/utf8"
)

// ProcessTemplate replaces placeholders in the template with actual values
// It also handles dynamic barcode width (by) calculation and sanitization
func ProcessTemplate(template string, replacements map[string]string) string {
	// 1. Sanitize Inputs & Calculate Logic
	
	// Default Logic Configuration
	const skuThreshold = 12
	const normalBY = "2"
	const longBY = "1"
	
	sku := replacements["sku_id"]
	
	// Calculate 'by' if not manually provided (it won't be from main.go typically)
	byVal := normalBY
	if utf8.RuneCountInString(sku) > skuThreshold {
		byVal = longBY
	}
	
	// Prepare final map
	finalReplacements := make(map[string]string)
	
	// Add 'by'
	finalReplacements["by"] = byVal
	
	// Copy and sanitize others
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

