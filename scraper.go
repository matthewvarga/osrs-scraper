package main

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
)

type TableBody struct {
	XMLName       xml.Name   `xml:"tbody" json:"-"`
	TableRowsList []TableRow `xml:"tr" json:"row"`
}
type TableRow struct {
	XMLName       xml.Name `xml:"tr" json:"-"`
	TableDataList []string `xml:"td" json: "item"`
}

// getContent sends a GET request to the provided url, and returns
// the page data if possible.
// @param url - the url the request is being sent to
// @returns content read from the webpage
func getContent(url string) ([]byte, error) {
	// send get request to the url
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("GET error: %v", err)
	}
	defer resp.Body.Close()
	// response from webpage was not OK
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Status error: %v", resp.StatusCode)
	}
	// if the response is OK, read the webpage
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("Read body: %v", err)
	}
	// return read data
	return data, nil
}

// getCleanedTableBodyData takes the html bytes from the osrs highscores
// page, and retrieves the table body section, and replaces all extra
// unwated tags to allow for the data to be parsed as xml
// @param HTMLData - html data being cleaned
// @returns cleaned html table body found in the XMLData
func getCleanedTableBodyData(HTMLData []byte) []byte {
	tbodyStartIndex := bytes.Index(HTMLData, []byte("<tbody>"))
	tbodyEndIndex := bytes.Index(HTMLData, []byte("</tbody>")) + 8
	tbodyContent := HTMLData[tbodyStartIndex:tbodyEndIndex]

	tbodyContentString := string(tbodyContent)
	// remove all \" with ' since can't have double quotes inside of string
	tbodyContentString = strings.Replace(tbodyContentString, "\"", "'", -1)
	// remove encoded new lines
	tbodyContentString = strings.Replace(tbodyContentString, "\n", "", -1)
	// remove encoded spaces
	tbodyContentString = strings.Replace(tbodyContentString, "\xa0", " ", -1)

	return []byte(tbodyContentString)
}

func main() {
	HTMLData, err := getContent("https://secure.runescape.com/m=hiscore_oldschool/overall.ws?table=0&page=1")
	if err != nil {
		log.Printf("Failed to get XML: %v\n", err)
	} else {
		log.Println("Received XML!")

		highscoreData := getCleanedTableBodyData(HTMLData)

		var tb TableBody
		xml.Unmarshal(highscoreData, &tb)

		jsonData, _ := json.Marshal(tb)
		log.Println(string(jsonData))
	}
}
