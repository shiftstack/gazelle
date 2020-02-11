package gsheets

import (
	"fmt"
	"log"
	"strconv"

	"github.com/shiftstack/gazelle/pkg/job"

	"golang.org/x/net/context"
	"google.golang.org/api/sheets/v4"
)

type Sheet struct {
	JobName string
	Client  *Client

	idxList []string
}

func (s *Sheet) getID() int64 {
	innerMap := map[string]int64{
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

// Get ID of the most recent job in the spreadsheet
func (s *Sheet) GetLatestId() int {
	readRange := s.getName() + "!A7"
	resp, err := s.Client.service.Spreadsheets.Values.Get(s.Client.spreadsheetId, readRange).Do()
	if err != nil {
		log.Fatalf("Unable to retrieve data from sheet: %v", err)
	}

	if len(resp.Values) == 0 || len(resp.Values[0]) == 0 {
		return 0
	}

	data := fmt.Sprint(resp.Values[0][0])
	id, err := strconv.Atoi(data)
	if err != nil {
		return 0
	}

	return id
}

func (s *Sheet) AddRow(j job.Job) {
	sheetId := s.getID()

	idr := &sheets.InsertDimensionRequest{
		Range: &sheets.DimensionRange{
			Dimension:  "ROWS",
			SheetId:    sheetId,
			StartIndex: 6,
			EndIndex:   7,
		},
		InheritFromBefore: false,
	}

	pdr := &sheets.PasteDataRequest{
		Data: j.ToHtml(),
		Coordinate: &sheets.GridCoordinate{
			SheetId:     sheetId,
			RowIndex:    6,
			ColumnIndex: 0,
		},
		Type: "PASTE_NORMAL",
		Html: true,
	}

	request := &sheets.BatchUpdateSpreadsheetRequest{
		Requests: []*sheets.Request{
			&sheets.Request{
				InsertDimension: idr,
			},
		},
	}

	_, err := s.Client.service.Spreadsheets.BatchUpdate(s.Client.spreadsheetId, request).Context(context.Background()).Do()
	if err != nil {
		log.Fatal(err)
	}

	request = &sheets.BatchUpdateSpreadsheetRequest{
		Requests: []*sheets.Request{
			&sheets.Request{
				PasteData: pdr,
			},
		},
	}

	_, err = s.Client.service.Spreadsheets.BatchUpdate(s.Client.spreadsheetId, request).Context(context.Background()).Do()
	if err != nil {
		log.Fatal(err)
	}
}
