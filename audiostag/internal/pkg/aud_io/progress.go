package aud_io

import (
	"context"
	"image/color"
	"log/slog"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

func NewInfiniteProgress(title string) *InfiniteProgress {
	return &InfiniteProgress{
		win:  fyne.CurrentApp().NewWindow(title),
		pBar: widget.NewProgressBarInfinite(),
	}

}

type InfiniteProgress struct {
	win            fyne.Window
	pBar           *widget.ProgressBarInfinite
	progressTitle  *canvas.Text
	progressUpdate *canvas.Text
	cancel         context.CancelFunc
	isStart        bool

	//progressUpdate *widget.Label
}

func (i *InfiniteProgress) SetCancel(cancel context.CancelFunc) {
	i.SetIsStart(false)
	i.cancel = cancel

}

func (i *InfiniteProgress) Start(message string) {
	//a := fyne.CurrentApp()
	//a := app.NewWithID("audios.audiostag.progress")
	//i.progressLabel = widget.NewLabel(message)
	i.win.SetCloseIntercept(func() {
		slog.Debug("SetCloseIntercept()|intercepted close")
		i.pBar.Stop()
		i.win.Close()
		fyne.CurrentApp().Quit()
		i.cancel()

	})

	go func() {
		fyne.DoAndWait(func() {
			stringSize := fyne.MeasureText(message, 14, fyne.TextStyle{})
			padding := theme.Padding() * 4
			targetWidth := max(200.0, stringSize.Width+float32(padding))
			targetHeight := stringSize.Height

			//i.win = a.NewWindow(title)
			//i.pBar = widget.NewProgressBarInfinite()
			i.progressTitle = canvas.NewText(message, color.White)
			i.progressTitle.TextSize = 12
			//i.progressTitle = widget.NewLabel(message)

			//i.progressUpdate = widget.NewLabel(message)
			i.progressUpdate = canvas.NewText("", color.White)
			i.progressUpdate.TextSize = 12
			progressContainer := container.New(
				layout.NewStackLayout(),
				i.pBar,
			)
			progressContainer.Resize(fyne.NewSize(targetWidth, 10))
			i.win.SetContent(container.NewVBox(
				i.progressTitle,
				i.progressUpdate,
				i.pBar,
			))

			i.win.Resize(fyne.NewSize(targetWidth, targetHeight))
			i.win.CenterOnScreen()
			i.win.Show()
			i.pBar.Start()

		})

		/*defer fyne.DoAndWait(func() {
			pBar.Stop()
			win.Close()

		})*/

		//time.Sleep(5 * time.Second)
		//work()

	}()

	i.SetIsStart(true)
	//i.A.Run()

}

func (i *InfiniteProgress) SetIsStart(start bool) {
	i.isStart = start

}

func (i *InfiniteProgress) Update(text string) {
	fyne.DoAndWait(func() {
		i.progressUpdate.Text = text
		i.progressUpdate.Refresh()
		//i.win.Show()
		//i.pBar.Start()

	})

}

func (i *InfiniteProgress) Stop() {
	go func() {
		defer fyne.DoAndWait(func() {
			i.pBar.Stop()
			i.win.Close()

		})

	}()

	i.SetIsStart(false)

}

/*func Start(title, message string, work func()) {
	a = fyne.CurrentApp()
	//a := app.NewWithID("audios.audiostag.progress")
	progressLabel := widget.NewLabel(message)

	go func() {
		fyne.DoAndWait(func() {
			win = a.NewWindow(title)
			pBar = widget.NewProgressBarInfinite()
			win.SetContent(container.NewVBox(
				//widget.NewLabel(message),
				progressLabel,
				pBar,
			))

			stringSize := fyne.MeasureText(title, 16, fyne.TextStyle{})
			padding := theme.Padding() * 4
			targetWidth := max(200.0, stringSize.Width+float32(padding))
			targetHeight := stringSize.Height

			win.Resize(fyne.NewSize(targetWidth, targetHeight))
			win.CenterOnScreen()
			win.Show()
			pBar.Start()

		})

		fyne.Do(func() {
			time.Sleep(2 * time.Second)
			progressLabel.Text = "toodles"
			progressLabel.Refresh()
			win.Show()
			pBar.Start()

		})

		//defer fyne.DoAndWait(func() {
		//	pBar.Stop()
		//	win.Close()

		//})

		//time.Sleep(5 * time.Second)
		//work()

	}()

	a.Run()

}

func Stop() {
	go func() {
		defer fyne.DoAndWait(func() {
			pBar.Stop()
			win.Close()

		})

	}()

}*/
