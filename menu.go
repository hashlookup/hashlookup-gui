package main

import (
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"os"
	"path/filepath"

	"hashlookup-gui/hashlookup"
)

func (h *hgui) showSaveBloomDialog() {
	parent := widget.NewButton("Choose directory", nil)
	dir := defaultDir()
	if dir != nil {
		parent.SetText(dir.Name())
	}
	parent.OnTapped = func() {
		dialog.ShowFolderOpen(func(u fyne.ListableURI, err error) {
			if err != nil {
				dialog.ShowError(err, h.win)
				return
			}
			if u == nil {
				return
			}

			dir = u
			parent.SetText(u.Name())
		}, h.win)
	}

	name := widget.NewEntry()
	dialog.ShowForm("Save Filter", "Download", "Cancel", []*widget.FormItem{
		widget.NewFormItem("Parent directory", parent),
		widget.NewFormItem("File name", name),
	}, func(ok bool) {
		if !ok {
			return
		}
		var err error
		h.filter = hashlookup.NewHashlookupBloom(filepath.Join(dir.Path(), name.Text))
		if err != nil {
			dialog.ShowError(err, h.win)
		} else {
			h.OpenBloomFilter()
		}
	}, h.win)
}

func (h *hgui) menuActionRunOffline() {
	fmt.Println("Menu action run offline hashlookup analysis")
}

func (h *hgui) menuActionRunOnline() {
	fmt.Println("Menu action run online hashlookup analysis")
}

func (h *hgui) menuActionSave() {
	fmt.Println("Menu action save")
}

func (h *hgui) makeMenu() *fyne.MainMenu {
	return fyne.NewMainMenu(
		fyne.NewMenu("File",
			fyne.NewMenuItemSeparator(),
			fyne.NewMenuItem("Download Filter", h.showSaveBloomDialog),
			fyne.NewMenuItemSeparator(),
			fyne.NewMenuItem("Save", h.menuActionSave),
			fyne.NewMenuItem("Run Online", h.menuActionRunOnline),
			fyne.NewMenuItem("Run Offline", h.menuActionRunOffline),
		))
}

func (h *hgui) makeToolbar() *widget.Toolbar {
	return widget.NewToolbar(
		widget.NewToolbarAction(theme.FileIcon(), h.showSaveBloomDialog),
		widget.NewToolbarAction(theme.DocumentSaveIcon(), h.menuActionSave),
		widget.NewToolbarAction(theme.MailForwardIcon(), h.menuActionRunOnline),
		widget.NewToolbarAction(theme.MailForwardIcon(), h.menuActionRunOffline),
	)
}

func defaultDir() fyne.ListableURI {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		fyne.LogError("Failed to get user home directory", err)
		return nil
	}
	defaultDir := storage.NewFileURI(homeDir)
	newDir, err := storage.ListerForURI(defaultDir)
	if err != nil {
		fyne.LogError("Failed to list home directory", err)
		return nil
	}
	return newDir
}
