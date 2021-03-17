package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"golang.org/x/net/html"
)

/* tableStructureCheckAndCleanup is a method that compares column headers provided by user with column headers scraped by tokenizer. Function also deletes tables not matching those headers
Args:
- string (comma separated list) of required column headers for scraped table
*/
func (tables *TableContainer) tableStructureCheckAndCleanup(tableHeaders string) {
	// Split given headers in csv format to slice
	providedColumnHeaders := strings.Split(tableHeaders, ",")
	// Loop throught all tables that tokenizer managed to parse and that are saved as TableContainer struct
	// Loop is doing reverse lookup - from len()-1 index to 0, because we are also doing deletions from slice if table is not valid
	for i := len(tables.Table) - 1; i >= 0; i-- {

		// Delete scraped table if number of provided column headers is less than number of column headers of a single table
		if len(tables.Table[i].ColumnHeaders) < len(providedColumnHeaders) {
			// Remove the table from slice
			tables.deleteTableFromSlice(i)
			continue
		}

		// Check if provided collumn headers match table column headers scraped by tokenizer
		for j := range providedColumnHeaders {
			if !(providedColumnHeaders[j] == tables.Table[i].ColumnHeaders[j]) {
				// If it does not match, delete table as the table is not valid format
				tables.deleteTableFromSlice(i)
				break
			}
		}
	}
	if len(tables.Table) == 0 {
		log.Fatal("No tables matching provided filter. Check values from --table-headers flag and HTML <table> structure")
	}
}

/* deleteTableFromSlice is a method that replaces current table at index with last table in slice and returns the same slice without last element (duplicate table)
Args:
- int that represents index of item in slice
*/
func (tables *TableContainer) deleteTableFromSlice(index int) {
	// Copy last table to index of current one
	tables.Table[index] = tables.Table[len(tables.Table)-1]
	// Truncate slice - remove last table becase we already have a copy of it on current index
	tables.Table = tables.Table[:len(tables.Table)-1]
}

/* convertToJSON is a method that converts TableContainer sturct to JSON. It creates a map where keys are table column headers and values are data inside each row
Return values:
- TableContainerJSON struct, containing each table as JSON object
*/
func (tables *TableContainer) convertToJSON() TableContainerJSON {
	jsonOutput := TableContainerJSON{}

	// Loop through all tables
	for _, table := range tables.Table {
		// Create slice of maps containing all rows of a given table
		outputObjects := []map[string]string{}
		// For each table, loop through all of its rows
		for _, rows := range table.TableRows {
			// Create a map containing all data of a given row with column names as keys
			valueMapping := map[string]string{}
			// For each data in a row, save that data to a map with key matching column header name
			for j, item := range rows {
				valueMapping[table.ColumnHeaders[j]] = item
			}
			outputObjects = append(outputObjects, valueMapping)
		}
		jsonOutput.Tables = append(jsonOutput.Tables, outputObjects)
	}

	return jsonOutput
}

/* getKVPairs is a method that filters out JSON data for provided filter and returns KV paris
Args:
- string with 2 values that are separated by : and will be extracted from table
Return values:
- map[string]string containing KV pairs
*/
func (jsonContainer *TableContainerJSON) getKVPairs(kvFilter string) map[string]string {
	// Initialize map that will be retured by function
	ipHost := map[string]string{}
	// Split string values to slice
	kvListToExtract := strings.Split(kvFilter, ":")

	// Iterate through all tables and rows
	for _, table := range jsonContainer.Tables {
		for _, object := range table {
			// Save matching keys/values to new map
			ipHost[object[kvListToExtract[0]]] = object[kvListToExtract[1]]
		}
	}
	// Cleanup new map, delete values that are empty
	for k, v := range ipHost {
		if v == "" {
			delete(ipHost, k)
		}
	}

	return ipHost
}

func getPageAsJSON(config CLIflags) Content {
	// Create new request object
	confluenceRESTAPIEndpoint := "https://" + config.ConfluenceDomain + "/wiki/rest/api/content/" + strconv.Itoa(int(config.ConfluencePageID)) + "?expand=body.storage.value"
	req, err := http.NewRequest("GET", confluenceRESTAPIEndpoint, nil)
	if err != nil {
		log.Fatal("Failed creating request object. Exiting. Error: ", err)
	}

	req.SetBasicAuth(config.ConfluenceUser, config.ConfluenceAPIKey)

	client := &http.Client{Timeout: time.Second * 10}
	// Send request
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal("Failed creating request. Exiting. Error: ", err)
	}
	if resp.StatusCode != 200 {
		log.Fatal("Failed scraping webpage. HTTP status code: ", resp.StatusCode)
	}

	// Read Request and save it to struct
	jsonResponse := Content{}
	rawData, err := ioutil.ReadAll(resp.Body) // Read response from io.ReadCloser and save it as byte to data
	if err != nil {
		log.Fatal("Failed reading response. Exiting. Error: ", err)
	}
	err = json.Unmarshal(rawData, &jsonResponse) // Try to unmarshal to raw encoded JSON value to map, notice &
	if err != nil {
		log.Fatal("Failed converting response to JSON. Exiting. Error: ", err)
	}
	// Close Request
	resp.Body.Close()

	return jsonResponse
}

