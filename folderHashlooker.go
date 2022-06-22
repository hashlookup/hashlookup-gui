package main

import (
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
	"hashlookup-gui/hashlookup"
)

// Declare conformity with editor interface
var _ hashlooker = (*folderHashlooker)(nil)

type folderHashlooker struct {
	uri    fyne.URI
	win    fyne.Window
	client hashlookup.Client
}

func newFolderHashlooker(u fyne.URI, win fyne.Window) hashlooker {
	return &folderHashlooker{uri: u, win: win}
}

func (g *folderHashlooker) content() fyne.CanvasObject {
	content := widget.NewLabel("placeholder")
	return content
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
