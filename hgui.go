package main

import (
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	xWidget "fyne.io/x/fyne/widget"
	"time"
)

// hashlookupTab it tied to a URI under analysis
// it holds a hashlooker interface responsible for its behaviour
type hashlookupTab struct {
	hashlooker
	uri fyne.URI
}

// bloofilterTab holds the hashlookup filter and its tab-related data
type bloomfilterTab struct {
	tab      *container.TabItem
	isOpened bool
}

type hgui struct {
	win              fyne.Window
	app              *fyne.App
	projectRoot      fyne.URI
	resultsTabs      *container.DocTabs
	fileTree         *xWidget.FileTree
	openedHashlooker map[*container.TabItem]*hashlookupTab
	offlineMode      bool
	offlineTool      *widget.ToolbarAction
	toolBar          *widget.Toolbar
	// The Bloom filter is tied to the application
	Filter *HashlookupBloom
	bfTab  *bloomfilterTab
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
			if !h.offlineMode {
				hl = hashlookerByURI["file"](u, h)
				icon = theme.FileIcon()
			} else {
				dialog.ShowInformation("Offline Mode", "Cannot lookup details in offline mode.", h.win)
				return
			}
		} else if err != nil {
			dialog.ShowError(err, h.win)
			return
		}
	}

	newTab := container.NewTabItemWithIcon(u.Name(), icon, hl.content())
	h.openedHashlooker[newTab] = &hashlookupTab{hl, u}

	h.resultsTabs.Append(newTab)
	h.resultsTabs.Select(newTab)
}

// OpenBloomFilter download / load the filter and is
// a special tab that presents the filter's details
// as well as its download progress
func (h *hgui) OpenBloomFilter(operation string) {
	// TODO write a newBloomFilterTab function
	if h.Filter == nil {
		fmt.Println("filter is nil")
		h.Filter = &HashlookupBloom{}
	}
	// There was no tab struct so let's create one
	if h.bfTab == nil {
		fmt.Println("bfTab is nil")
		h.bfTab = &bloomfilterTab{nil, false}
	}

	// It opened, it already exist
	if h.bfTab.isOpened {
		fmt.Println("bfTab is isOpened")
		// need to cancel what was happening
		h.Filter.cancelDownload <- struct{}{}
		// reselect the tab if already opened
		for _, item := range h.resultsTabs.Items {
			if item == h.bfTab.tab {
				h.resultsTabs.Select(h.bfTab.tab)
				h.bfTab.isOpened = true
			}
		}
		// create fresh content
		h.bfTab.tab.Content = h.Filter.Content()
		// If not opened yet, we open it
	} else {
		h.bfTab.tab = container.NewTabItemWithIcon("Bloom filter", theme.InfoIcon(), h.Filter.Content())
		h.resultsTabs.Append(h.bfTab.tab)
		h.resultsTabs.Select(h.bfTab.tab)
		h.bfTab.isOpened = true
	}

	switch operation {
	case "download":
		// Closing the tab won't kill it
		go h.Filter.DownloadFilterToFile()
		// Let's launch a routing to monitor when it finishes
		go func() {
			for !h.Filter.Complete && !h.Filter.Cancelled {
				time.Sleep(time.Second * 1)
			}
			if h.Filter.Cancelled {
				// Reset the value in case of object reuse
				h.Filter.Cancelled = false
				return
			}
			if h.Filter.Complete {
				// Load the Filter and provide the filter details
				h.Filter.LoadFilterFromFile()
			}
		}()
	case "load":
		go h.Filter.LoadFilterFromFile()
	case "remote":
		go h.Filter.DownloadFilterToFilter()
	}
}
