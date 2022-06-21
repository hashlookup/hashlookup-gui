package main

import (
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"

	xWidget "fyne.io/x/fyne/widget"
)

type hgui struct {
	win         fyne.Window
	projectRoot fyne.URI
	resultsTabs *container.AppTabs
	fileTree    *xWidget.FileTree
}

func (d *hgui) doStuff(u fyne.URI) {
	fmt.Println("I do stuff.")
}
