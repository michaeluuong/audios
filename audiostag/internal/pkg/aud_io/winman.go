package aud_io

import (
	"context"
	"fmt"
	"log/slog"
	"sync"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
)

var (
	winmanInstance *Winman
	once           sync.Once
)

func GetWinmanInstance() *Winman {
	once.Do(func() {
		slog.Debug("NewWinman()|initialize")

		winmanInstance = &Winman{
			winApp: app.NewWithID("audios.audiostag.winman"),
		}

	})

	return winmanInstance

}

type Winman struct {
	winApp             fyne.App
	infiniteProgress   *InfiniteProgress
	isInfiniteProgress bool
	isAppRunning       bool
}

func (w *Winman) Quit(isForceOpt ...bool) {
	fmt.Printf("Winman.Quit()\n")
	var isForce bool
	if len(isForceOpt) > 0 {
		isForce = isForceOpt[0]

	}

	if w.IsAppRunning() || isForce {
		if w.IsInfiniteProgress() {
			w.InfiniteProgressStop()
			w.SetIsInfiniteProgress(false)

		}

		fyne.DoAndWait(func() {
			w.winApp.Quit()

		})

	}

}

func (w *Winman) NewWindow() fyne.Window {
	win := w.winApp.NewWindow("audiostag")

	return win

}

func (w *Winman) IsAppRunning() bool {
	return w.isAppRunning

}

func (w *Winman) SetIsAppRunning(isAppRunning bool) {
	w.isAppRunning = isAppRunning

}

func (w *Winman) Run(runfuncOpt ...func()) {
	if len(runfuncOpt) > 0 && runfuncOpt[0] != nil {
		go runfuncOpt[0]()

	}

	if !w.IsAppRunning() {
		w.SetIsAppRunning(true)
		w.winApp.Run()

	}

}

func (w *Winman) SetCancel(cancel context.CancelFunc) {
	w.infiniteProgress.SetCancel(cancel)

}

func (w *Winman) NewInfiniteProgress(title string) {
	w.infiniteProgress = NewInfiniteProgress(title)
	w.SetIsInfiniteProgress(true)

}

func (w *Winman) InfiniteProgressStart(cancel context.CancelFunc, title string, messageOpt ...string) {
	var message string
	if len(messageOpt) > 0 && messageOpt[0] != "" {
		message = messageOpt[0]

	}

	if w.infiniteProgress == nil {
		w.NewInfiniteProgress(title)
		w.isInfiniteProgress = true
		w.infiniteProgress.Start(message)

		if cancel != nil {
			w.SetCancel(cancel)

		}

	}

}

func (w *Winman) InfiniteProgressStop() {
	if w.IsInfiniteProgress() {
		w.infiniteProgress.Stop()

	}

}

func (w *Winman) InfiniteProgressUpdate(text string) {
	w.infiniteProgress.Update(text)

}

func (w *Winman) SetIsInfiniteProgress(isInfiniteProgress bool) {
	w.isInfiniteProgress = isInfiniteProgress

}

func (w *Winman) IsInfiniteProgress() bool {
	return w.isInfiniteProgress

}

func (w *Winman) TagTableShow(title string, data [][]string, cImg CoverImage, showActions bool, hasCompleted bool) {
	win := w.NewWindow()

	ShowTagTable(title, data, cImg, win, showActions, hasCompleted)

}
