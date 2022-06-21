package main

import (
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
)

// Declare conformity with editor interface
var _ hashlooker = (*folderHashlooker)(nil)

// Dummy data for the table
// TODO remove
var data = [][]string{[]string{"top left", "top right"},
	[]string{"bottom left", "bottom right"}}

type folderHashlooker struct {
	uri fyne.URI
	win fyne.Window
}

func newFolderHashlooker(u fyne.URI, win fyne.Window) hashlooker {
	return &folderHashlooker{uri: u, win: win}
}

func (g *folderHashlooker) content() fyne.CanvasObject {
	// Here we build the list files under scrutiny
	list := widget.NewTable(
		func() (int, int) {
			return len(data), len(data[0])
		},
		func() fyne.CanvasObject {
			return widget.NewLabel("wide content")
		},
		func(i widget.TableCellID, o fyne.CanvasObject) {
			o.(*widget.Label).SetText(data[i.Row][i.Col])
		})
	return list
}

func (g *folderHashlooker) close() {
	// Close the tab
	fmt.Println("Here I should be running the analysis.")
}

func (g *folderHashlooker) run() {
	// Run the hashlookup analysis
	fmt.Println("Here I should be running the analysis.")
}

func (g *folderHashlooker) export() {
	fmt.Println("Here I should be performing some export function.")
}
