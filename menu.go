package main

import (
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

func (d *hgui) menuDownloadBloom() {
	fmt.Println("Menu action Download new bloom filter")
}

func (d *hgui) menuActionRunOffline() {
	fmt.Println("Menu action run offline hashlookup analysis")
}

func (d *hgui) menuActionRunOnline() {
	fmt.Println("Menu action run online hashlookup analysis")
}

func (d *hgui) menuActionSave() {
	fmt.Println("Menu action save")
}

func (d *hgui) makeMenu() *fyne.MainMenu {
	return fyne.NewMainMenu(
		fyne.NewMenu("File",
			fyne.NewMenuItemSeparator(),
			fyne.NewMenuItem("Download Filter", d.menuDownloadBloom),
			fyne.NewMenuItemSeparator(),
			fyne.NewMenuItem("Save", d.menuActionSave),
			fyne.NewMenuItem("Run Online", d.menuActionRunOnline),
			fyne.NewMenuItem("Run Offline", d.menuActionRunOffline),
		))
}

func (d *hgui) makeToolbar() *widget.Toolbar {
	return widget.NewToolbar(
		widget.NewToolbarAction(theme.FileIcon(), d.menuDownloadBloom),
		widget.NewToolbarAction(theme.DocumentSaveIcon(), d.menuActionSave),
		widget.NewToolbarAction(theme.MailForwardIcon(), d.menuActionRunOnline),
		widget.NewToolbarAction(theme.MailForwardIcon(), d.menuActionRunOffline),
	)
}
