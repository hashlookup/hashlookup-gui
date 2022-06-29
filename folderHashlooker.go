package main

import (
	"bytes"
	"crypto/sha1"
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/Jeffail/gabs/v2"
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
	client     *Client
	fileList   []*fileLookup
	folderList []fyne.URI
	tunnyHash  TunnyJob
	tunnyReq   TunnyJob
}

type resultsWorkerType struct {
	err     error
	results interface{}
}

func hashingWorkerFunc(myinterface interface{}) interface{} {
	tmpuri := myinterface.(fyne.URI)
	f, err := os.Open(tmpuri.Path())
	if err != nil {
		return resultsWorkerType{err: err, results: fmt.Sprintf("File error on %v\n", tmpuri.Name())}
	}
	defer f.Close()

	h := sha1.New()
	if _, err := io.Copy(h, f); err != nil {
		log.Fatal(err)
		return resultsWorkerType{err: err, results: fmt.Sprintf("Copy error on %v\n", tmpuri.Name())}
	}

	return resultsWorkerType{err: nil, results: fmt.Sprintf("%x", h.Sum(nil))}
}

type requestingWorkerType struct {
	c           *Client
	b           *HashlookupBloom
	offlineMode bool
	fileLookup  *fileLookup
}

func requestingWorkerFunc(myinterface interface{}) interface{} {
	tmpWorkerType := myinterface.(requestingWorkerType)
	var err error
	if !tmpWorkerType.offlineMode {
		var results *gabs.Container
		results, err = tmpWorkerType.c.LookupSHA1(*tmpWorkerType.fileLookup.Sha1Str)
		if err != nil {
			return resultsWorkerType{err: err, results: fmt.Sprintf("Hashlookup service error on %v\n", *tmpWorkerType.fileLookup.Sha1Str)}
		}
		//fmt.Printf("Request performed for %v\n", *tmpWorkerType.fileLookup.Sha1Str)
		return resultsWorkerType{err: nil, results: results}
	} else {
		var results bool
		//fmt.Printf("Checking %v on filter\n", *tmpWorkerType.fileLookup.Sha1Str)
		results = tmpWorkerType.b.B.Check(bytes.ToUpper([]byte(*tmpWorkerType.fileLookup.Sha1Str)))
		return resultsWorkerType{err: nil, results: results}
	}
}

func newFolderHashlooker(u fyne.URI, hgui *hgui) hashlooker {
	// List folder
	data, err := storage.List(u)
	if err != nil {
		dialog.ShowError(err, hgui.win)
	}
	fileList := []*fileLookup{}
	folderList := []fyne.URI{}
	tunnyHash := newTunnyJob(0, hashingWorkerFunc)
	tunnyReq := newTunnyJob(0, requestingWorkerFunc)

	// Init hashlookup client
	defaultTimeout := time.Second * 10
	client := NewClient("https://hashlookup.circl.lu", os.Getenv("HASHLOOKUP_API_KEY"), defaultTimeout)

	// Triaging files and folders
	for _, uri := range data {
		// tmp variable to detach from the iteration variable
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

			go func() {
				results := tunnyHash.Pool.Process(tmpUri).(resultsWorkerType)
				tmpFileLookup.Sha1.Set(results.results.(string))
				if results.err != nil {
					return
				} else {
					if !hgui.offlineMode {
						reqResults := tunnyReq.Pool.Process(requestingWorkerType{c: client, b: hgui.Filter, offlineMode: hgui.offlineMode, fileLookup: &tmpFileLookup}).(resultsWorkerType)
						if reqResults.err != nil {
							tmpFileLookup.Known.Set(reqResults.results.(string))
						} else {
							if reqResults.results.(*gabs.Container).S("message").String() == "\"Non existing SHA-1\"" {
								tmpFileLookup.Known.Set("Unknown")
							} else {
								tmpFileLookup.Known.Set("Known")
							}
						}
					} else if hgui.offlineMode {
						reqResults := tunnyReq.Pool.Process(requestingWorkerType{c: client, b: hgui.Filter, offlineMode: hgui.offlineMode, fileLookup: &tmpFileLookup}).(resultsWorkerType)
						if reqResults.results.(bool) {
							tmpFileLookup.Known.Set("Known")
						} else {
							tmpFileLookup.Known.Set("Unknown")
						}
					}
				}
			}()

			fileList = append(fileList, &tmpFileLookup)

		} else if err != nil {
			dialog.ShowError(err, hgui.win)
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
		return widget.NewLabel("No Files to display in this folder.")
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
