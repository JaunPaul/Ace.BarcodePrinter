package main

import (
	"fmt"
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
	
	// Calculate 'by' if not manually provided
	byVal := normalBY
	if utf8.RuneCountInString(sku) > skuThreshold {
		byVal = longBY
	}
	
	// Prepare final map
	finalReplacements := make(map[string]string)
	
	// Add 'by'
	finalReplacements["by"] = byVal
	
	// Calculate Barcode Centering
	// Label Width = 384 dots (48mm * 8dmm) (Assuming 203dpi or similar standard 8dpmm)
	// Actually 384 is typical for 48mm wide paper (48 * 8 = 384)
	const labelWidth = 384
	
	// Convert byVal to integer width for calculation (1 or 2)
	moduleWidth := 2
	if byVal == "1" {
		moduleWidth = 1
	}
	
	// Estimate width
	bcWidth := EstimateCode128Width(sku, moduleWidth)
	
	// Centers: X = (LabelWidth - BarcodeWidth) / 2
	// ZPL ^FT coordinates for barcodes are roughly Bottom-Left origin of the code usually,
	// but ^FT uses the Field Origin relative to rotation.
	// However, for ^BC (Code 128), the position specified in ^FO or ^FT is the starting point (top-left or bottom-left depending on command).
	// In the template, it was ^FT114,112. ^FT is Field Typeset (baseline based).
	// If the user's previous hardcoded value was 114 for a normal sku, let's see.
	// Let's assume standard centering formula.
	
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
// This is an estimation. Code 128 uses variable width encoding.
// Subsets B (alphanumeric) and C (numeric pairs) are most common.
// ZPL's ^BC usually auto-switches.
// Each character is 11 modules wide (except Stop char which is 13).
// Quiet zones are usually handled by the printer buffer or should be added, 
// but ZPL positioning calculates from the start of the bars.
func EstimateCode128Width(content string, moduleWidth int) int {
	// Start code (11) + Check digit (11) + Stop code (13) = 35 modules overhead
	// Then data characters.
	// Heuristic:
	// If pure numeric and length is even -> Subset C (1 char per 2 digits) = 5.5 modules per digit
	// If pure numeric and length is odd -> Subset C for pairs + 1 Subset B? Or Subset C with shifts.
	// ZPL is smart. Let's assume standard alphanumeric (Subset B) for worst-case width (safety),
	// or try to detect numeric optimization which shrinks it significantly.
	
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

