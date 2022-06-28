package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/widget"
	xWidget "fyne.io/x/fyne/widget"
	"log"
)

func (h *hgui) makeFilesPanel() *xWidget.FileTree {
	files := xWidget.NewFileTree(h.projectRoot)
	files.Sorter = func(u1, u2 fyne.URI) bool {
		return u1.String() < u2.String() // Sort alphabetically
	}

	files.OnSelected = func(uid widget.TreeNodeID) {
		u, err := storage.ParseURI(uid)
		if err != nil {
			log.Fatal(err)
		}

		h.OpenHashlooker(u)
		return
	}
	return files
}

func (h *hgui) makeResultsPanel() fyne.CanvasObject {
	h.openedHashlooker = make(map[*container.TabItem]*hashlookupTab)
	welcome := widget.NewLabel("Welcome to Hashlookup-gui, the blahblah.\n\nChoose a starting folder in the list.")
	h.resultsTabs = container.NewDocTabs(
		container.NewTabItem("Welcome", welcome),
	)

	return container.NewMax(h.resultsTabs)
}
