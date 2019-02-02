package main

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
)

// struct representing the html table body on the osrs highscores page
type TableBody struct {
	XMLName       xml.Name   `xml:"tbody" json:"-"`
	TableRowsList []TableRow `xml:"tr" json:"PageData"`
}

// struct representing a table row within the html table body
type TableRow struct {
	XMLName       xml.Name    `xml:"tr" json:"-"`
	TableDataList []TableData `xml:"td" json:"RowData"`
}

// struct representing a data value in the tsable row
type TableData struct {
	XMLName xml.Name `xml:"td" json:"-"`
	Value   string   `xml:",chardata" json:"value"`
	Name    string   `xml:"a,omitempty" json:"name,omitempty"`
}

// struct for storing the scraped highscores list in
type Highscores struct {
	Users []UserData
}

// struct for storing scraped user data in
type UserData struct {
	Rank  int
	Name  string
	Level int
	XP    int
}

// getPageContentByPageNumber sends a GET request to the osrs highscores
// at the provided page number and scrapges the page data.
// @param pageNum - the page of highscores being scraped.
// @returns content read from the page
func getPageContentByPageNumber(pageNum int) ([]byte, error) {
	// send get request to the url
	resp, err := http.Get(fmt.Sprintf("https://secure.runescape.com/m=hiscore_oldschool/overall.ws?table=0&page=%d", pageNum))
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
	// remove commas which are used for number formatting
	tbodyContentString = strings.Replace(tbodyContentString, ",", "", -1)

	return []byte(tbodyContentString)
}

func main() {
	HTMLData, err := getPageContentByPageNumber(1)
	if err != nil {
		log.Printf("Failed to get XML: %v\n", err)
	} else {
		log.Println("Received XML!")

		cleanedTableBodyData := getCleanedTableBodyData(HTMLData)

		fmt.Printf("%#v", string(cleanedTableBodyData))

		var tb TableBody
		xml.Unmarshal(cleanedTableBodyData, &tb)

		//loop through table data, and store in highscores
		var hs Highscores
		for _, rd := range tb.TableRowsList {
			rdList := rd.TableDataList

			rank, _ := strconv.Atoi(rdList[0].Value)
			name := rdList[1].Name
			level, _ := strconv.Atoi(rdList[2].Value)
			xp, _ := strconv.Atoi(rdList[3].Value)
			hs.Users = append(hs.Users, UserData{Rank: rank, Name: name, Level: level, XP: xp})
		}

		jsonData, _ := json.Marshal(hs)
		log.Println(string(jsonData))
	}
}
