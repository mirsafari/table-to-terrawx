//https://stackoverflow.com/questions/35961491/how-to-convert-html-table-to-array-with-golang
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"golang.org/x/net/html"
)

// CLIflags is a struct that defines script config file structure and provides storage for CLI flags
type CLIflags struct {
	URL          string
	Username     string
	Password     string
	TableHeaders string
}

func getContent(config CLIflags) io.ReadCloser {
	// Create new request object
	req, err := http.NewRequest("GET", config.URL, nil)
	if err != nil {
		log.Fatal("Failed creating request object. Exiting. Error: ", err)
	}
	// Setup authentication if username and password provided
	if len(config.Username) > 0 && len(config.Password) > 0 {
		req.SetBasicAuth(config.Username, config.Password)
	}

	client := &http.Client{Timeout: time.Second * 10}
	// Send request
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal("Failed creating request. Exiting. Error: ", err)
	}
	if resp.StatusCode != 200 {
		log.Fatal("Failed scraping webpage. HTTP status code", resp.StatusCode)
	}
	return resp.Body
}

/* tableStructureCheckAndCleanup checks all tables fetched by tokenizer and compares table headers to provided headers that needed to be extracted
Function arguments:
- map cotaining allTables
- string of headers that are comma separated
Return values:
- stripped map containing only tables matching the headers
*/
func tableStructureCheckAndCleanup(allTables map[int][][]string, tableHeaders string) map[int][][]string {
	headers := strings.Split(tableHeaders, ",")

	// Loop throught all tables that tokenizer managed to parse
	for table, rows := range allTables {
		// Delete table if number of collums is less than provided collums and go to next iteration. Avoids index-out-of-range error
		if len(rows[0]) < len(headers) {
			delete(allTables, table)
			continue
		}
		// Check if first item in slice (First item always contains headers) matches provided filters
		for i := range headers {
			if !(headers[i] == rows[0][i]) {
				// If it does not match, delete table as the table is not valid format
				delete(allTables, table)
			}
		}
	}

	if len(allTables) == 0 {
		log.Fatal("Scraped tables do not match provided filter or no tables")
	}

	return allTables
}

/*func convertMapToJSON(allTables map[int][][]string) {

	//novaMapa := map[string]string{}

	for table := range allTables {
		fmt.Println(table)
	}
}
*/
func main() {
	// Set variables for CLI flags
	targetWeb := CLIflags{}

	// Create and parse CLI flags
	flag.StringVar(&targetWeb.URL, "url", "https://www.google.com", "URL of web to scrape")
	flag.StringVar(&targetWeb.Username, "username", "", "Username if web uses authentication")
	flag.StringVar(&targetWeb.Password, "password", "", "Password if web uses authentication")
	flag.StringVar(&targetWeb.TableHeaders, "table-headers", "", "Comma-separated list of table headers that needed to to be extracted. Case sensitive")
	flag.Parse()
	log.Printf("CLI flags successfuly initialized. Fetching website ...")

	// Call function to scrape webpage
	webpageHTML := getContent(targetWeb)
	defer webpageHTML.Close()
	log.Printf("Succesfuly fetched " + targetWeb.URL + ". Starting tokenization ...")

	// Initialize variables used in tokenization
	var tableColumns []string
	var tableRow []string
	var loopNum int = 0
	var tableIdentifier int

	// Main variable for storing alreay extracted data
	tableData := map[int][][]string{}

	z := html.NewTokenizer(webpageHTML)
	for {
		// Go to next element and increase counter
		tt := z.Next()
		t := z.Token()
		// Keep track how many times did we loop to give table UID
		loopNum++

		// ErrorToken is if we reached end
		if tt == html.ErrorToken {
			// Check if table is valid sturcture

			cleanedTables := tableStructureCheckAndCleanup(tableData, targetWeb.TableHeaders)

			enc := json.NewEncoder(os.Stdout)
			err := enc.Encode(cleanedTables)
			if err != nil {
				fmt.Println(err.Error())
			}
			//convertMapToJSON(tableStructureCheckAndCleanup(tableData, targetWeb.TableHeaders))
			return
		}

		// Search for start of table - <table> tag
		if t.Data == "table" && tt == html.StartTagToken {
			// Set table identifier
			tableIdentifier = loopNum
			// Cleanup table columns since they are written to data container. Each table elements has it's own colum names
			tableColumns = []string{}

			// Go to next element
			tt = z.Next()
			t = z.Token()
			// Loop untill we hit table again, but this time it is closing table tag </table>. This means we do not go back, untill we go throught whole table
			for t.Data != "table" {
				// Finding <thead> to extract collumn names.
				if t.Data == "thead" && tt == html.StartTagToken {
					// Go to next element inside thead
					tt = z.Next()
					t = z.Token()
					// Loop untill we hit thead again, but this time it is closing thead tag </thead>. This means we do not go back untill we go throught whole thead and get all collumn names
					for t.Data != "thead" {
						// Get only text values to extract headers from table, ignore all other tags. Only plain text will be saved and that is collum name
						if tt == html.TextToken {
							tableColumns = append(tableColumns, t.Data)
						}
						// Go to next element and check again if we reached end of thead
						tt = z.Next()
						t = z.Token()
					}
					// Add table columns to map
					tableData[tableIdentifier] = append(tableData[tableIdentifier], tableColumns)
				}

				// Since we found thead again, this means that it was closing tag </thead> and we got our column names
				// Next thing is to find raw data inside <tr> and <td> elements. First we rearch for rows
				if t.Data == "tr" && tt == html.StartTagToken {
					// Set tableRow to empty slice, since each row will have its own data inside <td>
					tableRow = []string{}
					// Go to next element
					tt = z.Next()
					t = z.Token()
					// Iterate further untill we hit tr again, this means we got to </tr>. We are not exiting loop untill we get raw data (html.TextToken). This data is only inside <td>
					for t.Data != "tr" {
						if tt == html.TextToken {
							// If we found raw data, we apped it to slice of that row
							tableRow = append(tableRow, t.Data)
						}
						// And then we go to next element
						tt = z.Next()
						t = z.Token()
					}
					// Once we found closing </tr> tag, we apped row to slice and go further untill we hit like 96 again and start going through new row
					tableData[tableIdentifier] = append(tableData[tableIdentifier], tableRow)
				}

				tt = z.Next()
				t = z.Token()
			}
		}

	}
}
