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
	"io"
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
	fileList   []*fileLookup
	folderList []fyne.URI
	grothons   TunnyJob
}

func workerFunc(myinterface interface{}) interface{} {
	var result string

	tmpuri := myinterface.(fyne.URI)
	fmt.Printf("Hashing %v\n", tmpuri.Name())
	f, err := os.Open(tmpuri.Path())
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	h := sha1.New()
	if _, err := io.Copy(h, f); err != nil {
		log.Fatal(err)
	}

	result = fmt.Sprintf("%x", h.Sum(nil))

	//singleFile, err := ioutil.ReadFile(tmpuri.Path())
	//if err != nil {
	//	log.Fatal(err)
	//}
	//h := sha1.New()
	//h.Write(singleFile)
	//result = fmt.Sprintf("%x", h.Sum(nil))

	return result
}

func newFolderHashlooker(u fyne.URI, hgui *hgui) hashlooker {
	// List folder
	data, err := storage.List(u)
	if err != nil {
		log.Fatal(err)
	}
	fileList := []*fileLookup{}
	folderList := []fyne.URI{}
	grothons := newTunnyJob(0, workerFunc)

	// Triaging files and folders
	for _, uri := range data {
		// tmp variable to detach from the iteartion variable
		tmpUri := uri
		if isDir, err := storage.CanList(tmpUri); err == nil && isDir {
			folderList = append(folderList, tmpUri)
		} else if err == nil && !isDir {
			digest := ""
			tmpKnown := "Unknown"
			tmpFileLookup := fileLookup{
				Uri:        &tmpUri,
				ReqOffline: false,
				ReqOnline:  false,
				KnownStr:   &tmpKnown,
				Known:      binding.BindString(&tmpKnown),
				Sha1Str:    &digest,
				Sha1:       binding.BindString(&digest),
			}

			// TODO check cycling bug
			go func() {
				results := grothons.Pool.Process(tmpUri).(string)
				tmpFileLookup.Sha1.Set(results)
			}()

			fileList = append(fileList, &tmpFileLookup)

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

	// TODO make this a preference, dead code for now
	var enableFolder = false
	if len(g.folderList) > 0 && enableFolder {
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
							log.Println(results)
							log.Println(err)
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
		// TODO fix this ugly thing when displaying folders and files
		return container.NewGridWithColumns(1, toDisplay[0], toDisplay[1])
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
