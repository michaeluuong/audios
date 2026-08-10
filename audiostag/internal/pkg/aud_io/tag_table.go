package aud_io

import (
	"fmt"
	_ "image/jpeg"
	_ "image/png"
	"log/slog"
	"slices"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"github.com/kbinani/screenshot"
)

func NewTagTableData(data [][]string) *TagTableData {
	t := &TagTableData{Data: data}
	t.Headers = make([]string, len(t.Data[0][1:]))
	copy(t.Headers, t.Data[0][1:])

	t.Rows = make([][]string, len(t.Data[1:]))
	for i, d := range t.Data[1:] {
		t.Rows[i] = make([]string, len(d[1:]))
		copy(t.Rows[i], d[1:])

	}

	return t

}

type TagTableData struct {
	Headers []string
	Rows    [][]string
	Data    [][]string
}

func (t *TagTableData) Len() int {
	return len(t.Headers) + len(t.Rows)

}

func (t *TagTableData) GetText(selectedCell widget.TableCellID) (string, string) {
	row := selectedCell.Row
	col := selectedCell.Col

	return t.Headers[col], t.Rows[row][col]

}

func (t *TagTableData) preferredWindowSize() fyne.Size {
	headers := t.Headers
	rows := t.Rows
	if len(headers) == 0 {
		return fyne.NewSize(360, 1000)

	}

	colWidths := t.measureColumnWidths()
	var totalW float32
	for _, w := range colWidths {
		totalW += w

	}
	totalW += (2.0 * float32(len(rows[0]))) // bit o extra

	rowH := fyne.MeasureText("Cd", theme.TextSize(), fyne.TextStyle{}).Height + theme.InnerPadding()*2
	// header row + data rows + window chrome (filePath label, close button, padding)
	chromeH := rowH*2 + theme.Padding()*8 + 80
	//totalH := rowH*float32(len(rows)+1) + chromeH
	const estTextEntry float32 = 100.00
	//var imageHeight float32 = 75.0
	totalH := rowH*float32(len(rows)+1) + chromeH + estTextEntry // + imageHeight

	ptLen := float32(t.Len())
	fmt.Printf("totalW: %f, totalH: %f, ptLen: %f\n", totalW, totalH, ptLen)
	dWidth, dHeight := DesktopDimensions()
	var (
		maxW float32 = dWidth
		maxH float32 = dHeight
		//minW float32 = width / ptLen
		//minH float32 = height / ptLen
		minW float32 = dWidth / ptLen
		minH float32 = 10
	)

	fmt.Printf("minW: %f, minH: %f, maxW: %f, maxH: %f\n", minW, minH, maxW, maxH)
	if totalW < minW {
		totalW = minW

	}

	if totalH < minH {
		totalH = minH

	}

	if totalW > maxW {
		totalW = maxW

	}

	if totalH > maxH {
		totalH = maxH

	}

	// room for scrollbars when content is capped
	totalW += theme.ScrollBarSize() + 40
	totalH += theme.ScrollBarSize() + 20
	fmt.Printf("totalW: %f, totalH: %f\n", totalW, totalH)

	return fyne.NewSize(totalW, totalH)

}

// measureColumnWidths measures width for cell size
func (t *TagTableData) measureColumnWidths() []float32 {
	pad := theme.InnerPadding()*2 + theme.Padding()*2
	style := fyne.TextStyle{}
	bold := fyne.TextStyle{Bold: true}
	textSize := theme.TextSize()
	widths := make([]float32, len(t.Headers))

	for i, h := range t.Headers {
		widths[i] = fyne.MeasureText(h, textSize, bold).Width + pad

	}

	for _, row := range t.Rows {
		for i, cell := range row {
			if i >= len(widths) {
				break

			}

			w := fyne.MeasureText(cell, textSize, style).Width + pad

			if w > widths[i] {
				widths[i] = w

			}

		}

	}

	const minCol float32 = 48
	for i := range widths {
		if widths[i] < minCol {
			widths[i] = minCol

		}

	}

	return widths

}

type TagTableDataAppState struct {
	selectedCell *widget.TableCellID
	selectedText string
	headerText   string
}

// ShowPipeTableWindow opens a resizable window with a scrollable, selectable table.
// It sizes itself to the content (capped by the screen) but remains user-resizable.
//   - filePath is the full path to the data
//   - data is a table of all tags to display
//   - cImg is the cover image
//   - win is the window to display this table
//   - showActions is true to display tagging options or false if display only
//   - hasCompleted is true if a round of tagging has already been completed false otherwise
func ShowTagTable(filePath string, data [][]string, cImg CoverImage, win fyne.Window, showActions bool, hasCompleted bool) {
	//t := &TagTableData{Headers: data[0][1:], Rows: data[1:], Data: data}
	t := NewTagTableData(data)

	fyne.DoAndWait(func() {
		//t.Headers, t.Rows = data[0], data[1:]
		win.SetContent(buildTagTable(win, filePath, t, cImg, showActions, hasCompleted))

		win.Resize(t.preferredWindowSize())
		win.CenterOnScreen()
		win.Show()
	})

}

