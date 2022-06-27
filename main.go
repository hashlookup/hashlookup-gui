package main

import (
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/storage"
	"os"
	"path/filepath"
)

func (d *hgui) setProject(u fyne.URI) {
	d.projectRoot = u

	content := d.makeResultsPanel()
	d.fileTree = d.makeFilesPanel()
	mainSplit := container.NewHSplit(d.fileTree, content)
	mainSplit.Offset = 0.2

	d.win.SetMainMenu(d.makeMenu())
	d.win.SetContent(container.NewBorder(d.makeToolbar(), nil, nil, nil, mainSplit))
}

func main() {
	a := app.NewWithID("lu.circl.hashlookup-gui")
	//a.SetIcon(resourceIconPng)
	w := a.NewWindow("Hashlookup-gui")
	w.Resize(fyne.NewSize(1024, 768))

	ide := &hgui{win: w}
	if len(os.Args) > 1 {
		path, _ := filepath.Abs(os.Args[1])
		root := storage.NewFileURI(path)
		ide.setProject(root)

		//w.Show()
	} else {
		fmt.Println("else no args")
	}

	w.ShowAndRun()
}
