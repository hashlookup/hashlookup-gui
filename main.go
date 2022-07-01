package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/storage"
	"os"
	"path/filepath"
)

func (h *hgui) setProject(u fyne.URI) {
	h.projectRoot = u

	content := h.makeResultsPanel()
	h.fileTree = h.makeFilesPanel()
	mainSplit := container.NewHSplit(h.fileTree, content)
	mainSplit.Offset = 0.2

	h.win.SetMainMenu(h.makeMenu())
	h.win.SetContent(container.NewBorder(h.makeToolbar(), nil, nil, nil, mainSplit))
}

func main() {
	a := app.NewWithID("lu.circl.hashlookup-gui")
	//a.SetIcon(resourceIconPng)
	w := a.NewWindow("Hashlookup-gui")
	w.Resize(fyne.NewSize(1024, 768))

	hgui := &hgui{win: w, offlineMode: false, app: &a}
	if len(os.Args) > 1 {
		path, _ := filepath.Abs(os.Args[1])
		root := storage.NewFileURI(path)
		hgui.setProject(root)
	} else {
		root := defaultDir()
		hgui.setProject(root)
	}

	w.ShowAndRun()
}
