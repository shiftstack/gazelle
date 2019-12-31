package cache

import (
	"net/http"
	"strconv"
)

type ErrUnexpectedStatusCode int

func (e ErrUnexpectedStatusCode) Error() string {
	return "Unexpected status code: " + strconv.Itoa(int(e)) + " " + http.StatusText(int(e))
}
