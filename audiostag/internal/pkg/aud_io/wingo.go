package aud_io

import (
	"fmt"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

func PrintTable(pipeData string) {

}

type customRowLayout struct {
	colWidths []float32
}

func (l *customRowLayout) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	var currentX float32
	for i, obj := range objects {
		if i >= len(l.colWidths) {
			break

		}
		// Force the exact calculated width and match the container height
		obj.Resize(fyne.NewSize(l.colWidths[i], size.Height))
		obj.Move(fyne.NewPos(currentX, 0))
		currentX += l.colWidths[i] + 10 // 10px spacing between columns

	}

}

func (l *customRowLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
	var totalWidth float32
	var maxHeight float32

	for i, obj := range objects {
		if i >= len(l.colWidths) {
			break
		}
		totalWidth += l.colWidths[i]
		if i < len(objects)-1 {
			totalWidth += 10 // account for spacing
		}
		if obj.MinSize().Height > maxHeight {
			maxHeight = obj.MinSize().Height
		}
	}
	return fyne.NewSize(totalWidth, maxHeight)
}

func PrintGridWithColumns(pipeString string) {
	// 1. Initialize App FIRST
	a := app.NewWithID("michaeluuong.audios.audiostag")
	w := a.NewWindow("Audiostag")

	// Split pipeString into a 2D slice
	lines := strings.Split(pipeString, "\n")
	var gridData [][]string
	maxCols := 0
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue

		}

		parts := strings.Split(line, "|")
		gridData = append(gridData, parts)
		if len(parts) > maxCols {
			maxCols = len(parts)

		}

	}

	// Get the minimum width of each column
	colWidths := make([]float32, maxCols)
	for _, row := range gridData {
		for colIdx, text := range row {
			if colIdx >= len(colWidths) {
				continue

			}

			labelWidth := widget.NewLabel(text).MinSize().Width
			if labelWidth > colWidths[colIdx] {
				colWidths[colIdx] = labelWidth

			}

		}

	}

	// Generate structured rows using our custom layout
	rowLayout := &customRowLayout{colWidths: colWidths}

	// Static header row
	var headerRow []fyne.CanvasObject
	if len(gridData) > 0 {
		for _, text := range gridData[0] {
			label := widget.NewLabel(text)
			label.Selectable = true
			headerRow = append(headerRow, label)

		}

	}
	headerContainer := container.New(rowLayout, headerRow...)

	var rows []fyne.CanvasObject
	for _, row := range gridData[1:] {
		var rowItems []fyne.CanvasObject
		for _, text := range row {
			label := widget.NewLabel(text)
			label.Selectable = true
			rowItems = append(rowItems, label)

		}

		// Create a container with our explicit column layout rule
		rowContainer := container.New(rowLayout, rowItems...)
		rows = append(rows, rowContainer)

	}

	// 5. Stack rows vertically and attach the close button
	vBox := container.NewVBox(rows...)
	scrollContainer := container.NewHScroll(vBox)

	closeBtn := widget.NewButton("Close", func() { w.Close() })

	mainLayout := container.NewBorder(headerContainer, closeBtn, nil, nil, scrollContainer)
	//borderLayout := container.NewBorder(nil, nil, nil, nil, closeBtn)
	//mainLayout := container.NewHScroll(container.NewHBox(scrollContainer))
	//mainLayout := container.NewHScroll(scrollContainer)

	w.SetContent(mainLayout)

	headerHeight := headerContainer.MinSize().Height
	bodyHeight := vBox.MinSize().Height
	btnHeight := closeBtn.MinSize().Height

	// Sum the full content requirements plus margins
	//totalWidth := vBox.MinSize().Width
	totalWidth := vBox.MinSize().Width
	totalHeight := headerHeight + bodyHeight + btnHeight + 20

	// Cap the window size gracefully if there are too many rows
	totalHeight = min(1000, totalHeight)
	totalWidth = min(1000, totalWidth)
	totalWidth = 400

	// Window resizes perfectly to frame the entire content block
	//w.Resize(mainLayout.MinSize().Add(fyne.NewSize(20, 40)))
	w.Resize(fyne.NewSize(totalWidth, totalHeight))
	w.ShowAndRun()

}

func PrintNewGridWithColumns(pipeString string) {
	// Parse lines and determine column count
	lines := strings.Split(pipeString, "\n")
	colsCount := 0
	var cells []fyne.CanvasObject

	for _, line := range lines {
		parts := strings.Split(line, "|")
		if len(parts) > colsCount {
			colsCount = len(parts)

		}

		for _, part := range parts {
			// Using Bold for headers, regular for data
			label := widget.NewLabel(part)
			cells = append(cells, label)

		}

	}

	a := app.New()
	w := a.NewWindow("Audiostag")

	gridLayout := container.NewGridWithColumns(colsCount, cells...)

	closeBtn := widget.NewButton("Close", func() { w.Close() })

	mainLayout := container.NewBorder(nil, closeBtn, nil, nil, gridLayout)
	w.SetContent(mainLayout)

	w.Resize(mainLayout.MinSize())
	w.ShowAndRun()

}

