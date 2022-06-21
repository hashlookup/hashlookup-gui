package main

import (
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
)

// Declare conformity with editor interface
var _ hashlooker = (*fileHashlooker)(nil)

type fileHashlooker struct {
	uri fyne.URI
	win fyne.Window
}

func newFileHashlooker(u fyne.URI, win fyne.Window) hashlooker {
	return &fileHashlooker{uri: u, win: win}
}

func (g *fileHashlooker) content() fyne.CanvasObject {
	// Here we detail each field we received from the hashlookup service
	// TODO - dummy label for the time being
	return widget.NewLabel("text")
}

func (g *fileHashlooker) close() {
	// Close the tab
	fmt.Println("Here I should be running the analysis.")
}

func (g *fileHashlooker) run() {
	// Run the hashlookup analysis
	fmt.Println("Here I should be running the analysis.")
}

func (g *fileHashlooker) export() {
	fmt.Println("Here I should be performing some export function.")
}
