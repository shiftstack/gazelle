package cache

import (
	"bytes"
	"io"
	"io/ioutil"
	"net/http"
	"sync"
)

type Cache struct {
	sync.RWMutex
	cache map[string][]byte

	client http.Client
}

func (c *Cache) Get(url string) (io.Reader, bool) {
	c.RLock()
	defer c.RUnlock()

	r, ok := c.cache[url]

	return bytes.NewReader(r), ok
}

func (c *Cache) Fetch(url string) (io.Reader, error) {
	c.Lock()
	defer c.Unlock()

	if c.cache == nil {
		c.cache = make(map[string][]byte)
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
