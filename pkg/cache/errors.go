package cache

import (
	"net/http"
	"strconv"
)

type ErrUnexpectedStatusCode struct {
	URL    string
	Status int
}

func (e ErrUnexpectedStatusCode) Error() string {
	return "Unexpected status code calling '" + e.URL + "': " + strconv.Itoa(e.Status) + " " + http.StatusText(e.Status)
}
