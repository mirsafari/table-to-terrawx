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
	URL            string
	Username       string
	Password       string
	TableHeaders   string
	ExtractColumns string
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

// TableContainerJSON is a struct containing filtered tables scraped from webpage ready to be outputed. Each table is a JSON object, where array of data is located and keys are column headers
type TableContainerJSON struct {
	Tables [][]map[string]string
}

/* tableStructureCheckAndCleanup is a method tgat compares column headers provided by user with column headers scraped by tokenizer. Function also deletes tables not matching those headers
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
				//fmt.Println(table.ColumnHeaders[i], "-->", item)
				valueMapping[table.ColumnHeaders[j]] = item
			}
			outputObjects = append(outputObjects, valueMapping)
		}
		jsonOutput.Tables = append(jsonOutput.Tables, outputObjects)
	}

	return jsonOutput
}

func getHTMLContent(config CLIflags) io.ReadCloser {
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

func scrapeTablesFromHTML(webpageHTML io.ReadCloser) TableContainer {
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
						if t.Data == "br" && tt == html.SelfClosingTagToken {
							// it means that it's an empty cell and we put empty string inside
							// this is to aviod mismatch in number of elemets
							tableRow = append(tableRow, "")
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

func main() {
	// Set variables for CLI flags
	targetWeb := CLIflags{}

	// Create and parse CLI flags
	flag.StringVar(&targetWeb.URL, "url", "https://www.google.com", "URL of web to scrape")
	flag.StringVar(&targetWeb.Username, "username", "", "Username if web uses authentication")
	flag.StringVar(&targetWeb.Password, "password", "", "Password if web uses authentication")
	flag.StringVar(&targetWeb.TableHeaders, "table-headers", "", "Comma-separated list of table headers against each table on web will be compared. Case sensitive")
	flag.StringVar(&targetWeb.ExtractColumns, "extract-columns", "", "Comma-separated list of table headers will be extracted from table. Case sensitive")
	flag.Parse()
	log.Println("CLI flags successfuly initialized. Fetching website ...")

	// Call function to scrape webpage
	webpageHTML := getHTMLContent(targetWeb)
	defer webpageHTML.Close()
	log.Println("Succesfuly fetched " + targetWeb.URL + ". Starting tokenization ...")

	// Call function to tokenize HTML and filters out tables
	allTables := scrapeTablesFromHTML(webpageHTML)
	log.Println("Succesfuly finished tokenization.")

	// Filter out tables that are not needed
	allTables.tableStructureCheckAndCleanup(targetWeb.TableHeaders)
	log.Println("Succesfuly finished table filtering. Tables matching filter:", len(allTables.Table))
	//fmt.Println(allTables)

	// Get table data as JSON
	jsonOutput := allTables.convertToJSON()

	extractColumns := strings.Split(targetWeb.ExtractColumns, ",")
	fmt.Println(extractColumns)

	for _, table := range jsonOutput.Tables {
		for k, v := range table {
			fmt.Println(k, v)
		}
	}
	/*
		//fmt.Printf("%v\n", allTables)
		b, err := json.Marshal(jsonOutput)
		if err != nil {
			fmt.Println(err)
			return
		}
		//fmt.Println(string(b))
	*/
}
