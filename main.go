package main

import (
	"fmt"
	"os"
	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/widget"
)

func main() {
	a := app.New()
	w := a.NewWindow("Barcode Printer")
	w.Resize(fyne.NewSize(600, 500))

	// UI State Variables
	var csvHeaders []string
	var csvData [][]string
	var templateContent string = loadDefaultTemplate()
	
	// Widgets
	fileLabel := widget.NewLabel("No CSV file loaded")
	
	// Dropdowns for mapping
	itemMapSelect := widget.NewSelect([]string{}, func(s string) {})
	priceMapSelect := widget.NewSelect([]string{}, func(s string) {})
	skuMapSelect := widget.NewSelect([]string{}, func(s string) {})
	
	// Printer Name
	printerNameEntry := widget.NewEntry()
	printerNameEntry.SetPlaceHolder("Enter Printer Name (e.g. Zebra)")
	printerNameEntry.Text = "ZDesigner GK420d" // Default suggestion

	// Quantity
	qtyEntry := widget.NewEntry()
	qtyEntry.SetPlaceHolder("1")
	qtyEntry.Text = "1"

	// Status Log
	statusLabel := widget.NewLabel("Ready")

	// File Button
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
			
			// Update Selects
			itemMapSelect.Options = csvHeaders
			priceMapSelect.Options = csvHeaders
			skuMapSelect.Options = csvHeaders
			
			itemMapSelect.Refresh()
			priceMapSelect.Refresh()
			skuMapSelect.Refresh()
			
			// Auto-select if matches found
			for _, h := range csvHeaders {
				if h == "Item Name" || h == "item_name" { itemMapSelect.SetSelected(h) }
				if h == "Price" || h == "price" { priceMapSelect.SetSelected(h) }
				if h == "SKU" || h == "sku_id" { skuMapSelect.SetSelected(h) }
			}

			statusLabel.SetText(fmt.Sprintf("Loaded %d records", len(csvData)))
		}, w)
		
		fd.SetFilter(storage.NewExtensionFileFilter([]string{".csv"}))
		fd.Show()
	})

	// Print Button
	printBtn := widget.NewButton("Print Labels", func() {
		if len(csvData) == 0 {
			dialog.ShowError(fmt.Errorf("no data loaded"), w)
			return
		}
		
		pName := printerNameEntry.Text
		if pName == "" {
			dialog.ShowError(fmt.Errorf("printer name required"), w)
			return
		}
		
		qty, _ := strconv.Atoi(qtyEntry.Text)
		if qty < 1 { qty = 1 }

		itemNameCol := itemMapSelect.Selected
		priceCol := priceMapSelect.Selected
		skuCol := skuMapSelect.Selected
		
		if itemNameCol == "" || priceCol == "" || skuCol == "" {
			dialog.ShowError(fmt.Errorf("please map all columns"), w)
			return
		}

		// Find indices
		itemIdx, priceIdx, skuIdx := -1, -1, -1
		for i, h := range csvHeaders {
			if h == itemNameCol { itemIdx = i }
			if h == priceCol { priceIdx = i }
			if h == skuCol { skuIdx = i }
		}

		successCount := 0
		for _, row := range csvData {
			replacements := map[string]string{
				"item_name": row[itemIdx],
				"price":     row[priceIdx],
				"sku_id":    row[skuIdx],
			}
			
			zpl := ProcessTemplate(templateContent, replacements)
			
			// Send to printer qty times
			for i := 0; i < qty; i++ {
				err := SendToPrinter(pName, []byte(zpl))
				if err != nil {
					statusLabel.SetText("Error: " + err.Error())
					return
				}
			}
			successCount++
		}
		statusLabel.SetText(fmt.Sprintf("Successfully printed %d labels", successCount))
		dialog.ShowInformation("Success", "Printing completed!", w)
	})

	// Layout
	content := container.NewVBox(
		widget.NewLabelWithStyle("Ace Barcode Printer", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		fileBtn,
		fileLabel,
		widget.NewSeparator(),
		widget.NewLabel("Column Mapping:"),
		container.NewGridWithColumns(2, widget.NewLabel("Item Name:"), itemMapSelect),
		container.NewGridWithColumns(2, widget.NewLabel("Price:"), priceMapSelect),
		container.NewGridWithColumns(2, widget.NewLabel("SKU ID:"), skuMapSelect),
		widget.NewSeparator(),
		widget.NewLabel("Print Settings:"),
		container.NewGridWithColumns(2, widget.NewLabel("Printer Name:"), printerNameEntry),
		container.NewGridWithColumns(2, widget.NewLabel("Quantity per Item:"), qtyEntry),
		widget.NewSeparator(),
		printBtn,
		statusLabel,
	)

	w.SetContent(content)
	w.ShowAndRun()
}

func loadDefaultTemplate() string {
    // Basic default if file read fails, though ideally we read the .prn file
    // For now we hardcode the content from the user's file or read it if present
    content, err := os.ReadFile("ace_label_template.prn")
    if err == nil {
        return string(content)
    }
    return `^XA^FDError Loading Template^FS^XZ`
}
