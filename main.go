//https://stackoverflow.com/questions/35961491/how-to-convert-html-table-to-array-with-golang
package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
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

// TableContainer is a struct containing all tables scraped from given webpage
type TableContainer struct {
	Table []Table
}

// Table is a struct containing data for single table
type Table struct {
	ID            int
	ColumnHeaders []string
	TableRows     [][]string
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
func tableStructureCheckAndCleanup(tables *TableContainer, tableHeaders string) {
	providedColumnHeaders := strings.Split(tableHeaders, ",")
	// Loop throught all tables that tokenizer managed to parse and that are saved as TableContainer struct
	// Loop is doing reverse lookup - from len()-1 index to 0, because we are also doing deletions from slice if table is not valid
	for i := len(tables.Table) - 1; i >= 0; i-- {
		// Delete scraped table if number of provided column headers does not match the number of column headers of a single table
		if len(tables.Table[i].ColumnHeaders) < len(providedColumnHeaders) {
			// Remove the table from slice
			tables.Table = deleteTableFromSlice(tables.Table, i)
			continue
		}
		// Check if collumn headers match provided column headers
		for j := range providedColumnHeaders {
			if !(providedColumnHeaders[j] == tables.Table[i].ColumnHeaders[j]) {
				// If it does not match, delete table as the table is not valid format
				tables.Table = deleteTableFromSlice(tables.Table, i)
				continue
			}
		}

	}

	if len(tables.Table) == 0 {
		log.Fatal("Scraped tables do not match provided filter or no tables")
	}
}

// Function deleteTableFromSlice replaces current table at index with last table in slice and returns the same slice without last element (duplicate table)
func deleteTableFromSlice(tables []Table, index int) []Table {
	// Copy last table to index of current one
	tables[index] = tables[len(tables)-1]
	// Truncate slice - remove last table becase we already have a copy of it on current index
	return tables[:len(tables)-1]
}

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

		// ErrorToken is if we reached end
		if tt == html.ErrorToken {
			// Check if table is valid sturcture

			tableStructureCheckAndCleanup(&tableContainer, targetWeb.TableHeaders)
			fmt.Printf("%v\n", tableContainer)
			return
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
				// Finding <thead> to extract collumn names.
				if t.Data == "thead" && tt == html.StartTagToken {
					// Go to next element inside thead
					tt = z.Next()
					t = z.Token()
					// Loop untill we hit thead again, but this time it is closing thead tag </thead>. This means we do not go back untill we go throught whole thead and get all collumn names
					for t.Data != "thead" {
						// Get only text values to extract headers from table, ignore all other tags. Only plain text will be saved and that is collum name
						if tt == html.TextToken {
							table.ColumnHeaders = append(table.ColumnHeaders, t.Data)
						}
						// Go to next element and check again if we reached end of thead
						tt = z.Next()
						t = z.Token()
					}
				}
				// Once we found thead again, it means that this was closing tag </thead> and we got our column names so we can continue getting the data

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
					// Once we found closing </tr> tag, we apped row to slice and go further untill we hit a new row
					table.TableRows = append(table.TableRows, tableRow)
				}
				tt = z.Next()
				t = z.Token()
			}
		}
		// On closing </table> tag, append scraped table to TableContainer
		if t.Data == "table" && tt == html.EndTagToken {
			tableContainer.Table = append(tableContainer.Table, table)
		}
	}
}
