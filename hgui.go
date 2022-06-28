package main

import (
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/theme"
	xWidget "fyne.io/x/fyne/widget"
	"hashlookup-gui/hashlookup"
	"log"
)

// hashlookupTab it tied to a URI under analysis
// it holds a hashlooker interface responsible for its behaviour
type hashlookupTab struct {
	hashlooker
	uri fyne.URI
}

type hgui struct {
	win              fyne.Window
	projectRoot      fyne.URI
	resultsTabs      *container.DocTabs
	fileTree         *xWidget.FileTree
	openedHashlooker map[*container.TabItem]*hashlookupTab
	// The Bloom filter is tied to the application
	filter *hashlookup.HashlookupBloom
}

func (h *hgui) OpenHashlooker(u fyne.URI) {
	// Tab already opened, set selected
	for tab, item := range h.openedHashlooker {
		if item.uri.String() == u.String() {
			h.resultsTabs.Select(tab)
			return
		}
	}

	// Instantiate a new hashlooker
	// Can be a file, or a folder.
	var hl hashlooker
	var icon fyne.Resource
	if hl == nil {
		if isDir, err := storage.CanList(u); err == nil && isDir {
			hl = hashlookerByURI["folder"](u, h)
			icon = theme.FolderOpenIcon()
		} else if err == nil && !isDir {
			hl = hashlookerByURI["file"](u, h)
			icon = theme.FileIcon()
		} else if err != nil {
			log.Fatal(err)
		}
	}

	newTab := container.NewTabItemWithIcon(u.Name(), icon, hl.content())
	h.openedHashlooker[newTab] = &hashlookupTab{hl, u}

	h.resultsTabs.Append(newTab)
	h.resultsTabs.Select(newTab)
}

func (d *hgui) doStuff(u fyne.URI) {
	fmt.Println("I do stuff.")
}
