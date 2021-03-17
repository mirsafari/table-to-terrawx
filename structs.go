package main

// User parameters data structure
// CLIflags is a struct that defines script config file structure and provides storage for CLI flags
type CLIflags struct {
	ConfluenceDomain string
	ConfluencePageID int64
	ConfluenceUser   string
	ConfluenceAPIKey string
	TableHeaders     string
	KVList           string
	Output           string
}

// Data structures for parsing table data
// TableContainer is a struct containing all tables scraped from given webpage
type TableContainer struct {
	Table []Table
}

// Table is a helper struct containing data for single table
type Table struct {
	ID            int
	ColumnHeaders []string
	TableRows     [][]string
}

// TableContainerJSON is a struct containing filtered tables scraped from webpage ready to be outputed. Each table is a JSON object, where array of data is located and keys are column headers
type TableContainerJSON struct {
	Tables [][]map[string]string
}

// Confluence Response data structure
// Content is a struct containing properties of JSON response recieved by Confluence API
type Content struct {
	ID    string `json:"id,omitempty"`
	Type  string `json:"type"`
	Title string `json:"title"`
	Body  Body   `json:"body"`
}

// Body struct holds the storage information and is nested under Content
type Body struct {
	Storage Storage `json:"storage"`
}

// Storage struct .Value holds the real <body> elemet of HTML and is nested under Body
type Storage struct {
	Value string `json:"value"`
}
