package prow

import (
	"log"
	"net/http"
	"strconv"

	"github.com/PuerkitoBio/goquery"
)

// Get ID of the most recent job from prow
func GetLatestId(jobName string) int {
	// Request the HTML page.
	res, err := http.Get("https://prow.svc.ci.openshift.org/job-history/origin-ci-test/logs/" + jobName)
	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		log.Fatalf("status code error: %d %s", res.StatusCode, res.Status)
	}

	// Load the HTML document
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		log.Fatal(err)
	}

	jobId := ""

	doc.Find(".mdl-data-table__cell--non-numeric a").EachWithBreak(func(i int, s *goquery.Selection) bool {
		if !s.Parent().Parent().HasClass("run-pending") {
			jobId = s.Text()
			return false
		}
		return true
	})

	id, err := strconv.Atoi(jobId)
	if err != nil {
		log.Fatal(err)
	}
	return id
}