func buildTagTable(win fyne.Window, filePath string, tagTableData *TagTableData, cImg CoverImage, showActions bool, hasCompleted bool) fyne.CanvasObject {
	if len(tagTableData.Headers) == 0 {
		tagTableData.Headers = []string{""}

	}

	table := widget.NewTable(
		func() (int, int) {
			return len(tagTableData.Rows), len(tagTableData.Headers)

		},
		func() fyne.CanvasObject {
			l := widget.NewLabel("Cell Data")
			//l.Selectable = true
			l.Truncation = fyne.TextTruncateOff

			return l

		},
		func(id widget.TableCellID, o fyne.CanvasObject) {
			l := o.(*widget.Label)
			if id.Row < 0 || id.Row >= len(tagTableData.Rows) || id.Col < 0 || id.Col >= len(tagTableData.Rows[id.Row]) {
				l.SetText("")
				return

			}

			/*if id.Row == 0 && tagTableData.Rows[id.Row][id.Col] == tagTableData.Headers[id.Col] {
				l.TextStyle = fyne.TextStyle{Bold: true}

			}*/

			l.SetText(tagTableData.Rows[id.Row][id.Col])

		},
	)

	table.ShowHeaderRow = true
	table.CreateHeader = func() fyne.CanvasObject {
		l := widget.NewLabel("Header")
		l.TextStyle = fyne.TextStyle{Bold: true}
		l.Selectable = true
		l.Truncation = fyne.TextTruncateOff

		return l

	}

	table.UpdateHeader = func(id widget.TableCellID, o fyne.CanvasObject) {
		l := o.(*widget.Label)
		if id.Row < 0 && id.Col >= 0 && id.Col < len(tagTableData.Headers) {
			l.SetText(tagTableData.Headers[id.Col])

			return

		}

		l.SetText("")

	}

	/*colWidths := tagTableData.measureColumnWidths()
	for i, width := range colWidths {
		table.SetColumnWidth(i, width)

	}*/
	setColumnWidths(tagTableData, table)

	heading := widget.NewLabel(filePath)
	heading.TextStyle = fyne.TextStyle{Bold: true}
	heading.Selectable = true

	fmt.Printf("buildTagTable()|cImg.Filename: %s\n", cImg.Filename)
	//image := "/Users/irenepiechota-wong/DEV/GO/audios/audiostag/audiofiles/cover.jpg"
	fileCheck := widget.NewCheck("Show File", func(checked bool) {
		if checked {
			tagTableData.Headers = slices.Insert(tagTableData.Headers, 0, tagTableData.Data[0][0])
			for i, d := range tagTableData.Data[1:] {
				tagTableData.Rows[i] = slices.Insert(tagTableData.Rows[i], 0, d[0])

			}

		} else {
			tagTableData.Headers = tagTableData.Headers[1:]
			for i, row := range tagTableData.Rows {
				tagTableData.Rows[i] = row[1:]

			}

		}

		/*colWidths := tagTableData.measureColumnWidths()
		for i, width := range colWidths {
			table.SetColumnWidth(i, width)

		}*/
		//table.Refresh()
		//dWidth, dHeight := DesktopDimensions()
		setColumnWidths(tagTableData, table)
		win.Hide()
		win.Resize(tagTableData.preferredWindowSize())
		win.CenterOnScreen()
		win.Show()

	})
	headingContainer := container.NewBorder(heading, fileCheck, nil, nil)

	//var topContainer *fyne.Container = container.NewBorder(nil, nil, heading, nil)
	var topContainer *fyne.Container = container.NewBorder(nil, nil, headingContainer, nil)

	imgContainer := coverImage(win, cImg)
	if imgContainer != nil {
		topContainer = container.NewBorder(nil, nil, headingContainer, imgContainer, nil)
		//topContainer = container.NewBorder(nil, c, heading, imgContainer, nil)

	}

	tagTableDataAppState := &TagTableDataAppState{}
	closeBtn := widget.NewButton("Close", func() {
		slog.Debug("widget.NewButton()|close button clicked")
		win.Close()

	})
	closeBtn.Importance = widget.MediumImportance

	var bottomContainer *fyne.Container
	var selectEntry *ActionSelectEntry
	if showActions {
		//selectEntry = SelectEntry(win, filePath, tagTableDataAppState)
		selectEntry = ActionSelect(win, filePath, hasCompleted)
		selectEntryClose := container.NewBorder(selectEntry.actionContainer, closeBtn, nil, nil, nil)
		bottomContainer = container.NewPadded(selectEntryClose)

	} else {
		bottomContainer = container.NewPadded(closeBtn)

	}

	table.OnSelected = func(id widget.TableCellID) {
		tagTableDataAppState.selectedCell = &id
		tagTableDataAppState.headerText, tagTableDataAppState.selectedText = tagTableData.GetText(id)
		fmt.Printf("buildTagTable()|fileCheck.Checked: %t, id.Col: %d\n", fileCheck.Checked, id.Col)
		/*if fileCheck.Checked && id.Col == 0 {
			return

		}*/

		if selectEntry != nil {
			selectEntry.UpdateEntries(tagTableDataAppState.headerText, tagTableDataAppState.selectedText)

		}

	}
	/*table.OnUnselected = func(id widget.TableCellID) {
		tableState.selectedCell = nil
	}*/

	// Table provides its own vertical and horizontal scroll bars.
	return container.NewBorder(
		topContainer,
		//container.NewPadded(selectEntryClose),
		bottomContainer,
		nil,
		nil,
		table,
	)

}

