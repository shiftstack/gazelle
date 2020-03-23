package main

import (
	"bytes"
	"context"
	"encoding/json"
	"io/ioutil"
	"log"

	"github.com/elastic/go-elasticsearch/v7"
	"github.com/elastic/go-elasticsearch/v7/esapi"
)

type esDoc struct {
	jobRun
}

func (d esDoc) MarshalJSON() ([]byte, error) {
	var buildLog string
	{
		r, err := d.BuildLog()
		if err != nil {
			return nil, err
		}

		b, err := ioutil.ReadAll(r)
		if err != nil {
			return nil, err
		}

		buildLog = string(b)
	}

	var machines string
	{
		r, err := d.Machines()
		if err != nil {
			return nil, err
		}

		b, err := ioutil.ReadAll(r)
		if err != nil {
			return nil, err
		}

		machines = string(b)
	}

	var nodes string
	{
		r, err := d.Nodes()
		if err != nil {
			return nil, err
		}

		b, err := ioutil.ReadAll(r)
		if err != nil {
			return nil, err
		}

		nodes = string(b)
	}

	return json.Marshal(struct {
		BuildLog string `json:"build-log.txt"`
		Machines string `json:"machines.json"`
		Nodes    string `json:"openstack_nodes.log"`
	}{
		BuildLog: buildLog,
		Machines: machines,
		Nodes:    nodes,
	})
}

// ensure saves the job in ElasticSearch, or does nothing if it's there
// already.
func ensure(j jobRun) error {
	es, err := elasticsearch.NewDefaultClient()
	if err != nil {
		return err
	}

	doc := esDoc{j}

	jsonDoc, err := doc.MarshalJSON()
	if err != nil {
		return err
	}

	req := esapi.IndexRequest{
		Index:   "test",
		Body:    bytes.NewReader(jsonDoc),
		Refresh: "true",
	}

	// Perform the request with the client.
	res, err := req.Do(context.Background(), es)
	if err != nil {
		log.Fatalf("Error getting response: %s", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		log.Printf("[%s] Error indexing document ID=%d", res.Status(), i+1)
	} else {
		// Deserialize the response into a map.
		var r map[string]interface{}
		if err := json.NewDecoder(res.Body).Decode(&r); err != nil {
			log.Printf("Error parsing the response body: %s", err)
		} else {
			// Print the response status and indexed document version.
			log.Printf("[%s] %s; version=%d", res.Status(), r["result"], int(r["_version"].(float64)))
		}
	}

}
