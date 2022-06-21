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

type fileHashlooker struct {
	uri                   fyne.URI
	win                   fyne.Window
	client                *hashlookup.Client
	sha1                  string
	hashlookupBindingData binding.ExternalString
	hashlookupData        string
}

func newFileHashlooker(u fyne.URI, win fyne.Window) hashlooker {
	singleFile, err := ioutil.ReadFile(u.Path())
	if err != nil {
		log.Fatal(err)
	}
	h := sha1.New()
	h.Write(singleFile)
	digest := fmt.Sprintf("%x", h.Sum(nil))
	defaultTimeout := time.Second * 10
	client := hashlookup.NewClient("https://hashlookup.circl.lu", os.Getenv("HASHLOOKUP_API_KEY"), defaultTimeout)
	return &fileHashlooker{uri: u, win: win, sha1: digest, client: client}
}

func (g *fileHashlooker) content() fyne.CanvasObject {
	g.hashlookupData = "placeholder"
	g.hashlookupBindingData = binding.BindString(&g.hashlookupData)
	text := widget.NewLabelWithData(g.hashlookupBindingData)
	text.Wrapping = fyne.TextTruncate
	go func() {
		results, err := g.client.LookupSHA1(g.sha1)
		if err != nil {
			log.Fatal(err)
		}
		for key, child := range results.ChildrenMap() {
			//fmt.Printf("key: %v, value: %v\n", key, child.Data())
			g.hashlookupData += fmt.Sprintf("key: %v, value: %v\n", key, child.Data())
		}
		g.hashlookupBindingData.Reload()
	}()

	return text
}

//func (g *fileHashlooker) content() fyne.CanvasObject {
//	Here we detail each field we received from the hashlookup service
//	TODO - dummy label for the time being
//results, err := g.client.LookupSHA1(g.sha1)
//if err != nil {
//	log.Fatal(err)
//}
//return widget.NewLabel(fmt.Sprintf("%s", results))
//}

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
