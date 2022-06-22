package main

import (
	"crypto/sha1"
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/widget"
	"hashlookup-gui/hashlookup"
	"io/ioutil"
	"log"
	"os"
	"time"
)

// Declare conformity with editor interface
var _ hashlooker = (*fileHashlooker)(nil)

type line struct {
	Key   string
	Value string
}

type fileHashlooker struct {
	uri                fyne.URI
	hgui               *hgui
	client             *hashlookup.Client
	sha1               string
	hashlookupData     []line
	hashlookupBindings []binding.DataMap
}

func newFileHashlooker(u fyne.URI, hgui *hgui) hashlooker {
	singleFile, err := ioutil.ReadFile(u.Path())
	if err != nil {
		log.Fatal(err)
	}
	h := sha1.New()
	h.Write(singleFile)
	digest := fmt.Sprintf("%x", h.Sum(nil))
	defaultTimeout := time.Second * 10
	client := hashlookup.NewClient("https://hashlookup.circl.lu", os.Getenv("HASHLOOKUP_API_KEY"), defaultTimeout)
	// Init bindings
	hashlookupData := []line{}
	hashlookupBindings := []binding.DataMap{}
	return &fileHashlooker{uri: u, hgui: hgui, sha1: digest, client: client, hashlookupData: hashlookupData, hashlookupBindings: hashlookupBindings}
}

func (g *fileHashlooker) content() fyne.CanvasObject {
	t := widget.NewTable(
		func() (int, int) { return len(g.hashlookupData), 2 },
		func() fyne.CanvasObject {
			return widget.NewLabel("Cell 000, 000")
		},
		func(id widget.TableCellID, cell fyne.CanvasObject) {
			//label := cell.(*widget.Label)
			switch id.Col {
			case 0:
				tmphead, err := g.hashlookupBindings[id.Row].GetItem("Key")
				if err != nil {
					log.Fatal(err)
				}
				cell.(*widget.Label).Bind(tmphead.(binding.String))
			case 1:
				tmpvalue, err := g.hashlookupBindings[id.Row].GetItem("Value")
				if err != nil {
					log.Fatal(err)
				}
				cell.(*widget.Label).Bind(tmpvalue.(binding.String))
			}
		})
	t.SetColumnWidth(0, 200)
	t.SetColumnWidth(1, 800)

	go func() {
		var err error
		results, err := g.client.LookupSHA1(g.sha1)
		if err != nil {
			log.Fatal(err)
		}

		// update the string from hashlookup
		for key, value := range results.ChildrenMap() {
			tmpline := line{Key: fmt.Sprintf("%v", key), Value: fmt.Sprintf("%v", value)}
			g.hashlookupData = append(g.hashlookupData, tmpline)
			g.hashlookupBindings = append(g.hashlookupBindings, binding.BindStruct(&tmpline))
		}

		t.Refresh()
	}()

	return t
}

func (g *fileHashlooker) close() {
	// Close the tab
	fmt.Println("Here I should be closing the tab.")
}

func (g *fileHashlooker) run() {
	// Run the hashlookup analysis
	fmt.Println("Here I should be running the analysis.")
}

func (g *fileHashlooker) export() {
	fmt.Println("Here I should be performing some export function.")
}
