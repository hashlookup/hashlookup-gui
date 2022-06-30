package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	xWidget "fyne.io/x/fyne/widget"
)

func (h *hgui) makeFilesPanel() *xWidget.FileTree {
	files := xWidget.NewFileTree(h.projectRoot)
	files.Sorter = func(u1, u2 fyne.URI) bool {
		return u1.String() < u2.String() // Sort alphabetically
	}

	files.OnSelected = func(uid widget.TreeNodeID) {
		u, err := storage.ParseURI(uid)
		if err != nil {
			dialog.ShowError(err, h.win)
			return
		}
		h.OpenHashlooker(u)
		files.UnselectAll()
		return
	}

	return files
}

func (h *hgui) makeResultsPanel() fyne.CanvasObject {
	h.openedHashlooker = make(map[*container.TabItem]*hashlookupTab)
	welcome := widget.NewLabel("Welcome to Hashlookup-gui, the GUI to query hashlookup service.\n\nChoose a starting folder in the list to investigate whether the files it contains are known from hashlookup.")
	h.resultsTabs = container.NewDocTabs(
		container.NewTabItemWithIcon("Welcome", theme.HomeIcon(),
			container.NewCenter(welcome)))

	h.resultsTabs.CloseIntercept = func(t *container.TabItem) {
		hl, ok := h.openedHashlooker[t]
		if !ok { // welcome tab or bloom filter tab won't close for now
			return
		} else {
			hl.close()
			h.resultsTabs.Remove(t)
			delete(h.openedHashlooker, t)
			return
		}
	}

	return container.NewMax(h.resultsTabs)
}
