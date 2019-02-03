package main

import (
	"bytes"
	"context"
	"encoding/xml"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/mongodb/mongo-go-driver/bson"
	"github.com/mongodb/mongo-go-driver/mongo"
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
	Users []UserData `json:"Highscores" bson:"Users"`
}

// struct for storing scraped user data in
type UserData struct {
	Rank  int    `bson:"Rank"`
	Name  string `bson:"Name"`
	Level int    `bson:"Level"`
	XP    int    `bson:"XP"`
}

// getPageContentByPageNumber sends a GET request to the osrs highscores
// at the provided page number and scrapges the page data. If no table body
// is found, returns an empty byte slice.
// @param pageNum - the page of highscores being scraped.
// @returns content read from the page
func getPageContentByPageNumber(pageNum int) ([]byte, error) {
	// send get request to the url
	resp, err := http.Get(fmt.Sprintf("https://secure.runescape.com/m=hiscore_oldschool/overall.ws?table=0&page=%d", pageNum))
	if err != nil {
		return []byte{}, fmt.Errorf("GET error: %v", err)
	}
	defer resp.Body.Close()
	// response from webpage was not OK
	if resp.StatusCode != http.StatusOK {
		return []byte{}, fmt.Errorf("Status error: %v", resp.StatusCode)
	}
	// if the response is OK, read the webpage
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return []byte{}, fmt.Errorf("Read body: %v", err)
	}
	// return read data
	return data, nil
}

// getCleanedTableBodyData takes the html bytes from the osrs highscores
// page, and retrieves the table body section, and replaces all extra
// unwated tags to allow for the data to be parsed as xml
// @param HTMLData - html data being cleaned
// @returns cleaned html table body found in the XMLData
func getCleanedTableBodyData(HTMLData []byte) ([]byte, error) {
	tbodyStartIndex := bytes.Index(HTMLData, []byte("<tbody>"))
	tbodyEndIndex := bytes.Index(HTMLData, []byte("</tbody>")) + 8
	if tbodyStartIndex == -1 || tbodyEndIndex == -1 {
		fmt.Printf("%#V", string(HTMLData))
		return []byte{}, errors.New("no body found")
	}
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

	return []byte(tbodyContentString), nil
}

// getTableBodyStructFromXML takes cleaned bodyData, and parses the XML
// into the Table body struct.
// @param cleanedTablBodyData - XML table body
// @returns Table Body struct containing the parsed data
func getTableBodyStructFromXML(cleanedTableBodyData []byte) TableBody {
	var tb TableBody
	xml.Unmarshal(cleanedTableBodyData, &tb)
	// remove first row form table body since it is always empty
	tr := tb.TableRowsList
	tb.TableRowsList = tr[1:]
	return tb
}

// addUsersToHighscores takes a pointer to the highscores struct and a table
// body struct containing parsed user data, and puts the user data in the highscores.
// @param hs - pointer to highscores struct
// @param tb - table body struct containing parsed data
func addUsersToHighscores(hs *Highscores, tb TableBody) {
	for _, rd := range tb.TableRowsList {
		rdList := rd.TableDataList

		rank, _ := strconv.Atoi(rdList[0].Value)
		name := rdList[1].Name
		level, _ := strconv.Atoi(rdList[2].Value)
		xp, _ := strconv.Atoi(rdList[3].Value)
		hs.Users = append(hs.Users, UserData{Rank: rank, Name: name, Level: level, XP: xp})
	}
}

// loadMongoClient loads the mongo client on localhost
// @returns a mongo client for the local host
func loadMongoClient() (*mongo.Client, error) {
	client, err := mongo.NewClient("mongodb://localhost:27017")
	if err != nil {
		fmt.Println(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	err = client.Connect(ctx)
	if err != nil {
		fmt.Println(err)
	}
	return client, err
}

// retrieveMongoCollection takes a mongo client, database name, and collection name
// and returns the mongo collection, to be used for CRUD operations
func retrieveMongoCollection(client *mongo.Client, db string, collection string) *mongo.Collection {
	col := client.Database(db).Collection(collection)
	return col
}

// retrieveHighscoreData takes a pointer to the Highscores struct
// as well as a page number that will be scraped from the osrs highscores.
// it then scrapes that data, and adds it to the list of users data in the
// Highscores struct.
// @param hs - pointer to Highscores struct where scraped data is being stored
// @param page- the osrs highscores page number being scraped
func retrieveHighscoreData(hs *Highscores, page int) {
	// get page data
	HTMLData, err := getPageContentByPageNumber(page)
	if err != nil {
		log.Printf("Failed to get XML: %v\n", err)
		wg.Done()
		return
	}
	// clean the data
	cleanedTableBodyData, err := getCleanedTableBodyData(HTMLData)
	if err != nil {
		fmt.Println("error cleaning the data: ")
		//wg.Done()
	} else {
		// read XML into struct
		tb := getTableBodyStructFromXML(cleanedTableBodyData)
		// store formatted data in highscores struct
		addUsersToHighscores(hs, tb)
	}

	wg.Done()
}

// writeHighscoresToMongo takes the passed Highscores struct, and writes
// it into the mongo db running on localhost:27017.
// it writes the data to the db: osrs collection: highscores
// with the key being the current timestamp, and val the scraped user data
// @param hs - Highscores struct being written to db
func writeHighscoresToMongo(hs *Highscores) {
	mongodbClient, err := loadMongoClient()
	if err != nil {
		fmt.Println("error getting client")
	}

	osrshsCollection := retrieveMongoCollection(mongodbClient, "osrs", "highscores")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	res, err := osrshsCollection.InsertOne(ctx, bson.M{"test": hs.Users})
	id := res.InsertedID
	fmt.Printf("finished writing to db, the resulting id is: %s", id)
}

var wg sync.WaitGroup

func main() {
	var hs Highscores

	overallStartTime := time.Now()
	batch := 0
	for batch < 1 {

		startPage := batch*1000 + 1
		endPage := (batch + 1) * 1000

		for startPage <= endPage {
			wg.Add(1)
			go retrieveHighscoreData(&hs, startPage)
			startPage++
		}

		wg.Wait()
		batch++
	}

	overallFinishTime := time.Now()
	overallDiffTime := overallFinishTime.Sub(overallStartTime)
	fmt.Printf("It took %s to retrieve the data from %d pages\n", overallDiffTime, 10000)

	writeHighscoresToMongo(&hs)
}