func PrintGrid(pipeString string) {
	//newApp := app.New()
	app := app.NewWithID("michaeluuong.audios.audiostag")
	window := app.NewWindow("Audiostag")

	// 1. Parse the string into a 2D slice
	var tableData [][]string
	lines := strings.Split(pipeString, "\n")
	//var filename string
	for _, line := range lines {
		cells := strings.Split(line, "|")
		//fmt.Printf("line: %s\n", line)
		/*if len(cells) == 1 {
			filename = cells[0]
			continue

		}*/

		tableData = append(tableData, cells)

	}
	//window := app.NewWindow(filepath.Base(filename))

	// 2. Define the Fyne Table widget
	// It requires three callbacks: Length, CreateCell, and UpdateCell
	table := widget.NewTable(
		func() (int, int) {
			// Length: returns (rows, columns)
			if len(tableData) == 0 {
				return 0, 0

			}

			return len(tableData), len(tableData[0])

		},

		func() fyne.CanvasObject {
			// CreateCell: template for how a cell looks
			return widget.NewLabel("Cell Content")

		},

		func(id widget.TableCellID, o fyne.CanvasObject) {
			// UpdateCell: updates the cell's text based on its Row and Col
			label := o.(*widget.Label)
			label.SetText(tableData[id.Row][id.Col])

		},
	)

	for col := range tableData[0] {
		var maxW float32
		for row := range tableData {
			if w := widget.NewLabel(tableData[row][col]).MinSize().Width; w > maxW {
				maxW = w

			}

		}

		table.SetColumnWidth(col, maxW+5)

	}
	//textSize, _ := fyne.CurrentApp().Driver().RenderedTextSize(lines[0], theme.TextSize(), fyne.TextStyle{}, nil)
	//table.SetColumnWidth(0, textSize.Width)

	closeButton := widget.NewButton("Close", func() {
		window.Close()

	})

	content := container.NewBorder(nil, closeButton, nil, nil, table)

	window.SetContent(content)
	//window.Resize(fyne.NewSize(400, 200))
	window.Resize(table.MinSize().Add(fyne.NewSize(400, 200)))
	window.ShowAndRun()

}

func PrintWindow1(text string) {
	myApp := app.NewWithID("michaeluuong.audios.audiostag")

	outWindow := myApp.NewWindow("output")

	textGrid := widget.NewTextGrid()
	textGrid.SetText(text)
	textGrid.ShowLineNumbers = false
	textGrid.ShowWhitespace = false

	scrollContainer := container.NewScroll(textGrid)
	sMin := scrollContainer.MinSize()
	//scrollContainer.Resize(fyne.NewSize(sMin.Width, sMin.Height))
	fmt.Printf("sMin.Height: %v, sMin.Width: %v\n", sMin.Height, sMin.Width)

	closeButton := widget.NewButton("Close", func() {
		outWindow.Close()
	})
	//content := container.NewVBox(textWidget, closeButton)

	borderContainer := container.NewBorder(nil, closeButton, nil, nil, scrollContainer)
	//borderContainer := container.NewBorder(nil, closeButton, nil, nil, textWidget)

	/*outWindow.SetCloseIntercept(func() {
		outWindow.Hide()
	})*/

	//scrollContent.SetMinSize(fyne.NewSize(500, 400))
	//d := dialog.NewCustom("Multi-line Output", "Close", scrollContent, outWindow)
	//d.Show()
	outWindow.SetContent(borderContainer)
	minSize := borderContainer.MinSize()
	fmt.Printf("minSize.Height: %v, minSize.Width: %v\n", minSize.Height, minSize.Width)
	outWindow.Resize(fyne.NewSize(minSize.Width+5, minSize.Height+5))
	outWindow.ShowAndRun()

}

func PrintWindow2(text string) {
	myApp := app.NewWithID("michaeluuong.audios.audiostag")

	outWindow := myApp.NewWindow("output")
	outWindow.Resize(fyne.NewSize(500, 500))

	//textCanvas := canvas.NewText(text, color.White)
	textWidget := widget.NewRichText(
		&widget.TextSegment{
			Style: widget.RichTextStyle{
				ColorName: theme.ColorNameForeground, // Uses the primary readable text color
				//SizeName:  theme.SizeNameHeadingText, // Scales the text up to a larger heading size
				//TextStyle: fyne.TextStyle{Bold: true},
			},
			Text: text,
		},
	)
	//textCanvas.TextSize = 50
	textWidget.Wrapping = fyne.TextWrapWord

	//textWidget := widget.NewMultiLineEntry()
	//textWidget.SetText(text)
	//textWidget.Disable()
	scrollContainer := container.NewScroll(textWidget)
	//scrollContainer.objects[0].(*widget.Entry).SetText(text)

	closeButton := widget.NewButton("Close", func() {
		outWindow.Close()
	})
	//content := container.NewVBox(textWidget, closeButton)

	borderContainer := container.NewBorder(nil, closeButton, nil, nil, scrollContainer)

	/*outWindow.SetCloseIntercept(func() {
		outWindow.Hide()
	})*/

	//scrollContent.SetMinSize(fyne.NewSize(500, 400))
	//d := dialog.NewCustom("Multi-line Output", "Close", scrollContent, outWindow)
	//d.Show()
	outWindow.SetContent(borderContainer)
	outWindow.ShowAndRun()

}
