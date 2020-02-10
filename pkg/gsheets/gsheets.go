package gsheets

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"

	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/sheets/v4"
)

type Client struct {
	service       *sheets.Service
	spreadsheetId string
}

func NewClient() Client {
	c := Client{
		service: getService(),
		// NOTE(mandre) use a copy of the CI spreadsheet for testing
		spreadsheetId: "19sV5IvC2xL8yC86ELaD8P30TKyiC1Okv0Ikr-ohRisI",
	}
	return c
}

func getService() *sheets.Service {
	b, err := ioutil.ReadFile("credentials.json")
	if err != nil {
		log.Fatalf("Unable to read client secret file: %v", err)
	}

	// If modifying these scopes, delete your previously saved token.json.
	config, err := google.ConfigFromJSON(b, "https://www.googleapis.com/auth/spreadsheets")
	if err != nil {
		log.Fatalf("Unable to parse client secret file to config: %v", err)
	}
	client := getClient(config)

	srv, err := sheets.New(client)
	if err != nil {
		log.Fatalf("Unable to retrieve Sheets client: %v", err)
	}
	return srv
}

// Retrieve a token, saves the token, then returns the generated client.
func getClient(config *oauth2.Config) *http.Client {
	// The file token.json stores the user's access and refresh tokens, and is
	// created automatically when the authorization flow completes for the first
	// time.
	tokFile := "token.json"
	tok, err := tokenFromFile(tokFile)
	if err != nil {
		tok = getTokenFromWeb(config)
		saveToken(tokFile, tok)
	}
	return config.Client(context.Background(), tok)
}

// Request a token from the web, then returns the retrieved token.
func getTokenFromWeb(config *oauth2.Config) *oauth2.Token {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Go to the following link in your browser then type the "+
		"authorization code: \n%v\n", authURL)

	var authCode string
	if _, err := fmt.Scan(&authCode); err != nil {
		log.Fatalf("Unable to read authorization code: %v", err)
	}

	tok, err := config.Exchange(context.TODO(), authCode)
	if err != nil {
		log.Fatalf("Unable to retrieve token from web: %v", err)
	}
	return tok
}

// Retrieves a token from a local file.
func tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	tok := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(tok)
	return tok, err
}

// Saves a token to a file path.
func saveToken(path string, token *oauth2.Token) {
	fmt.Printf("Saving credential file to: %s\n", path)
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Fatalf("Unable to cache oauth token: %v", err)
	}
	defer f.Close()
	json.NewEncoder(f).Encode(token)
}

func getSheetID(jobName string) int64 {
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

	sheetId, ok := innerMap[jobName]
	if !ok {
		log.Fatalf("Unknown job %v", jobName)
	}

	return sheetId
}

func getSheetName(jobName string) string {
	innerMap := map[string]string{
		"release-openshift-ocp-installer-e2e-openstack-4.4":           "Parallel OCP 4.4",
		"release-openshift-ocp-installer-e2e-openstack-serial-4.4":    "Serial OCP 4.4",
		"release-openshift-ocp-installer-e2e-openstack-4.3":           "Parallel OCP 4.3",
		"release-openshift-ocp-installer-e2e-openstack-serial-4.3":    "Serial OCP 4.3",
		"release-openshift-ocp-installer-e2e-openstack-4.2":           "Parallel OCP 4.2",
		"release-openshift-ocp-installer-e2e-openstack-serial-4.2":    "Serial OCP 4.2",
		"release-openshift-origin-installer-e2e-openstack-4.2":        "Parallel Origin 4.3",
		"release-openshift-origin-installer-e2e-openstack-serial-4.2": "Serial Origin 4.4",
	}

	sheetName, ok := innerMap[jobName]
	if !ok {
		log.Fatalf("Unknown job %v", jobName)
	}

	return sheetName
}

// Get ID of the most recent job in the spreadsheet
func (c *Client) GetLatestIdFromSheet(jobName string) int {
	readRange := getSheetName(jobName) + "!A7"
	resp, err := c.service.Spreadsheets.Values.Get(c.spreadsheetId, readRange).Do()
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

func (c *Client) AddRow(row, jobName string) {
	sheetId := getSheetID(jobName)

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
		Data: row,
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

	_, err := c.service.Spreadsheets.BatchUpdate(c.spreadsheetId, request).Context(context.Background()).Do()
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

	_, err = c.service.Spreadsheets.BatchUpdate(c.spreadsheetId, request).Context(context.Background()).Do()
	if err != nil {
		log.Fatal(err)
	}
}
