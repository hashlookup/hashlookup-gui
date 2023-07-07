package main

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

func (h *hgui) setOffline() {
	if h.offlineMode {
		return
	} else {
		if h.Filter.Ready {
			h.offlineMode = true
			h.offlineTool.SetIcon(theme.VisibilityOffIcon())
			h.toolBar.Refresh()
			dialog.ShowInformation("Offline Mode", "Offline mode enabled", h.win)
		} else {
			dialog.ShowInformation("Offline Mode", "No Filter loaded", h.win)
		}
	}
}

func (h *hgui) switchOffline() {
	if h.offlineMode {
		h.offlineMode = false
		h.offlineTool.SetIcon(theme.VisibilityIcon())
		h.toolBar.Refresh()
		dialog.ShowInformation("Offline Mode", "Offline mode disabled", h.win)
	} else if !h.offlineMode {
		if h.Filter.Ready {
			h.offlineMode = true
			h.offlineTool.SetIcon(theme.VisibilityOffIcon())
			h.toolBar.Refresh()
			dialog.ShowInformation("Offline Mode", "Offline mode enabled", h.win)
		} else {
			dialog.ShowInformation("Offline Mode", "No Filter loaded", h.win)
		}
	}
}

func (h *hgui) loadFilterFromRemote() {
	h.Filter = NewHashlookupBloom("it's in your RAM dude.", h)
	h.OpenBloomFilter("remote")
}

func (h *hgui) loadFilterFromFile() {
	dialog.ShowFileOpen(func(u fyne.URIReadCloser, err error) {
		if err != nil {
			dialog.ShowError(err, h.win)
			return
		}
		if u == nil {
			return
		}
		h.Filter = NewHashlookupBloom(u.URI().Path(), h)
		if err != nil {
			dialog.ShowError(err, h.win)
		} else {
			h.OpenBloomFilter("load")
		}
	}, h.win)
}

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
		h.Filter = NewHashlookupBloom(filepath.Join(dir.Path(), name.Text), h)
		if err != nil {
			dialog.ShowError(err, h.win)
		} else {
			h.OpenBloomFilter("download")
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

func (h *hgui) changeProjectDir() {
	dialog.ShowFolderOpen(func(u fyne.ListableURI, err error) {
		if err != nil {
			dialog.ShowError(err, h.win)
			return
		}
		if u == nil {
			return
		}
		h.setProject(u)
		if err != nil {
			dialog.ShowError(err, h.win)
		} else {
			return
		}
	}, h.win)
}

func (h *hgui) makeMenu() *fyne.MainMenu {
	file := fyne.NewMenu("File",
		fyne.NewMenuItemSeparator(),
		fyne.NewMenuItem("Download Filter", h.showSaveBloomDialog),
		fyne.NewMenuItem("Load Filter From File", h.loadFilterFromFile),
		fyne.NewMenuItem("Load Filter From Remote", h.loadFilterFromRemote),
		fyne.NewMenuItemSeparator(),
		fyne.NewMenuItem("Switch Offline mode", h.switchOffline),
		fyne.NewMenuItemSeparator(),
		fyne.NewMenuItem("Reset Tree Root", h.changeProjectDir),
		//fyne.NewMenuItem("Save", h.menuActionSave),
		//fyne.NewMenuItem("Run Online", h.menuActionRunOnline),
		//fyne.NewMenuItem("Run Offline", h.menuActionRunOffline),
	)

	helpMenu := fyne.NewMenu("Help",
		fyne.NewMenuItem("Github", func() {
			u, _ := url.Parse("https://github.com/hashlookup/hashlookup-gui")
			_ = (*h.app).OpenURL(u)
		}))
	return fyne.NewMainMenu(file, helpMenu)
}

func (h *hgui) makeToolbar() *widget.Toolbar {
	if h.offlineMode {
		h.offlineTool = widget.NewToolbarAction(theme.VisibilityOffIcon(), h.switchOffline)
	} else {
		h.offlineTool = widget.NewToolbarAction(theme.VisibilityIcon(), h.switchOffline)
	}
	h.toolBar = widget.NewToolbar(
		//widget.NewToolbarAction(theme.FileIcon(), h.showSaveBloomDialog),
		//widget.NewToolbarAction(theme.DocumentSaveIcon(), h.menuActionSave),
		//widget.NewToolbarAction(theme.MailForwardIcon(), h.menuActionRunOnline),
		//widget.NewToolbarAction(theme.MailForwardIcon(), h.menuActionRunOffline),
		h.offlineTool,
	)
	return h.toolBar
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
