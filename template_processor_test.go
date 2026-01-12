package main

import (
	"strings"
	"testing"
)

func TestProcessTemplate_SimpleReplacement(t *testing.T) {
	template := "^FD{{sku_id}}^FS"
	replacements := map[string]string{
		"sku_id": "123456",
	}
	
	got := ProcessTemplate(template, replacements)
	
	if !strings.Contains(got, "^FD123456^FS") {
		t.Errorf("ProcessTemplate() = %v, want it to contain ^FD123456^FS", got)
	}
}
