package aud_io

import (
	"fmt"
	"log/slog"
	"maps"
	"os"
	"os/exec"
	"slices"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"github.com/michaeluuong/utilize/reflections"

	mp3Config "github.com/michaeluuong/audios/audiostag/internal/app/core/mp3/config"
)

// ActionSelectEntry contains an ActionSelect followed by text fields, Submit and Delete Last buttons
type ActionSelectEntry struct {
	currentActionName string
	actionEntries     *ActionEntries
	submitButton      *widget.Button
	deleteButton      *widget.Button
	updateEntry       *widget.Entry
	actionContainer   *fyne.Container
	widget.SelectEntry
}

func (a *ActionSelectEntry) MinSize() fyne.Size {
	// Base the size on the placeholder text width and standard padding
	pad := theme.Padding() * 4
	pad *= 4
	return fyne.NewSize(fyne.MeasureText(a.PlaceHolder, theme.TextSize(), fyne.TextStyle{}).Width+pad, a.SelectEntry.MinSize().Height)

}

func (a *ActionSelectEntry) UpdateEntries(header, data string) {
	if a.currentActionName == "" {
		return

	}

	action := a.actionEntries.nameToAction[a.currentActionName]
	entries := action.entries
EntryLoop:
	for i, entry := range entries {
		switch o := entry.(type) {
		case *widget.Check:
			o.SetChecked(true)

		case *widget.Entry:
			//lastEntry, ok := entries[len(entries)-1].(*widget.Entry)
			//if o.Text == "" || (ok && lastEntry.HasValue()) {
			if o.Text == "" || o.Text == o.PlaceHolder {
				isHead := action.placeHolders[i].isHead
				if isHead {
					o.SetText(header)

					if len(entries) > i+1 {
						if o2, ok := entries[i+1].(*widget.Entry); ok {
							o2.SetText(data)

						}

					}

				} else {
					o.SetText(data)

				}

				break EntryLoop

			} else if action.isSingle {
				o.Append(data)

			}

		case *widget.SelectEntry:
			fmt.Printf("UpdateEntries()|o.Text: %s, header: %s, data: %s\n", o.Text, header, data)
			if o.Text == "" || o.Text == o.PlaceHolder {
				isHead := action.placeHolders[i].isHead
				if isHead {
					if tagName, ok := mp3Config.IDToNameMod[header]; ok {
						o.SetText(tagName)

						/*} else if header == "File" {
						o.SetText(header)*/

					}

					if len(entries) > i+1 {
						if o2, ok := entries[i+1].(*widget.Entry); ok {
							o2.SetText(data)

						}

					}

				} else {
					o.SetText(data)

				}

				break EntryLoop

			}

		}

	}

}

func (a *ActionSelectEntry) submissionsToFlags() []string {
	var allFlags []string
	for actionName, submissions := range a.actionEntries.nameToSubmissions {
		//subText := strings.Join(*submissions, "~")
		subText := strings.Join(*submissions, mp3Config.TagSeparator)
		action := a.actionEntries.nameToAction[actionName].action
		allFlags = append(allFlags, action)
		allFlags = append(allFlags, subText)
		/*for _, flag := range *submissions {
			allFlags = append(allFlags, "\""+flag+"\"")

		}*/

	}

	return allFlags

}

