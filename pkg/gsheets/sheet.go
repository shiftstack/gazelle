package gsheets

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/shiftstack/gazelle/pkg/job"

	"golang.org/x/net/context"
	"google.golang.org/api/sheets/v4"
)

const offset int = 6

type Sheet struct {
	JobName string
	Client  *Client

	idxList []int64
}

func (s *Sheet) getID() int64 {
	innerMap := map[string]int64{
		"release-openshift-ocp-installer-e2e-openstack-4.7":           1835312929,
		"release-openshift-ocp-installer-e2e-openstack-serial-4.7":    1216933186,
		"release-openshift-ocp-installer-e2e-openstack-4.6":           663598205,
		"release-openshift-ocp-installer-e2e-openstack-serial-4.6":    1620677487,
		"release-openshift-ocp-installer-e2e-openstack-4.5":           1993874237,
		"release-openshift-ocp-installer-e2e-openstack-serial-4.5":    605104092,
		"release-openshift-ocp-installer-e2e-openstack-4.4":           552238361,
		"release-openshift-ocp-installer-e2e-openstack-serial-4.4":    923764376,
		"release-openshift-ocp-installer-e2e-openstack-4.3":           1408408210,
		"release-openshift-ocp-installer-e2e-openstack-serial-4.3":    1992493886,
		"release-openshift-ocp-installer-e2e-openstack-4.2":           493400895,
		"release-openshift-ocp-installer-e2e-openstack-serial-4.2":    254644539,
		"release-openshift-origin-installer-e2e-openstack-4.2":        61282267,
		"release-openshift-origin-installer-e2e-openstack-serial-4.2": 1799792403,
	}

	sheetId, ok := innerMap[s.JobName]
	if !ok {
		log.Fatalf("Unknown job %v", s.JobName)
	}

	return sheetId
}

func (s *Sheet) getName() string {
	innerMap := map[string]string{
		"release-openshift-ocp-installer-e2e-openstack-4.7":           "Parallel OCP 4.7",
		"release-openshift-ocp-installer-e2e-openstack-serial-4.7":    "Serial OCP 4.7",
		"release-openshift-ocp-installer-e2e-openstack-4.6":           "Parallel OCP 4.6",
		"release-openshift-ocp-installer-e2e-openstack-serial-4.6":    "Serial OCP 4.6",
		"release-openshift-ocp-installer-e2e-openstack-4.5":           "Parallel OCP 4.5",
		"release-openshift-ocp-installer-e2e-openstack-serial-4.5":    "Serial OCP 4.5",
		"release-openshift-ocp-installer-e2e-openstack-4.4":           "Parallel OCP 4.4",
		"release-openshift-ocp-installer-e2e-openstack-serial-4.4":    "Serial OCP 4.4",
		"release-openshift-ocp-installer-e2e-openstack-4.3":           "Parallel OCP 4.3",
		"release-openshift-ocp-installer-e2e-openstack-serial-4.3":    "Serial OCP 4.3",
		"release-openshift-ocp-installer-e2e-openstack-4.2":           "Parallel OCP 4.2",
		"release-openshift-ocp-installer-e2e-openstack-serial-4.2":    "Serial OCP 4.2",
		"release-openshift-origin-installer-e2e-openstack-4.2":        "Parallel Origin 4.2",
		"release-openshift-origin-installer-e2e-openstack-serial-4.2": "Serial Origin 4.2",
	}

	sheetName, ok := innerMap[s.JobName]
	if !ok {
		log.Fatalf("Unknown job %v", s.JobName)
	}

	return sheetName
}

func (s *Sheet) getAllIDs() {
	readRange := s.getName() + "!A7:A300"
	resp, err := s.Client.service.Spreadsheets.Values.Get(s.Client.spreadsheetId, readRange).Do()
	if err != nil {
		log.Fatalf("Unable to retrieve data from sheet: %v", err)
	}

	if len(resp.Values) == 0 {
		fmt.Println("No data found.")
		return
	}

	for _, row := range resp.Values {
		id, err := strconv.ParseInt(fmt.Sprint(row[0]), 10, 64)

		if err != nil {
			id = 0
		}
		s.idxList = append(s.idxList, id)
	}
}

