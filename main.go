package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"strconv"
)

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
