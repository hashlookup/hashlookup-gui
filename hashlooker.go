package main

import "fyne.io/fyne/v2"

var hashlookerByURI = map[string]func(fyne.URI, *hgui) hashlooker{
	"folder": newFolderHashlooker,
	"file":   newFileHashlooker,
}

type hashlooker interface {
	close()
	content() fyne.CanvasObject
	run()
	export()
}
