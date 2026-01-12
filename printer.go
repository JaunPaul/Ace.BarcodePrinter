package main

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"os/exec"
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
func SendToPrinter(printerName string, data []byte) error {
	// Ensure data has a trailing newline
	if len(data) > 0 && data[len(data)-1] != '\n' {
		data = append(data, '\n')
	}

	// For Windows, writing directly to the UNC path \\localhost\PrinterName is more robust
	// than using `cmd /c copy` because it avoids shell quoting issues.
	if runtime.GOOS == "windows" {
		destination := fmt.Sprintf("\\\\localhost\\%s", printerName)
		// Write directly to the printer share
		// os.WriteFile handles open/write/close
		err := os.WriteFile(destination, data, 0644)
		if err != nil {
			// If direct write fails, try fallback or just return error
			return fmt.Errorf("failed to send to printer %s: %v", destination, err)
		}
		return nil
	} else {
		// Linux/Mac fallback (`lp` or mock)
		// For consistency with previous logic, we can keep using lp command or mock
		fmt.Printf("MOCK PRINTING to %s: [Data Length %d]\n", printerName, len(data))
		return nil
	}
}