func (a *ActionSelectEntry) execute(win fyne.Window, filePath string) error {
	allFlags := []string{"--path", filePath}
	if !a.actionEntries.isNotAlbum {
		allFlags = append(allFlags, "--album")

	}

	allFlags = append(allFlags, a.submissionsToFlags()...)
	slog.Debug("flags", "allFlags", allFlags)

	executable, err := os.Executable()
	if err != nil {
		return err

	}

	cmd := exec.Command(executable, allFlags...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	err = cmd.Start()
	if err != nil {
		return err

	}
	win.Close()
	os.Exit(0)

	return nil

}

func (a *ActionSelectEntry) setOnChanged() {
	a.OnChanged = func(value string) {
		a.currentActionName = value
		actionEntries := a.actionEntries
		actionEntries.rightContainer.RemoveAll()
		if actionEntry, ok := actionEntries.nameToAction[value]; ok {
			entries := actionEntry.entries
			for _, entry := range entries {
				switch o := entry.(type) {
				case *widget.Check:
					//if actionEntry.isBool {
					o.SetChecked(true)

				case *widget.Entry:
					//} else {
					//o.SetText("")
					o.SetPlaceHolder(o.PlaceHolder)

				case *widget.SelectEntry:
					//o.SetText(o.PlaceHolder)
					o.SetPlaceHolder(o.PlaceHolder)

				}

				actionEntries.rightContainerAdd(entry)

			}

			actionEntries.rightContainerAdd(a.submitButton)

		}

		actionEntries.rightContainerAdd(a.deleteButton)

	}

}

func (a *ActionSelectEntry) setSubmitButton() {
	a.submitButton = widget.NewButton("Submit", func() {
		currentActionName := a.currentActionName
		submission := fmt.Sprintf("%s", a.actionEntries.nameToAction[currentActionName])
		fmt.Printf("setSubmitButton()|submission: %v\n", submission)
		actionEntries := a.actionEntries
		//fmt.Printf("action: %s, submission: %v, submissions: %v\n", action, submission, submissions)
		//*submissions = append(*submissions, submission)
		if submission != "" {
			actionEntries.addSubmission(currentActionName, submission)

		}

		a.updateEntry.SetText("")
		for actionName, submissions := range actionEntries.nameToSubmissions {
			//subText := strings.Join(*submissions, "~")
			subText := strings.Join(*submissions, mp3Config.TagSeparator)

			action := actionEntries.nameToAction[actionName].action
			if subText != "" {
				a.updateEntry.Append(action + " \"" + subText + "\" ")

			}

		}

		entries := actionEntries.nameToAction[currentActionName].entries
		for _, entry := range entries {
			switch o := entry.(type) {
			case *widget.Check:
				//if actionEntries.nameToAction[currentActionName].isBool {
				o.SetChecked(true)

			case *widget.Entry:
				o.SetText("")
				o.SetPlaceHolder(o.PlaceHolder)

			case *widget.SelectEntry:
				o.SetText("")
				o.SetPlaceHolder(o.PlaceHolder)

			}

		}

	})

}

func (a *ActionSelectEntry) setDeleteButton() {
	a.deleteButton = widget.NewButton("Delete", func() {
		currentActionName := a.currentActionName
		actionEntries := a.actionEntries
		actionEntries.deleteSubmission(currentActionName)

		//subText := strings.Join(actionEntries.submissions(currentActionName), "~")
		if actionEntry, ok := actionEntries.nameToAction[currentActionName]; ok {
			for _, entry := range actionEntry.entries {
				switch o := entry.(type) {
				case *widget.Check:
					o.SetChecked(true)

				case *widget.Entry:
					o.SetText("")
					o.SetPlaceHolder(o.PlaceHolder)

				case *widget.SelectEntry:
					o.SetText("")
					o.SetPlaceHolder(o.PlaceHolder)

				}

			}

		}

		a.updateEntry.SetText("")
		for actionName, submissions := range actionEntries.nameToSubmissions {
			//subText := strings.Join(*submissions, "~")
			subText := strings.Join(*submissions, mp3Config.TagSeparator)
			action := actionEntries.nameToAction[actionName].action
			if subText != "" {
				a.updateEntry.Append(action + " \"" + subText + "\" ")

			}

		}

	})

}

func NewActionSelectEntry(win fyne.Window, filePath string, actions []*ActionEntry, placeholder string) *ActionSelectEntry {
	fmt.Printf("NewActionSelectEntry()|start\n")
	updateEntry := widget.NewEntry()
	actionSelectEntry := &ActionSelectEntry{
		//updateEntry:   updateEntry,
		updateEntry: updateEntry,
	}
	actionSelectEntry.actionEntries = NewActionEntries(actions, updateEntry)

	executeBtn := widget.NewButton("Execute", func() {
		actionSelectEntry.execute(win, filePath)

	})
	updateEntryContainer := container.NewBorder(nil, nil, nil, executeBtn, updateEntry)

	var options []string
	for _, actionEntry := range actions {
		options = append(options, actionEntry.name)

	}
	actionSelectEntry.ExtendBaseWidget(actionSelectEntry)
	actionSelectEntry.SetOptions(options)
	actionSelectEntry.SetPlaceHolder(placeholder)

	actionSelectEntry.setSubmitButton()
	actionSelectEntry.setDeleteButton()
	actionSelectEntry.setOnChanged()

	//actionSelectEntry.actionContainer = container.NewBorder(nil, nil, nil, actionSelectEntry.actionEntries.rightContainer)
	actionSelectEntry.actionContainer = container.NewBorder(nil, updateEntryContainer, actionSelectEntry, nil, actionSelectEntry.actionEntries.rightContainer)

	return actionSelectEntry

}

type placeHolder struct {
	desc   string
	isHead bool
}

type ActionEntry struct {
	name         string
	action       string
	placeHolders []placeHolder
	//entries      []*widget.Entry
	//entries    []*fyne.CanvasObject
	entries    []fyne.CanvasObject
	isBool     bool
	isNotAlbum bool
	isSingle   bool
}

func (a ActionEntry) String() string {
	var sb strings.Builder
	for i, entry := range a.entries {
		switch o := entry.(type) {
		case *widget.Check:
			sb.WriteString(fmt.Sprintf("%t", o.Checked))

		case *widget.Entry:
			if o.Text != o.PlaceHolder {
				if sb.Len() > 0 {
					sb.WriteString("=")

				}

				//sb.WriteString(entry.Text)
				sb.WriteString(o.Text)
			}

		case *widget.SelectEntry:
			if o.Text != o.PlaceHolder {
				if sb.Len() > 0 {
					sb.WriteString("=")

				}

				if i <= len(a.placeHolders)-1 {
					if a.placeHolders[i].isHead {
						fmt.Printf("String()|isHead: %t\n", a.placeHolders[i].isHead)
						if id, ok := mp3Config.NameToIDMod[o.Text]; ok {
							fmt.Printf("String()|id: %s, ok: %t\n", id, ok)
							sb.WriteString(id)

						} else {
							return ""

						}

					} else {
						fmt.Printf("String()|else\n")
						sb.WriteString(o.Text)

					}

				}

			}

		}

	}

	//return a.action + " \"" + sb.String() + "\""
	return sb.String()

}

type ActionEntries struct {
	actions           []*ActionEntry
	nameToAction      map[string]*ActionEntry
	nameToSubmissions map[string]*[]string
	//submitButton      *widget.Button
	rightContainer *fyne.Container
	isNotAlbum     bool
}

func (a *ActionEntries) addSubmission(name, submission string) {
	if name == "" || submission == "" {
		return
	}

	if _, ok := a.nameToSubmissions[name]; !ok {
		a.nameToSubmissions[name] = &[]string{}

	}

	if a.nameToAction[name].isSingle && len(*a.nameToSubmissions[name]) == 1 {
		return

	}

	if a.nameToAction[name].isNotAlbum {
		a.isNotAlbum = true

	}

	if a.nameToAction[name].isBool {
		(*a.nameToSubmissions[name]) = []string{submission}

	} else {
		*a.nameToSubmissions[name] = append(*a.nameToSubmissions[name], submission)

	}

}

func (a *ActionEntries) deleteSubmission(name string) {
	if submission, ok := a.nameToSubmissions[name]; ok {
		if len(*submission) > 0 {
			*a.nameToSubmissions[name] = (*submission)[:len(*submission)-1]
			if a.nameToAction[name].isNotAlbum {
				a.isNotAlbum = false

			}

		}

	}

}

func (a *ActionEntries) submissions(action string) []string {
	if a.nameToSubmissions != nil {
		if submission, ok := (a.nameToSubmissions)[action]; ok {
			return *submission

		}

	}

	return []string{}

}

func (a *ActionEntries) SetActions(actions []*ActionEntry, updateEntry *widget.Entry) {
	a.actions = actions
	a.actionMap()
	//a.actionEntry(updateEntry)
	a.actionEntry()

}

func (a *ActionEntries) actionMap() {
	for _, actionEntry := range a.actions {
		a.nameToAction[actionEntry.name] = actionEntry

	}

}

// func (a *ActionEntries) actionEntry(updateEntry *widget.Entry) {
func (a *ActionEntries) actionEntry() {
	for _, a := range a.actions {
		for _, placeHolder := range a.placeHolders {
			if a.isBool {
				/*textField.Validator = func(text string) error {
					if text != "true" && text != "false" {
						return errors.New("must be true or false")

					}

					return nil

				}

				textField.SetText("true")*/
				/*checkLabel := widget.NewLabel("true")
				check := widget.NewCheck(placeHolder.desc, func(checked bool) {
					checkLabel.SetText(fmt.Sprintf("%t", checked))

				})*/
				check := widget.NewCheck(placeHolder.desc, func(checked bool) {})
				a.entries = append(a.entries, check)

			} else if placeHolder.isHead {
				options := slices.Collect(maps.Keys(mp3Config.NameToIDMod))
				//options = append(options, "File")
				slices.Sort(options)
				selEntry := widget.NewSelectEntry(options)
				selEntry.SetPlaceHolder(placeHolder.desc)
				a.entries = append(a.entries, selEntry)

			} else {
				textField := widget.NewEntry()
				textField.PlaceHolder = placeHolder.desc
				placeHolderSize := fyne.MeasureText(placeHolder.desc, 16, fyne.TextStyle{})
				textField.Resize(fyne.NewSize(placeHolderSize.Width, textField.MinSize().Height))
				a.entries = append(a.entries, textField)

			}

		}

	}

}

func (a *ActionEntries) rightContainerAdd(entry fyne.CanvasObject) {
	a.rightContainer.Add(entry)

}

func NewActionEntries(actions []*ActionEntry, updateEntry *widget.Entry) *ActionEntries {
	actionEntries := &ActionEntries{}
	_ = reflections.InitializeStruct(actionEntries)
	actionEntries.rightContainer = container.NewGridWithColumns(7)

	actionEntries.SetActions(actions, updateEntry)

	return actionEntries
}

/*func NewSelectEntryEntry(valueNum int, updateEntry *widget.Entry) []*widget.Entry {
	const maxEntries = 10
	var widgetEntries []*widget.Entry
	if valueNum >= 0 && valueNum <= maxEntries {
		for range valueNum {
			textField := widget.NewEntry()
			textField.PlaceHolder = "Field Name"
			textField.Resize(fyne.NewSize(100.0, textField.MinSize().Height))
			widgetEntries = append(widgetEntries, textField)

		}

	}

	return widgetEntries

}*/

// func SelectEntry(win fyne.Window, filePath string, state *TagTableDataAppState) *ActionSelectEntry {
func SelectEntryOld(win fyne.Window, filePath string) *ActionSelectEntry {
	fmt.Printf("SelectEntry()|start\n")
	//textEntry := widget.NewEntry()
	//state.textField = textEntry
	/*actionToFlag := map[string]string{
		"Tag":          "--tag",
		"Pre Replace":  "--pre-replace",
		"RegEx":        "--replace",
		"Album Folder": "--album-folder-name",
	}*/

	actions := []*ActionEntry{
		{
			name:         "Album Folder",
			action:       "--album-folder-name",
			placeHolders: []placeHolder{{desc: "folder name"}},
			isSingle:     true,
		}, {
			name:         "Dummy Files",
			action:       "--dummy-files",
			placeHolders: []placeHolder{{desc: "true/false"}},
			isBool:       true,
			isNotAlbum:   true,
		}, {
			name:         "Various Artists",
			action:       "--various",
			placeHolders: []placeHolder{{desc: "true/false"}},
			isBool:       true,
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
			name:   "Pre Condition",
			action: "--pre-condition",
			placeHolders: []placeHolder{
				{desc: "src tag", isHead: true},
				{desc: "regex"},
				{desc: "dest tag", isHead: true},
				{desc: "match value"},
				{desc: "else value"},
			},
		}, {
			name:   "Pre Replace",
			action: "--pre-replace",
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
			name:   "Tag",
			action: "--tag",
			placeHolders: []placeHolder{
				{desc: "tag", isHead: true},
				{desc: "value"},
				{desc: "optional value"},
			},
		},
	}

	//selectEntryEntries := NewActionEntries(textEntry)
	//selectEntry := NewActionSelectEntry([]string{"Album Folder", "Tag", "Pre Replace", "RegEx"}, "Choose or type")
	selectEntry := NewActionSelectEntry(win, filePath, actions, "Choose or type")
	/*submitButton := widget.NewButton("Submit", func() {
		var textEntryFields []*widget.Entry
		var flagText string
		switch selectEntry.Text {
		case "Album Folder":
			textEntryFields = selectEntryEntries.AlbumFolder
			flagText = fmt.Sprintf("%s \"%s", actionToFlag["Album Folder"], textEntryFields[0].Text)

		case "PreReplace":
			textEntryFields = selectEntryEntries.Replace
			flagText = fmt.Sprintf("%s \"%s=%s=%s", actionToFlag["PreReplace"], textEntryFields[0].Text, textEntryFields[1].Text)

		case "RegEx":
			textEntryFields = selectEntryEntries.Replace
			flagText = fmt.Sprintf("%s \"%s=%s=%s", actionToFlag["RegEx"], textEntryFields[0].Text, textEntryFields[1].Text)

		case "Tag":
			textEntryFields = selectEntryEntries.Tag
			flagText = fmt.Sprintf("%s \"%s=%s=%s", actionToFlag["Tag"], textEntryFields[0].Text, textEntryFields[1].Text)

		}

		textEntry.SetText(flagText)

	})
	submitButton.Importance = widget.LowImportance*/

	//rightContainer := container.NewHBox()
	//rightContainer := container.NewGridWithColumns(4)
	//rightContainer.Resize(fyne.NewSize(300, rightContainer.MinSize().Height))
	/*selectEntry.OnChanged = func(value string) {
		log.Println("clicked: ", value)
		if flag, ok := actionToFlag[value]; ok {
			textEntry.SetText(flag)
			rightContainer.RemoveAll()

			var textEntryFields []*widget.Entry
			switch value {
			case "Tag":
				textEntryFields = selectEntryEntries.Tag

			case "PreReplace":
				textEntryFields = selectEntryEntries.Replace

			case "RegEx":
				textEntryFields = selectEntryEntries.Replace

			case "Album Folder":
				textEntryFields = selectEntryEntries.AlbumFolder
			}

			for _, textEntry := range textEntryFields {
				fmt.Printf("Adding textEntry\n")
				rightContainer.Add(textEntry)

			}

			rightContainer.Add(submitButton)
			rightContainer.Refresh()

		}

	}*/

	//selectEntryContainer := container.NewHBox(selectEntry, rightContainer)
	//selectEntryContainer := container.NewBorder(nil, nil, selectEntry, nil, rightContainer)
	//return container.NewBorder(selectEntryContainer, textEntry, nil, nil, nil)
	//return container.NewBorder(nil, nil, selectEntry, nil)
	//return selectEntry.actionContainer
	return selectEntry

}
