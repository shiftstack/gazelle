package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/elastic/go-elasticsearch/v7"
)

type jobRun interface {
	Result() (string, error)
	BuildLog() (io.Reader, error)
	Machines() (io.Reader, error)
	Nodes() (io.Reader, error)
	JUnit() (io.Reader, error)
}

type Rule struct {
	ID          string   `json:"id"`
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Links       []string `json:"links"`
	Query       string   `json:"query"`
}

func (r Rule) Match(j jobRun) (bool, error) {
	es, err := elasticsearch.NewDefaultClient()
	if err != nil {
		return false, err
	}

	res, err := es.Search(
		es.Search.WithContext(context.TODO()),
		es.Search.WithIndex("run"),
		es.Search.WithBody(strings.NewReader(r.Query)),
	)
	if err != nil {
		return false, err
	}
	defer res.Body.Close()

	if res.IsError() {
		var e map[string]interface{}
		if err := json.NewDecoder(res.Body).Decode(&e); err != nil {
			return false, fmt.Errorf("Error parsing the response body: %s", err)
		} else {
			// Print the response status and error information.
			return false, fmt.Errorf("[%s] %s: %s",
				res.Status(),
				e["error"].(map[string]interface{})["type"],
				e["error"].(map[string]interface{})["reason"],
			)
		}
	}

	var v map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&v); err != nil {
		return false, fmt.Errorf("Error parsing the response body: %s", err)
	}

	fmt.Println(v)

	panic("not implemented")
}
