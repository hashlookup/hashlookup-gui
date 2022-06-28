package hashlookup

import (
	"fmt"
	"github.com/DCSO/bloom"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/Jeffail/gabs/v2"
)

const (
	_bloomFilterDefaultUrl = "https://cra.circl.lu/hashlookup/hashlookup-full.bloom"
	_bloomFilterGzip       = false
)

type (
	Client struct {
		host       string
		httpClient *http.Client
		apiKey     string
	}

	HashlookupBloom struct {
		b    bloom.BloomFilter
		path string
	}
)

const (
	hashlookupURL = "lookup/sha1"
)

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

// Make this non blocking
func NewFilterFromFile(path string, name string) (*HashlookupBloom, error) {
	var err error
	path = filepath.Join(path, name)
	fmt.Println(path)
	fmt.Println(name)

	out, err := os.Create(path)
	if err != nil {
		log.Fatal(err)
	}
	defer out.Close()
	resp, err := http.Get(_bloomFilterDefaultUrl)
	defer resp.Body.Close()
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	tmpBloom, err := bloom.LoadFilter(path, _bloomFilterGzip)
	if err != nil {
		log.Fatal(err)
	}

	return &HashlookupBloom{
		b:    *tmpBloom,
		path: path,
	}, err
}
