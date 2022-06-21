package main

import (
	"crypto/sha1"
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/widget"
	"hashlookup-gui/hashlookup"
	"io/ioutil"
	"log"
	"os"
	"time"

	xWidget "fyne.io/x/fyne/widget"
)

func (d *hgui) makeFilesPanel() *xWidget.FileTree {
	//d.openEditors = make(map[*container.TabItem]*fileTab)

	files := xWidget.NewFileTree(d.projectRoot)
	//files.Filter = filterHidden()
	files.Sorter = func(u1, u2 fyne.URI) bool {
		return u1.String() < u2.String() // Sort alphabetically
	}

	// See how the defyne's openEditor works
	// On selected:
	// create a newTabItem (app)
	// along with a fyne list data binded to ta hfile list
	// an hashlookup goroutine will be fired up to populate the hfile list

	files.OnSelected = func(uid widget.TreeNodeID) {
		u, err := storage.ParseURI(uid)
		isDir, _ := storage.CanList(u)
		if isDir {
			//Build recursive list of files
			fileList := HashFolder(u.Path())
			for k, _ := range fileList {
				// Call hashlookup API
				defaultTimeout := time.Second * 10
				client := hashlookup.NewClient("https://hashlookup.circl.lu", os.Getenv("HASHLOOKUP_API_KEY"), defaultTimeout)
				fileList[k].Blob, err = client.LookupSHA1(k)
				if err != nil {
					log.Fatal(err)
				}
			}
			fmt.Println(fileList)
			return
		} else {
			u, err := storage.ParseURI(uid)
			singleFile, err := ioutil.ReadFile(u.Path())
			h := sha1.New()
			h.Write(singleFile)
			digest := fmt.Sprintf("%x", h.Sum(nil))
			defaultTimeout := time.Second * 10
			client := hashlookup.NewClient("https://hashlookup.circl.lu", os.Getenv("HASHLOOKUP_API_KEY"), defaultTimeout)
			results, err := client.LookupSHA1(digest)
			if err != nil {
				log.Fatal(err)
			}

			fmt.Println(results)
		}

		if err != nil {
			dialog.ShowError(err, d.win)
		}

		//d.openEditor(u)
	}
	return files
}

func (d *hgui) makeResultsPanel() fyne.CanvasObject {
	welcome := widget.NewLabel("Welcome to Hashlookup-gui, the blahblah.\n\nChoose a starting folder in the list.")
	second := widget.NewLabel("Second tab.\n\n. Pouf Pouf.")
	third := widget.NewLabel("Third tab.\n\n. Pouf Pouf.")
	d.resultsTabs = container.NewAppTabs(
		container.NewTabItem("Welcome", welcome),
		container.NewTabItem("test de second tab", second),
	)

	d.resultsTabs.Append(container.NewTabItem("test de troisieme tab", third))

	return container.NewMax(d.resultsTabs)
}
