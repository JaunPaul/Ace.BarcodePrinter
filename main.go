package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

)

type PrintItem struct {
	Name  string
	Price string
	SKU   string
	Qty   string // Keep as string for Entry binding, parse on print
}

func main() {
	a := app.New()
	w := a.NewWindow("Barcode Printer")
	w.Resize(fyne.NewSize(800, 600))

	// Data
	var csvHeaders []string
	var csvData [][]string
	var items []PrintItem
	var templateContent string = loadDefaultTemplate()

	// ---------------------------------------------------------
	// UI Components
	// ---------------------------------------------------------

	// Status
	statusLabel := widget.NewLabel("Ready")

	// 1. File Selection
	fileLabel := widget.NewLabel("No CSV file loaded")
	
	// 2. Mapping Dropdowns
	itemMapSelect := widget.NewSelect([]string{}, func(s string) {})
	priceMapSelect := widget.NewSelect([]string{}, func(s string) {})
	skuMapSelect := widget.NewSelect([]string{}, func(s string) {})
	qtyMapSelect := widget.NewSelect([]string{}, func(s string) {}) // New Stock/Qty mapping

	// 3. Printer Settings
	printerSelect := widget.NewSelect([]string{}, func(s string) {})
	printerSelect.PlaceHolder = "Select Printer..."

	refreshPrinters := func() {
		printers, err := GetPrinters()
		if err != nil {
			// If we fail, just show one dummy or leave empty logic?
			// Best to just log error to status
			statusLabel.SetText("Failed to load printers: " + err.Error())
			return
		}
		printerSelect.Options = printers
		printerSelect.Refresh()
		
		// Try to select the previous one or default
		if len(printers) > 0 {
			printerSelect.SetSelected(printers[0])
			// If "ZDesigner" exists, prefer it
			for _, p := range printers {
				if strings.Contains(p, "ZDesigner") || strings.Contains(p, "Zebra") {
					printerSelect.SetSelected(p)
					break
				}
			}
		}
	}
	
	// Refresh button
	refreshPrintersBtn := widget.NewButtonWithIcon("", 
		theme.ViewRefreshIcon(), 
		func() { refreshPrinters() },
	)
	refreshPrintersBtn.SetText("Refresh") // Fallback if icon fails or just text

	// Initial load
	refreshPrinters()

	// 4. Data Table
	// We need a Table widget. 
	// Columns: 0=Name, 1=Price, 2=SKU, 3=Qty (Editable)
	t := widget.NewTable(
		func() (int, int) {
			return len(items), 4
		},
		func() fyne.CanvasObject {
			// Template cell. We use a container to switch between Label and Entry
			return container.NewStack(widget.NewLabel("Start"), widget.NewEntry())
		},
		func(id widget.TableCellID, o fyne.CanvasObject) {
			// Update cell
			// The container has 2 objects: [0]=Label, [1]=Entry
			// For cols 0-2 use Label, hide Entry. For col 3 use Entry, hide Label.
			stack := o.(*fyne.Container)
			label := stack.Objects[0].(*widget.Label)
			entry := stack.Objects[1].(*widget.Entry)
			
			if id.Row >= len(items) { return }
			item := items[id.Row]

			if id.Col == 3 {
				// Editable Quantity
				label.Hide()
				entry.Show()
				entry.SetText(item.Qty)
				// Important: Save changes back to data
				entry.OnChanged = func(s string) {
					items[id.Row].Qty = s
				}
			} else {
				// Read-only info
				entry.Hide()
				label.Show()
				switch id.Col {
				case 0:
					label.SetText(item.Name)
				case 1:
					label.SetText(item.Price)
				case 2:
					label.SetText(item.SKU)
				}
			}
		},
	)
	
	// Define column widths
	t.SetColumnWidth(0, 200) // Name
	t.SetColumnWidth(1, 80)  // Price
	t.SetColumnWidth(2, 100) // SKU
	t.SetColumnWidth(3, 80)  // Qty

	// ---------------------------------------------------------
	// Logic
	// ---------------------------------------------------------

	refreshTable := func() {
		// Re-scan CSV data with current mappings to populate 'items'
		if len(csvData) == 0 { return }
		
		itemNameCol := itemMapSelect.Selected
		priceCol := priceMapSelect.Selected
		skuCol := skuMapSelect.Selected
		qtyCol := qtyMapSelect.Selected // Optional
		
		if itemNameCol == "" || priceCol == "" || skuCol == "" {
			return // Not fully mapped yet
		}

		// Find indices
		itemIdx, priceIdx, skuIdx, qtyIdx := -1, -1, -1, -1
		for i, h := range csvHeaders {
			if h == itemNameCol { itemIdx = i }
			if h == priceCol { priceIdx = i }
			if h == skuCol { skuIdx = i }
			if h == qtyCol { qtyIdx = i }
		}

		items = nil // Clear current
		for _, row := range csvData {
			q := "1" // Default
			if qtyIdx != -1 && qtyIdx < len(row) {
				// If mapped, use value. 
				// Basic sanitization/defaulting could happen here, keeping as string is safest for UI
				val := row[qtyIdx]
				if val == "" { val = "1" }
				q = val
			}
			
			// Safety checks for bounds
			safeGet := func(idx int) string {
				if idx >= 0 && idx < len(row) { return row[idx] }
				return ""
			}

			items = append(items, PrintItem{
				Name:  safeGet(itemIdx),
				Price: safeGet(priceIdx),
				SKU:   safeGet(skuIdx),
				Qty:   q,
			})
		}
		t.Refresh()
		statusLabel.SetText(fmt.Sprintf("Mapped %d items", len(items)))
	}

	// Trigger refresh when mappings change
	itemMapSelect.OnChanged = func(s string) { refreshTable() }
	priceMapSelect.OnChanged = func(s string) { refreshTable() }
	skuMapSelect.OnChanged = func(s string) { refreshTable() }
	qtyMapSelect.OnChanged = func(s string) { refreshTable() }

	// File Button Logic
	fileBtn := widget.NewButton("Select CSV File", func() {
		fd := dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
			if err != nil {
				dialog.ShowError(err, w)
				return
			}
			if reader == nil {
				return
			}
			defer reader.Close()

			path := reader.URI().Path()
			fileLabel.SetText("Loaded: " + path)

			h, d, err := ReadCSV(path)
			if err != nil {
				dialog.ShowError(err, w)
				return
			}
			csvHeaders = h
			csvData = d
			
			// Update Selects (include Empty option for Qty)
			opts := csvHeaders
			itemMapSelect.Options = opts
			priceMapSelect.Options = opts
			skuMapSelect.Options = opts
			qtyMapSelect.Options = append([]string{"(Default to 1)"}, opts...)
			
			itemMapSelect.Refresh()
			priceMapSelect.Refresh()
			skuMapSelect.Refresh()
			qtyMapSelect.Refresh()
			
			// Auto-select smart defaults
			qtyMapSelect.SetSelected("(Default to 1)") // Reset
			for _, h := range csvHeaders {
				if h == "Item Name" || h == "item_name" { itemMapSelect.SetSelected(h) }
				if h == "Price" || h == "price" { priceMapSelect.SetSelected(h) }
				if h == "SKU" || h == "sku_id" { skuMapSelect.SetSelected(h) }
				if h == "Stock" || h == "Quantity" || h == "qty" { qtyMapSelect.SetSelected(h) }
			}
			
			refreshTable()
		}, w)
		
		fd.SetFilter(storage.NewExtensionFileFilter([]string{".csv"}))
		fd.Show()
	})

	// Print Button Logic
	printBtn := widget.NewButton("Print Labels", func() {
		if len(items) == 0 {
			dialog.ShowError(fmt.Errorf("no items to print"), w)
			return
		}
		pName := printerSelect.Selected
		if pName == "" {
			dialog.ShowError(fmt.Errorf("printer name required"), w)
			return
		}

		successCount := 0
		totalLabels := 0
		
		for _, item := range items {
			q, err := strconv.Atoi(item.Qty)
			if err != nil || q <= 0 {
				continue // Skip invalid quantity
			}

			replacements := map[string]string{
				"item_name": item.Name,
				"price":     item.Price,
				"sku_id":    item.SKU,
			}
			
			zpl := ProcessTemplate(templateContent, replacements)
			
			// Send to printer q times
			// Optimization: Could send Qty command ^PQq if generic, but template logic is simpler looped for now 
			// checking template for ^PQ...
			// The current template has ^PQ1,,,Y. We could modify the ZPL to Use ^PQ<qty> but string replace is safer.
			// Actually looping print jobs `q` times is slow for Windows spooler. 
			// Better to inject `^PQ` + q.
			// But user asked for "send updated instructions". 
			// Let's stick to loop for absolute simplicity, OR better: Modify the ^PQ command.
			// Current template ends with ^PQ1,,,Y
			// Let's replace ^PQ1 with ^PQ<q>
			
			// But wait, the template has hardcoded ^PQ1.
			// Let's just loop for now to be safe and robust, unless Qty is huge.
			for i := 0; i < q; i++ {
				err := SendToPrinter(pName, []byte(zpl))
				if err != nil {
					statusLabel.SetText(fmt.Sprintf("Error printing %s: %v", item.Name, err))
					return
				}
			}
			successCount++
			totalLabels += q
		}
		
		statusLabel.SetText(fmt.Sprintf("Sent jobs for %d items (%d total labels)", successCount, totalLabels))
		dialog.ShowInformation("Complete", fmt.Sprintf("Printed %d labels.", totalLabels), w)
	})

	// Layout
	// Top Control Area
	topControls := container.NewVBox(
		widget.NewLabelWithStyle("Ace Barcode Printer", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		fileBtn,
		fileLabel,
		widget.NewSeparator(),
		widget.NewLabel("Column Mapping:"),
		container.NewGridWithColumns(2, 
			widget.NewLabel("Item Name:"), itemMapSelect,
			widget.NewLabel("Price:"), priceMapSelect,
			widget.NewLabel("SKU ID:"), skuMapSelect,
			widget.NewLabel("Stock/Qty:"), qtyMapSelect,
		),
		widget.NewSeparator(),
		widget.NewLabel("Print Settings:"),
		container.NewGridWithColumns(2, widget.NewLabel("Printer:"), 
			container.NewBorder(nil, nil, nil, refreshPrintersBtn, printerSelect), // Layout: [Select...][Refresh]
		),
		printBtn,
		statusLabel,
		widget.NewSeparator(),
		widget.NewLabel("Review Items (Edit Quantity in Table):"),
	)

	// Combine Scrolled Table
	// We put the top controls in a fixed container and table takes rest
	content := container.NewBorder(topControls, nil, nil, nil, t)

	w.SetContent(content)
	w.ShowAndRun()
}

func loadDefaultTemplate() string {
    content, err := os.ReadFile("ACE_50X30.zpl")
    if err == nil {
        // Strip BOM if present
        s := string(content)
        if strings.HasPrefix(s, "\uFEFF") {
            s = strings.TrimPrefix(s, "\uFEFF")
        }
        return s
    }
    return `^XA^FDError Loading Template^FS^XZ`
}
