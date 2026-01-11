package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

// SendToPrinter sends the raw ZPL data to a Windows printer
// It uses "cmd /c copy /b" which is robust for raw printing on Windows
func SendToPrinter(printerName string, data []byte) error {
	// Create a temporary file
	tempDir := os.TempDir()
	tempFile := filepath.Join(tempDir, "label_print_job.zpl")
	
	err := os.WriteFile(tempFile, data, 0644)
	if err != nil {
		return fmt.Errorf("failed to create temp file: %v", err)
	}
	// defer os.Remove(tempFile) // Optional: keep for debugging if needed, or remove

	if runtime.GOOS == "windows" {
		// Construct the printer path. 
		// If it's a shared printer or local, accessing via \\localhost\PrinterName often works best for raw copying
		// Alternatively, if it's just a local printer name, we can try to copy directly to the port if it was LPT1,
		// but for USB printers, sharing it and using the share name is the most reliable "simple" method.
		
		destination := fmt.Sprintf("\\\\localhost\\%s", printerName)
		
		// Command: cmd /c copy /b <file> <destination>
		cmd := exec.Command("cmd", "/c", "copy", "/b", tempFile, destination)
		output, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("print command failed: %v, output: %s", err, string(output))
		}
	} else {
		// Linux/Mac fallback for testing (mock)
		fmt.Printf("MOCK PRINTING to %s:\n%s\n", printerName, string(data))
	}

	return nil
}
