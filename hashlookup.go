package main

import (
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/DCSO/bloom"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/Jeffail/gabs/v2"
)

type (
	Client struct {
		host       string
		httpClient *http.Client
		apiKey     string
	}

	HashlookupBloom struct {
		hgui   *hgui
		tabPtr *container.TabItem
		B      *bloom.BloomFilter
		path   string
		// if the download completed
		Complete bool
		// if the download was cancelled
		Cancelled bool
		// if the filter is ready to use
		Ready bool
		// number of bytes read
		counter        *WriteCounter
		content        fyne.CanvasObject
		cancelDownload chan struct{}
	}
)

const (
	hashlookupURL = "lookup/sha1"
)

// WriteCounter counts the number of bytes written to it.
type WriteCounter struct {
	Total int64 // Total # of bytes transferred
}

func (wc *WriteCounter) Write(p []byte) (int, error) {
	n := len(p)
	wc.Total += int64(n)
	return n, nil
}

func NewClient(host string, apiKey string, timeout time.Duration) *Client {
	client := &http.Client{
		Timeout: timeout,
	}
	return &Client{
		host:       host,
		httpClient: client,
		apiKey:     apiKey,
	}
}

func (c *Client) do(method, endpoint string, param string) (*http.Response, error) {
	baseURL := fmt.Sprintf("%s/%s/%s", c.host, endpoint, param)
	req, err := http.NewRequest(method, baseURL, nil)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	req.Header.Add("Content-Type", "application/json")
	return c.httpClient.Do(req)
}