func setColumnWidths(tagTableData *TagTableData, table *widget.Table) {
	colWidths := tagTableData.measureColumnWidths()
	for i, width := range colWidths {
		table.SetColumnWidth(i, width)

	}

}

func coverImage(win fyne.Window, cImg CoverImage) *fyne.Container {
	var imgContainer *fyne.Container = nil

	//cImg := &CoverImage{Filename: "tokyo tayucky"}

	if cImg.Image != nil {
		//canvasImage := canvas.NewImageFromImage(img)
		canvasImage := canvas.NewImageFromImage(cImg.Image)
		canvasImage.FillMode = canvas.ImageFillContain
		canvasImage.SetMinSize(fyne.NewSize(75, 75))
		fmt.Printf("coverImage()|cImg.Filename: %s\n", cImg.Filename)
		imgButton := widget.NewButton("", func() {
			//showEnlargedImage(win, img)
			showEnlargedImage(win, cImg)

		})
		imgButton.Importance = widget.LowImportance

		imgContainer = container.NewStack(canvasImage, imgButton)

	}

	return imgContainer

}

// func showEnlargedImage(parent fyne.Window, img image.Image) {
func showEnlargedImage(parent fyne.Window, cImg CoverImage) {
	img := cImg.Image
	//popupImg := canvas.NewImageFromImage(img)
	popupImg := canvas.NewImageFromImage(img)
	popupImg.FillMode = canvas.ImageFillOriginal
	//popupImgSize := popupImg.Size()
	imgDimensions := fmt.Sprintf("%dx%d", img.Bounds().Dx(), img.Bounds().Dy())
	//popupImg.SetMinSize(popupImgSize)

	dialog.ShowCustom("Original Artwork "+imgDimensions+" "+cImg.Filename, "Close", popupImg, parent)

}

func DesktopDimensions() (float32, float32) {
	bounds := screenshot.GetDisplayBounds(0)
	pixelwidth := float32(bounds.Dx())
	pixelHeight := float32(bounds.Dy())

	return pixelwidth, pixelHeight

}

func ActionSelect(win fyne.Window, filePath string, hasCompleted bool) *ActionSelectEntry {
	actions := []*ActionEntry{
		{
			name:         "Album Folder",
			action:       "--album-folder-name",
			placeHolders: []placeHolder{{desc: "folder name"}},
			isSingle:     true,
		}, {
			name:         "Cover Album",
			action:       "--cover-album",
			placeHolders: []placeHolder{{desc: "cover search album"}},
			isSingle:     true,
		}, {
			name:         "Cover Artist",
			action:       "--cover-artist",
			placeHolders: []placeHolder{{desc: "cover search artist"}},
			isSingle:     true,
		}, {
			name:         "Dummy Files",
			action:       "--dummy-files",
			placeHolders: []placeHolder{{desc: "true/false"}},
			isBool:       true,
			isNotAlbum:   true,
		}, {
			name:         "Exclude File",
			action:       "--extract-exclude",
			placeHolders: []placeHolder{{desc: "exclude file"}},
			isSingle:     true,
		}, {
			name:         "Keep Tag",
			action:       "--keep-tag",
			placeHolders: []placeHolder{{desc: "keep tag", isHead: true}},
		}, {
			name:         "Playlist Name",
			action:       "--playlist-name",
			placeHolders: []placeHolder{{desc: "playlist name"}},
			isSingle:     true,
		}, {
			name:   "Precondition",
			action: "--precondition",
			placeHolders: []placeHolder{
				{desc: "src tag", isHead: true},
				{desc: "regex"},
				{desc: "dest tag", isHead: true},
				{desc: "match value"},
				{desc: "else value"},
			},
		}, {
			name:   "Prereplace",
			action: "--prereplace",
			placeHolders: []placeHolder{
				{desc: "tag", isHead: true},
				{desc: "regex"},
				{desc: "replace value"},
			},
		}, {
			name:   "Replace",
			action: "--replace",
			placeHolders: []placeHolder{
				{desc: "tag", isHead: true},
				{desc: "regex"},
				{desc: "replace value"},
			},
		}, {
			name:         "Single Artist",
			action:       "--single-artist",
			placeHolders: []placeHolder{{desc: "true/false"}},
			isBool:       true,
		}, {
			name:   "Tag",
			action: "--tag",
			placeHolders: []placeHolder{
				{desc: "tag", isHead: true},
				{desc: "value"},
				{desc: "optional value"},
			},
		}, {
			name:         "Various Artists",
			action:       "--various",
			placeHolders: []placeHolder{{desc: "true/false"}},
			isBool:       true,
		},
	}

	actionSelect := NewActionSelectEntry(win, filePath, actions, "Choose or type", hasCompleted)

	return actionSelect

}
