package main

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"time"

	"golang.org/x/net/html"
)

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
