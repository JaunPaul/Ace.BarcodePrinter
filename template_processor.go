package main

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

// ProcessTemplate replaces placeholders in the template with actual values
// It also handles dynamic barcode width (by) calculation, centering, and sanitization
func ProcessTemplate(template string, replacements map[string]string) string {
	// 1. Sanitize Inputs & Calculate Logic
	
	// Default Logic Configuration
	const skuThreshold = 12
	const normalBY = "2"
	const longBY = "1"
	
	sku := replacements["sku_id"]
	
	// Calculate 'by' (Barcode Module Width)
	byVal := normalBY
	if utf8.RuneCountInString(sku) > skuThreshold {
		byVal = longBY
	}
	
	// Prepare final map
	finalReplacements := make(map[string]string)
	
	// Add 'by'
	finalReplacements["by"] = byVal
	
	// Calculate Barcode Centering
	// Label Width = 384 dots (48mm * 8dpmm)
	const labelWidth = 384
	
	// Convert byVal to integer width for calculation (1 or 2)
	moduleWidth := 2
	if byVal == "1" {
		moduleWidth = 1
	}
	
	// Estimate width
	bcWidth := EstimateCode128Width(sku, moduleWidth)
	
	// Centers: X = (LabelWidth - BarcodeWidth) / 2
	xPos := (labelWidth - bcWidth) / 2
	if xPos < 0 {
		xPos = 0 // Safety
	}
	
	finalReplacements["barcode_x"] = fmt.Sprintf("%d", xPos)
	
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

// EstimateCode128Width calculates approximate width in dots of a Code 128 barcode
func EstimateCode128Width(content string, moduleWidth int) int {
	// Start code (11) + Check digit (11) + Stop code (13) = 35 modules overhead
	// Then data characters.
	// Heuristic for Code 128 auto-switching:
	// If pure numeric and length is even -> Subset C (1 char per 2 digits) = 5.5 modules per digit (11 per pair)
	// Otherwise -> Subset B (11 modules per char)
	
	isNumeric := true
	for _, r := range content {
		if r < '0' || r > '9' {
			isNumeric = false
			break
		}
	}
	
	dataModules := 0
	if isNumeric && len(content) >= 2 {
		// Optimized packing (approximate)
		// Ceil(Length / 2) * 11
		pairs := (len(content) + 1) / 2
		dataModules = pairs * 11
	} else {
		// Subset B: 11 modules per character
		dataModules = len(content) * 11
	}
	
	totalModules := 35 + dataModules
	return totalModules * moduleWidth
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
