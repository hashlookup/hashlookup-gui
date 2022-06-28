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
	"github.com/Jeffail/gabs/v2"
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
	tunnyHash  TunnyJob
	tunnyReq   TunnyJob
}

func hashingWorkerFunc(myinterface interface{}) interface{} {
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
	return result
}

type requestingWorkerType struct {
	c          *hashlookup.Client
	fileLookup *fileLookup
}

func requestingWorkerFunc(myinterface interface{}) interface{} {
	tmpWorkerType := myinterface.(requestingWorkerType)
	var err error
	//results, err := tmpFileLookup.c.LookupSHA1()
	results, err := tmpWorkerType.c.LookupSHA1(*tmpWorkerType.fileLookup.Sha1Str)
	if err != nil {
		log.Println(results)
		log.Println(err)
	}

	fmt.Printf("Request performed for %v\n", *tmpWorkerType.fileLookup.Sha1Str)
	return results
}

func newFolderHashlooker(u fyne.URI, hgui *hgui) hashlooker {
	// List folder
	data, err := storage.List(u)
	if err != nil {
		log.Fatal(err)
	}
	fileList := []*fileLookup{}
	folderList := []fyne.URI{}
	tunnyHash := newTunnyJob(0, hashingWorkerFunc)
	tunnyReq := newTunnyJob(0, requestingWorkerFunc)

	// Init hashlookup client
	defaultTimeout := time.Second * 10
	client := hashlookup.NewClient("https://hashlookup.circl.lu", os.Getenv("HASHLOOKUP_API_KEY"), defaultTimeout)

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
				results := tunnyHash.Pool.Process(tmpUri).(string)
				tmpFileLookup.Sha1.Set(results)

				reqResults := tunnyReq.Pool.Process(requestingWorkerType{c: client, fileLookup: &tmpFileLookup}).(*gabs.Container)

				if reqResults.S("message").String() == "\"Non existing SHA-1\"" {
					tmpFileLookup.Known.Set("Unknown")
				} else {
					tmpFileLookup.Known.Set("Know")
				}
			}()

			fileList = append(fileList, &tmpFileLookup)

		} else if err != nil {
			log.Fatal(err)
		}
	}

	return &folderHashlooker{uri: u, hgui: hgui, fileList: fileList, folderList: folderList, client: client, tunnyHash: *tunnyHash, tunnyReq: *tunnyReq}
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
