package main

import (
	"crypto/sha1"
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"hashlookup-gui/hashlookup"
	"io/ioutil"
	"log"
	"os"
	"time"
)

// Declare conformity with editor interface
var _ hashlooker = (*folderHashlooker)(nil)

type fileLookup struct {
	Uri *fyne.URI
	// ReqOnline true if the value has been queried online
	ReqOnline bool
	// ReqOnline true if the value has been queried against the bloom filter
	ReqOffline bool
	KnownStr   *string
	Known      binding.String
	Sha1       binding.String
	Sha1Str    *string
}

type folderHashlooker struct {
	uri        fyne.URI
	hgui       *hgui
	client     *hashlookup.Client
	fileList   []fileLookup
	folderList []fyne.URI
	grothons   TunnyJob
}

func workerFunc(myinterface interface{}) interface{} {
	var result string

	tmpuri := myinterface.(fyne.URI)
	fmt.Printf("Hashing %v\n", tmpuri.Name())
	singleFile, err := ioutil.ReadFile(tmpuri.Path())
	if err != nil {
		log.Fatal(err)
	}
	h := sha1.New()
	h.Write(singleFile)
	result = fmt.Sprintf("%x", h.Sum(nil))

	return result
}

func newFolderHashlooker(u fyne.URI, hgui *hgui) hashlooker {
	// List folder
	data, err := storage.List(u)
	if err != nil {
		log.Fatal(err)
	}
	fileList := []fileLookup{}
	folderList := []fyne.URI{}
	grothons := newTunnyJob(0, workerFunc)

	// Triaging files and folders
	//for i, uri := range data {
	for i, uri := range data {
		if isDir, err := storage.CanList(uri); err == nil && isDir {
			folderList = append(folderList, uri)
		} else if err == nil && !isDir {
			digest := ""
			tmpKnown := "Unknown"
			tmpFileLookup := fileLookup{
				Uri:        &uri,
				ReqOffline: false,
				ReqOnline:  false,
				KnownStr:   &tmpKnown,
				Known:      binding.BindString(&tmpKnown),
				Sha1Str:    &digest,
				Sha1:       binding.BindString(&digest),
			}

			fileList = append(fileList, tmpFileLookup)

			// TODO check cycling bug
			go func() {
				results := grothons.Pool.Process(uri).(string)
				fileList[i].Sha1.Set(results)
			}()

		} else if err != nil {
			log.Fatal(err)
		}
	}

	// Init hashlookup client
	defaultTimeout := time.Second * 10
	client := hashlookup.NewClient("https://hashlookup.circl.lu", os.Getenv("HASHLOOKUP_API_KEY"), defaultTimeout)

	return &folderHashlooker{uri: u, hgui: hgui, fileList: fileList, folderList: folderList, client: client, grothons: *grothons}
}

func (g *folderHashlooker) content() fyne.CanvasObject {
	var toDisplay []*widget.List
	var listFolders *widget.List
	var listFiles *widget.List

	if len(g.folderList) > 0 {
		listFolders = widget.NewList(
			func() int {
				return len(g.folderList)
			},
			func() fyne.CanvasObject {
				return container.NewHBox(widget.NewIcon(theme.FolderOpenIcon()), widget.NewLabel("Template Object"))
			},
			func(id widget.ListItemID, item fyne.CanvasObject) {
				item.(*fyne.Container).Objects[1].(*widget.Label).SetText(g.folderList[id].Name())
			},
		)
		listFolders.OnSelected = func(id widget.ListItemID) {
			g.hgui.OpenHashlooker(g.folderList[id])
		}
		toDisplay = append(toDisplay, listFolders)
	}

	if len(g.fileList) > 0 {
		listFiles = widget.NewList(
			func() int {
				return len(g.fileList)
			},
			func() fyne.CanvasObject {
				return container.NewHBox(widget.NewIcon(theme.FileIcon()), widget.NewLabel("Template Object"), widget.NewLabel("Template Object"), widget.NewLabel("Template Object"))
			},
			func(id widget.ListItemID, item fyne.CanvasObject) {
				item.(*fyne.Container).Objects[1].(*widget.Label).SetText((*g.fileList[id].Uri).Name())
				item.(*fyne.Container).Objects[2].(*widget.Label).Bind(g.fileList[id].Sha1)
				item.(*fyne.Container).Objects[3].(*widget.Label).Bind(g.fileList[id].Known)

				// Launch the lookup
				// TODO offline mode against the bloom filter
				if !g.fileList[id].ReqOffline && !g.fileList[id].ReqOnline {
					go func() {
						// TODO mutex
						g.fileList[id].ReqOnline = true
						var err error
						results, err := g.client.LookupSHA1(*g.fileList[id].Sha1Str)
						if err != nil {
							log.Fatal(err)
						}
						if results.S("message").String() == "\"Non existing SHA-1\"" {
							g.fileList[id].Known.Set("Unknown")
						} else {
							g.fileList[id].Known.Set("Known")
						}
						fmt.Printf("Request performed for %v\n", *g.fileList[id].Sha1Str)
					}()
				}
			},
		)

		listFiles.OnSelected = func(id widget.ListItemID) {
			g.hgui.OpenHashlooker(*(g.fileList[id].Uri))
		}
		toDisplay = append(toDisplay, listFiles)
	}

	switch len(toDisplay) {
	case 1:
		return toDisplay[0]
	case 2:
		return container.NewVBox(listFolders, listFiles)
	default:
		return widget.NewLabel("Empty folder")
	}
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