// Get the index on which the job with the given id should be inserted
func (s *Sheet) getIndex(id int64) (int, bool) {
	// Populate list of IDs from sheet
	if len(s.idxList) == 0 {
		s.getAllIDs()
	}

	i := 0
	for ; i < len(s.idxList); i++ {
		// s.idxList is sorted in descending order
		if s.idxList[i] <= id {
			break
		}
	}

	exists := len(s.idxList) > 0 && s.idxList[i] == id

	return i, exists
}

// Get ID of the most recent job in the spreadsheet
func (s *Sheet) GetLatestId() int64 {
	readRange := s.getName() + "!A7"
	resp, err := s.Client.service.Spreadsheets.Values.Get(s.Client.spreadsheetId, readRange).Do()
	if err != nil {
		log.Fatalf("Unable to retrieve data from sheet: %v", err)
	}

	if len(resp.Values) == 0 || len(resp.Values[0]) == 0 {
		return 0
	}

	data := fmt.Sprint(resp.Values[0][0])
	id, err := strconv.ParseInt(data, 10, 64)
	if err != nil {
		return 0
	}

	return id
}

func (s *Sheet) AddRow(j job.Job, user string) {
	sheetId := s.getID()

	id, _ := strconv.ParseInt(j.ID, 10, 64)
	idx, exists := s.getIndex(id)

	// Let's try render the job early so that we do not create a empty row
	// if the job does not exist
	rendered_job, err := jobToHtml(j, user)
	if err != nil {
		log.Printf("Could not fetch information about job %v: %v", j.ID, err)
		return
	}

	if !exists {
		// Create a new row to save the report
		idr := &sheets.InsertDimensionRequest{
			Range: &sheets.DimensionRange{
				Dimension:  "ROWS",
				SheetId:    sheetId,
				StartIndex: int64(idx + offset),
				EndIndex:   int64(idx + offset + 1),
			},
			InheritFromBefore: false,
		}

		request := &sheets.BatchUpdateSpreadsheetRequest{
			Requests: []*sheets.Request{{InsertDimension: idr}},
		}

		_, err := s.Client.service.Spreadsheets.BatchUpdate(s.Client.spreadsheetId, request).Context(context.Background()).Do()
		if err != nil {
			log.Fatal(err)
		}

		// Insert the new id into s.idxList
		s.idxList = append(s.idxList, 0)
		copy(s.idxList[idx+1:], s.idxList[idx:])
		s.idxList[idx] = id
	}

	pdr := &sheets.PasteDataRequest{
		Data: rendered_job,
		Coordinate: &sheets.GridCoordinate{
			SheetId:     sheetId,
			RowIndex:    int64(idx + offset),
			ColumnIndex: 0,
		},
		Type: "PASTE_NORMAL",
		Html: true,
	}

	request := &sheets.BatchUpdateSpreadsheetRequest{
		Requests: []*sheets.Request{{PasteData: pdr}},
	}

	_, err = s.Client.service.Spreadsheets.BatchUpdate(s.Client.spreadsheetId, request).Context(context.Background()).Do()
	if err != nil {
		log.Fatal(err)
	}
}

func jobToHtml(j job.Job, user string) (string, error) {
	startTime, err := j.StartTime()
	if err != nil {
		return "", err
	}
	var s strings.Builder
	{
		s.WriteString(`<meta http-equiv="content-type" content="text/html; charset=utf-8"><meta name="generator" content="cireport"/><table xmlns="http://www.w3.org/1999/xhtml"><tbody><tr><td>`)
		s.WriteString(strings.Join([]string{
			`<a href="` + j.JobURL() + `">` + j.ID + `</a>`,
			startTime.String(),
			j.Duration().String(),
			j.ComputedResult,
			`<a href="` + j.BuildLogURL() + `">` + j.BuildLogURL() + `</a>`,
			user,
			strings.Join(j.RootCause, "<br />"),
		}, "</td><td>"))
		s.WriteString(`</td></tr></tbody></table>`)
	}
	return s.String(), nil
}
