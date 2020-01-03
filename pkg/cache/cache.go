package cache

import (
	"bytes"
	"io"
	"io/ioutil"
	"net/http"
	"sync"
)

type Cache struct {
	sync.Mutex
	cache map[string][]byte

	client http.Client
}

func (c *Cache) Get(url string) (io.Reader, error) {
	c.Lock()
	defer c.Unlock()

	// Initialise the map if this is the first call
	if c.cache == nil {
		c.cache = make(map[string][]byte)
	}

	// Return the cached content if it's available
	if r, ok := c.cache[url]; ok {
		return bytes.NewReader(r), nil
	}

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	res, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}

	if res.StatusCode != http.StatusOK {
		io.Copy(ioutil.Discard, res.Body)
		return nil, ErrUnexpectedStatusCode{url, res.StatusCode}
	}

	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	c.cache[url] = b

	return bytes.NewReader(b), nil
}
