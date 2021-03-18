package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"strconv"

	log "github.com/sirupsen/logrus"
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
	flag.StringVar(&confluenceScrapeConfig.Output, "output", "stdout", "How results should be displayed. stdout or tfvars")
	flag.BoolVar(&confluenceScrapeConfig.LogLevel, "debug", false, "Show debug information")
	flag.Parse()

	// Enable debug logging
	if confluenceScrapeConfig.LogLevel == true {
		log.SetFormatter(&log.TextFormatter{
			FullTimestamp: true,
		})
		log.SetLevel(log.DebugLevel)
	}

	log.Debug("CLI flags successfuly initialized. Fetching website ...")

	// Call function to scrape webpage
	pageBodyAndMetadata := getPageAsJSON(confluenceScrapeConfig)
	log.Debug("Succesfuly fetched "+confluenceScrapeConfig.ConfluenceDomain+" page:", strconv.Itoa(int(confluenceScrapeConfig.ConfluencePageID))+". Starting tokenization ...")

	// Convert pageBody.Body.Storage.Value to byte, so we can create ioReader based of it, which is a required type for HTML Tokenizer
	// pageBody.Body.Storage.Value contains HTML of the page - only data inside <body> elemet
	bodyElementHTML := []byte(pageBodyAndMetadata.Body.Storage.Value)
	bodyElementBytes := bytes.NewReader(bodyElementHTML)

	// Call function to get all tables from given HTML
	tables := scrapeTablesFromHTML(bodyElementBytes)
	log.Debug("Succesfuly finished tokenization.")

	// Filter out tables that are not needed
	tables.tableStructureCheckAndCleanup(confluenceScrapeConfig.TableHeaders)
	log.Debug("Succesfuly finished table filtering. Tables matching filter:", len(tables.Table))

	// Convert filtered tables to JSON
	jsonOutput := tables.convertToJSON()

	// Extract only wanted collums from tables
	filteredCollumns := jsonOutput.getKVPairs(confluenceScrapeConfig.KVList)

	// Output results
	switch confluenceScrapeConfig.Output {
	case "stdout":
		{
			b, err := json.Marshal(filteredCollumns)
			if err != nil {
				log.Fatal(err)
			}
			fmt.Println(string(b))
		}
	case "tfvars":
		{
			output := TFVars{}
			output.AWXInventories = make(map[string]Inventories)

			inventory := Inventories{}
			inventory.DocsURL = ("https://" + confluenceScrapeConfig.ConfluenceDomain + pageBodyAndMetadata.Links.Content + pageBodyAndMetadata.Links.WebUI)
			inventory.Hosts = filteredCollumns

			output.AWXInventories[pageBodyAndMetadata.Title] = inventory

			data, err := json.Marshal(output)

			if err != nil {
				log.Fatal(err)
			}
			fmt.Printf("%s\n", data)

		}
	}
}
