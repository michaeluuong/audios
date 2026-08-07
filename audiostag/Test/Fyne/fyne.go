package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type AppState struct {
	textField *widget.Entry
}

func main() {
	switch strings.ToLower(os.Args[1]) {
	case "table":
		Table()
	case "textentry":
		TextEntry()
	case "selectentry":
		SelectEntry()

	}

}

func Table() {
	fmt.Printf("Table()\n")
	myApp := app.New()
	myWindow := myApp.NewWindow("Table Selection Example")

	// 1. Your underlying data source (a 2D slice)
	data := [][]string{
		{"Row 1, Col 1", "Row 1, Col 2"},
		{"Row 2, Col 1", "Row 2, Col 2"},
		{"Row 3, Col 1", "Row 3, Col 2"},
	}

	// 2. Initialize the table
	table := widget.NewTable(
		// Length callback: returns rows and cols
		func() (int, int) {
			return len(data), len(data[0])
		},
		// CreateCell callback: use simple Label so clicks pass through
		func() fyne.CanvasObject {
			return widget.NewLabel("Wide content")
		},
		// UpdateCell callback: bind data
		func(id widget.TableCellID, cell fyne.CanvasObject) {
			label := cell.(*widget.Label)
			label.SetText(data[id.Row][id.Col])
		},
	)

	// 3. Define the OnSelected trigger
	table.OnSelected = func(id widget.TableCellID) {
		// Because the table doesn't hold data, query your slice
		selectedValue := data[id.Row][id.Col]
		fmt.Printf("Selected Cell [%d, %d]: %s\n", id.Row, id.Col, selectedValue)
	}

	myWindow.SetContent(container.NewScroll(table))
	myWindow.Resize(fyne.NewSize(400, 200))
	myWindow.ShowAndRun()

}

func TextEntry() {
	fmt.Printf("TextEntry()\n")
	myApp := app.New()
	myWindow := myApp.NewWindow("Toolbar Example")

	textEntry := widget.NewEntry()
	state := &AppState{
		textField: textEntry,
	}

	// Create toolbar with actions, separator, and spacer
	toolbar := widget.NewToolbar(
		widget.NewToolbarAction(theme.DocumentCreateIcon(), func() { log.Println("new " + state.textField.Text) }),
		widget.NewToolbarAction(theme.SearchReplaceIcon(), func() {
			log.Println("replace " + state.textField.Text)
			textEntry.SetText("")
		}),
		widget.NewToolbarSeparator(),
		widget.NewToolbarSpacer(), // Pushes following items to the right
		widget.NewToolbarAction(theme.HelpIcon(), func() { log.Println("Help") }),
	)

	// Layout with toolbar at the top
	//content := widget.NewLabel("Content Area")
	myWindow.SetContent(container.NewBorder(toolbar, nil, nil, nil, textEntry))
	myWindow.Resize(fyne.NewSize(400, 200))
	myWindow.ShowAndRun()
}

func SelectEntry() {
	fmt.Printf("SelectEntry()\n")
	myApp := app.New()
	myWindow := myApp.NewWindow("SelectEntry Widget")

	combo := widget.NewSelectEntry([]string{"Red", "Green", "Blue"})
	combo.PlaceHolder = "Choose or type a colour"
	combo.OnChanged = func(value string) {
		log.Println("clicked: ", value)
	}
	combo.OnSubmitted = func(value string) {
		log.Println("value: ", value)
	}

	myWindow.SetContent(combo)
	myWindow.Resize(fyne.NewSize(250, 100))
	myWindow.ShowAndRun()
}