func scrapeTablesFromHTML(webpageHTML io.Reader) TableContainer {
	// Initialize variables used in tokenization
	var tableRow []string
	var loopNum int = 0

	// Main variables for storing extracted data
	tableContainer := TableContainer{}
	table := Table{}

	z := html.NewTokenizer(webpageHTML)
	for {
		// Go to next element and increase counter
		tt := z.Next()
		t := z.Token()
		// Keep track how many times did we loop to give table UID
		loopNum++

		// ErrorToken is if we reached end of HTML document, so we return all tables to main function
		if tt == html.ErrorToken {
			if len(tableContainer.Table) == 0 {
				log.Fatal("Scraped Webpage does not contain any valid <table> elements.")
			} else {
				return tableContainer
			}
		}

		// Search for start of table - <table> tag
		if t.Data == "table" && tt == html.StartTagToken {

			// Cleanup Table struct for storing table values and set identifier
			table = Table{}
			table.ID = loopNum

			// Go to next element
			tt = z.Next()
			t = z.Token()
			// Loop untill we hit table again, but this time it is closing table tag </table>. This means we do not go back, untill we go throught whole table
			for t.Data != "table" {
				if t.Data == "tr" && tt == html.StartTagToken {
					tableRow = []string{}
					tt = z.Next()
					t = z.Token()
					// Loop whole row to get all data inside this row
					for t.Data != "tr" {
						if t.Data == "th" && tt == html.StartTagToken {
							tt = z.Next()
							t = z.Token()
							// Loop whole TH to get data
							for t.Data != "th" {
								// Get only text values to extract headers from table, ignore all other tags. Only plain text will be saved and that is collum name
								if tt == html.TextToken {
									table.ColumnHeaders = append(table.ColumnHeaders, t.Data)
								}
								// Go to next element and check again if we reached end of thead
								tt = z.Next()
								t = z.Token()
							}
						}

						if t.Data == "td" && tt == html.StartTagToken {
							// Set tableRow to empty slice, since each row will have its own data inside <td>
							// Go to next element
							tt = z.Next()
							t = z.Token()
							// Iterate further untill we hit tr again, this means we got to </tr>. We are not exiting loop untill we get raw data (html.TextToken). This data is only inside <td>
							for t.Data != "td" {
								if tt == html.TextToken {
									// If we found raw data, we apped it to slice of that row
									tableRow = append(tableRow, t.Data)
								}
								if tt == html.SelfClosingTagToken {
									tableRow = append(tableRow, "")
								}
								// And then we go to next element
								tt = z.Next()
								t = z.Token()
							}
						}
						tt = z.Next()
						t = z.Token()
					}
					// Once we found closing </tr> tag, we apped row to slice and go further untill we hit a new row
					table.TableRows = append(table.TableRows, tableRow)
				}
				tt = z.Next()
				t = z.Token()
			}
			// On closing </table> tag, append scraped table to TableContainer
			if t.Data == "table" && tt == html.EndTagToken {
				tableContainer.Table = append(tableContainer.Table, table)
			}
		}
	}
}

func main() {
	// Set variables for CLI flags
	confluenceScrapeConfig := CLIflags{}

	// Create and parse CLI flags
	flag.StringVar(&confluenceScrapeConfig.ConfluenceDomain, "confl-domain", "", "Confluence domain on atlassian.net")
	flag.Int64Var(&confluenceScrapeConfig.ConfluencePageID, "confl-pageid", 0, "PageID on atlassian.net")
	flag.StringVar(&confluenceScrapeConfig.ConfluenceUser, "confl-user", "", "Confluence Username")
	flag.StringVar(&confluenceScrapeConfig.ConfluenceAPIKey, "confl-apikey", "", "Confluence API Key")
	flag.StringVar(&confluenceScrapeConfig.TableHeaders, "table-headers", "", "Comma-separated list of table headers against each table on web will be compared. Case sensitive")
	flag.StringVar(&confluenceScrapeConfig.KVList, "kv-list", "", "Comma-separated list of 2 items that will be extracted. Case sensitive")
	flag.StringVar(&confluenceScrapeConfig.Output, "output", "st	dout", "Comma-separated list of 2 items that will be extracted. Case sensitive")
	flag.Parse()
	log.Println("CLI flags successfuly initialized. Fetching website ...")

	// Call function to scrape webpage
	//confluencePage := getPageAsJSON(confluenceScrapeConfig)
	log.Println("Succesfuly fetched "+confluenceScrapeConfig.ConfluenceDomain+" page:", strconv.Itoa(int(confluenceScrapeConfig.ConfluencePageID))+". Starting tokenization ...")

	// Get page body as string out of JSON response
	pageBody := getPageAsJSON(confluenceScrapeConfig)
	// Call function to tokenize HTML and filters out tables
	byteData := []byte(pageBody.Body.Storage.Value)

	r := bytes.NewReader(byteData)

	allTables := scrapeTablesFromHTML(r)
	log.Println("Succesfuly finished tokenization.")

	//fmt.Printf("%+v\n", allTables)

	// Filter out tables that are not needed
	allTables.tableStructureCheckAndCleanup(confluenceScrapeConfig.TableHeaders)
	log.Println("Succesfuly finished table filtering. Tables matching filter:", len(allTables.Table))

	// Get table data as JSON
	jsonOutput := allTables.convertToJSON()

	fmt.Println(jsonOutput)

	//fmt.Printf("%v\n", allTables)
	b, err := json.Marshal(jsonOutput.getKVPairs(confluenceScrapeConfig.KVList))
	if err != nil {
		fmt.Println(err)
	}
	log.Println("Writing tables as JSON to file ...")
	err = ioutil.WriteFile("/tmp/dat1", b, 0644)
	if err != nil {
		fmt.Println(err)
	}
	log.Println("Successfuly written tables to /tmp/dat1")
}
