package main

import (
	"log"
	"strings"
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