func (c *Client) LookupSHA1(sha1 string) (resp *gabs.Container, err error) {
	res, err := c.do(http.MethodGet, hashlookupURL, sha1)
	if err != nil {
		return
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	fmt.Printf("%s", body)
	if err != nil {
		log.Println(err)
		return resp, err
	}

	resp, err = gabs.ParseJSON([]byte(body))
	if err != nil {
		return resp, err
	}

	return resp, nil
}

// Create a new HashlookupBloom
// holding its tabs and the different string bindings
// the filter itself may not be ready after creation
func NewHashlookupBloom(path string, h *hgui) *HashlookupBloom {
	tmpBloom := bloom.BloomFilter{}
	counter := &WriteCounter{}
	tmpChan := make(chan struct{})
	return &HashlookupBloom{
		B:              &tmpBloom,
		path:           path,
		counter:        counter,
		Complete:       false,
		Ready:          false,
		hgui:           h,
		cancelDownload: tmpChan,
	}
}

// DownloadFilterToFilter downloads the bloom filter
// and load it without touching the disk
func (h *HashlookupBloom) DownloadFilterToFilter() error {
	var err error
	resp, err := http.Get(_bloomFilterDefaultUrl)
	defer resp.Body.Close()
	// wrap http.Get in the WriteCounter
	src := io.TeeReader(resp.Body, h.counter)
	if err != nil {
		return err
	}
	ch := make(chan struct{})
	go func() {
		select {
		case <-h.cancelDownload:
			fmt.Println("receive cancel")
			// dirty but it's the easiest - the alternative is to wrap the reader in a custom
			// reader that would understand contexts.
			resp.Body.Close()
			h.Cancelled = true
		case <-ch:
			return
		}
	}()
	h.B, err = bloom.LoadFromReader(src, _bloomFilterGzip)
	if err != nil {
		fmt.Println("Issue when reading from the remote %s: %s", _bloomFilterDefaultUrl, err)
	} else {
		h.Complete = true
	}
	ch <- struct{}{}
	h.GetFilterDetails()
	h.Ready = true
	h.StopBar()
	h.hgui.setOffline()
	return nil
}

// DownloadFilterToFile downloads the bloom filter into
// the file at path. Beware this is blocking and takes time
func (h *HashlookupBloom) DownloadFilterToFile() error {
	var err error
	out, err := os.Create(h.path)
	if err != nil {
		return err
	}
	defer out.Close()
	resp, err := http.Get(_bloomFilterDefaultUrl)
	ch := make(chan struct{})
	go func() {
		select {
		case <-h.cancelDownload:
			fmt.Println("receive cancel")
			// dirty but it's the easiest - the alternative is to wrap the reader in a custom
			// reader that would understand contexts.
			resp.Body.Close()
			h.Cancelled = true
		case <-ch:
			return
		}
	}()
	// wrap http.Get in the WriteCounter
	src := io.TeeReader(resp.Body, h.counter)
	_, err = io.Copy(out, src)
	if err != nil {
		//log.Fatalf("Issue when reading from the remote %s: %s", _bloomFilterDefaultUrl, err)
		fmt.Println("Issue when reading from the remote %s: %s", _bloomFilterDefaultUrl, err)
	} else {
		h.Complete = true
	}
	ch <- struct{}{}
	return nil
}

// LoadFilterFromFile loads the bloom filter from the file
// located at path. Beware this is blocking and takes time
func (h *HashlookupBloom) LoadFilterFromFile() error {
	var err error
	h.B, err = bloom.LoadFilter(h.path, _bloomFilterGzip)
	h.GetFilterDetails()
	if err != nil {
		log.Fatal(err)
	}
	h.Ready = true
	h.StopBar()
	h.hgui.setOffline()
	return nil
}

func (h *HashlookupBloom) Content() fyne.CanvasObject {
	content := widget.NewLabel("Bloom Filter Download:")
	tmpProgress := h.GetProgress()
	tmpDetails := ""
	progressBinding := binding.BindString(&tmpProgress)
	detailsBinging := binding.BindString(&tmpDetails)
	progress := widget.NewLabelWithData(progressBinding)
	details := widget.NewLabelWithData(detailsBinging)

	go func() {
		for !h.Complete {
			time.Sleep(time.Millisecond * 200)
			tmpProgress = h.GetProgress()
			progressBinding.Set(tmpProgress)
			progressBinding.Reload()
		}
		tmpProgress = fmt.Sprintf("Download complete with %s", tmpProgress)
		progressBinding.Set(tmpProgress)
		progressBinding.Reload()
	}()

	go func() {
		for !h.Ready {
			time.Sleep(time.Second * 1)
		}
		tmpDetails = h.GetFilterDetails()
		detailsBinging.Set(tmpDetails)
		detailsBinging.Reload()
	}()

	infinite := widget.NewProgressBarInfinite()
	cancelBtn := widget.NewButtonWithIcon("Cancel", theme.CancelIcon(), func() {
		h.cancelDownload <- struct{}{}
		h.hgui.resultsTabs.Remove(h.tabPtr)
		return
	})
	bottomHboxContainer := container.NewBorder(nil, nil, nil, cancelBtn, infinite)
	grid := container.New(layout.NewGridLayout(2), content, progress)
	// details will be stretched in the middle
	container := container.NewBorder(grid, bottomHboxContainer, nil, nil, details)
	h.content = container

	return container
}

func (h *HashlookupBloom) GetProgress() string {
	tmpStr := fmt.Sprintf("%d bytes.", h.counter.Total)
	return tmpStr
}

func (h *HashlookupBloom) GetFilterDetails() string {
	tmpStr := ""
	tmpStr += fmt.Sprintf("File:\t\t\t%s\n", h.path)
	tmpStr += fmt.Sprintf("Capacity:\t\t%d\n", h.B.MaxNumElements())
	tmpStr += fmt.Sprintf("Elements present:\t%d\n", h.B.N)
	tmpStr += fmt.Sprintf("FP probability:\t\t%.2e\n", h.B.FalsePositiveProb())
	tmpStr += fmt.Sprintf("Bits:\t\t\t%d\n", h.B.NumBits())
	tmpStr += fmt.Sprintf("Hash functions:\t\t%d\n", h.B.NumHashFuncs())
	return tmpStr
}

func (h *HashlookupBloom) StopBar() {
	h.content.(*fyne.Container).Objects[2].(*fyne.Container).Hide()
}
