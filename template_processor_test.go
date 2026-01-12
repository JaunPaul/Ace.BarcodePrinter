package main

import (
	"strings"
	"testing"
)

func TestEstimateCode128Width(t *testing.T) {
	tests := []struct {
		name        string
		content     string
		moduleWidth int
		want        int // Approx
	}{
		{"Numeric Short", "12", 2, (35 + 11) * 2}, // 1 pair = 11 modules
		{"Numeric Long", "123456", 2, (35 + 33) * 2}, // 3 pairs = 33 modules
		{"Alpha Short", "AB", 2, (35 + 22) * 2}, // 2 chars = 22
		{"Mixed", "A1", 2, (35 + 22) * 2}, // 2 chars (no optimization likely)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := EstimateCode128Width(tt.content, tt.moduleWidth)
			if got != tt.want {
				t.Errorf("EstimateCode128Width() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestProcessTemplate_Centering(t *testing.T) {
	template := "^FT{{barcode_x}},112^BCN"
	replacements := map[string]string{
		"sku_id": "123456", // Numeric optimization -> 3 pairs * 11 = 33 mod. Base 35. Total 68 mod. Width 68*2=136.
		// Label 384. Center = (384 - 136)/2 = 248/2 = 124.
	}
	
	// Module width defaults to 2 for short strings
	got := ProcessTemplate(template, replacements)
	
	if !strings.Contains(got, "^FT124,112") {
		t.Errorf("ProcessTemplate() = %v, want it to contain ^FT124,112", got)
	}
}
