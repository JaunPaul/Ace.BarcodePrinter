package main

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// GetPrinters returns a list of installed printer names
func GetPrinters() ([]string, error) {
	var printers []string

	if runtime.GOOS == "windows" {
		// Use PowerShell to get printers
		// Command: Get-Printer | Select-Object -ExpandProperty Name
		cmd := exec.Command("powershell", "-NoProfile", "-NonInteractive", "-Command", "Get-Printer | Select-Object -ExpandProperty Name")
		output, err := cmd.CombinedOutput()
		if err != nil {
			return nil, fmt.Errorf("failed to list printers: %v", err)
		}

		scanner := bufio.NewScanner(bytes.NewReader(output))
		for scanner.Scan() {
			name := strings.TrimSpace(scanner.Text())
			if name != "" {
				printers = append(printers, name)
			}
		}
	} else {
		// Linux/Mac implementation (lpstat)
		// Command: lpstat -p | awk '{print $2}'
		// lpstat output format: "printer <name> is idle..."
		cmd := exec.Command("lpstat", "-a")
		output, err := cmd.CombinedOutput()
		if err == nil {
			scanner := bufio.NewScanner(bytes.NewReader(output))
			for scanner.Scan() {
				line := scanner.Text()
				// Expected format: "PrinterName accepting requests since..."
				parts := strings.Fields(line)
				if len(parts) > 0 {
					printers = append(printers, parts[0])
				}
			}
		} else {
			// Fallback for dev environment without printing system
			printers = []string{"Mock_Printer_1", "Mock_Zebra_GK420d"}
		}
	}

	if len(printers) == 0 {
		return nil, fmt.Errorf("no printers found")
	}

	return printers, nil
}

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
		
		// Use explicit quoting for safety against spaces in paths/names
		// cmd /c "copy /b "source" "dest""
		commandStr := fmt.Sprintf("copy /b \"%s\" \"%s\"", tempFile, destination)
		cmd := exec.Command("cmd", "/c", commandStr)
		output, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("print command failed: %v, output: %s, command: %s", err, string(output), commandStr)
		}
	} else {
		// Linux/Mac fallback for testing (mock)
		fmt.Printf("MOCK PRINTING to %s:\n%s\n", printerName, string(data))
	}

	return nil
}

