package hashlookup

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/Jeffail/gabs/v2"
)

type (
	Client struct {
		host       string
		httpClient *http.Client
		apiKey     string
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
	if err != nil {
		return resp, err
	}

	resp, err = gabs.ParseJSON([]byte(body))
	if err != nil {
		return resp, err
	}

	return resp, nil
}
