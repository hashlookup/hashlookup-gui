package main

import (
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"hashlookup-gui/hashlookup"
	"log"
)

// Declare conformity with editor interface
var _ hashlooker = (*folderHashlooker)(nil)

type folderHashlooker struct {
	uri    fyne.URI
	hgui   *hgui
	client hashlookup.Client
}

func newFolderHashlooker(u fyne.URI, hgui *hgui) hashlooker {
	return &folderHashlooker{uri: u, hgui: hgui}
}

func (g *folderHashlooker) content() fyne.CanvasObject {
	data, err := storage.List(g.uri)
	if err != nil {
		log.Fatal(err)
	}

	list := widget.NewList(
		func() int {
			return len(data)
		},
		func() fyne.CanvasObject {
			return container.NewHBox(widget.NewIcon(theme.DocumentIcon()), widget.NewLabel("Template Object"))
		},
		func(id widget.ListItemID, item fyne.CanvasObject) {
			var icon fyne.Resource
			if isDir, err := storage.CanList(data[id]); err == nil && isDir {
				icon = theme.FolderOpenIcon()
			} else if err == nil && !isDir {
				icon = theme.FileIcon()
			} else if err != nil {
				log.Fatal(err)
			}
			item.(*fyne.Container).Objects[0].(*widget.Icon).Resource = icon
			item.(*fyne.Container).Objects[0].(*widget.Icon).Refresh()
			item.(*fyne.Container).Objects[1].(*widget.Label).SetText(data[id].Name())
		},
	)
	list.OnSelected = func(id widget.ListItemID) {
		g.hgui.OpenHashlooker(data[id])
	}
	list.OnUnselected = func(id widget.ListItemID) {
	}

	return list
}

func (g *folderHashlooker) close() {
	// Close the tab
	fmt.Println("Here I should be closing the tab.")
}

func (g *folderHashlooker) run() {
	// Run the hashlookup analysis
	fmt.Println("Here I should be running the analysis.")
}

func (g *folderHashlooker) export() {
	fmt.Println("Here I should be performing some export function.")
}
