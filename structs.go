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
	LogLevel         bool
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

// TFVars is a struct defining output format if --output=tfvars is set
type TFVars struct {
	AWXInventories map[string]Inventories `json:"awx_inventories"`
}

// Inventories is a struct defining each invertory for output in TFVars
type Inventories struct {
	DocsURL string            `json:"docs_url"`
	Hosts   map[string]string `json:"hosts"`
}

// Confluence Response data structure
// Content is a struct containing properties of JSON response recieved by Confluence API
type Content struct {
	ID    string `json:"id,omitempty"`
	Type  string `json:"type"`
	Title string `json:"title"`
	Body  Body   `json:"body"`
	Links Links  `json:"_links,omitempty"`
}

// Body struct holds the storage information and is nested under Content
type Body struct {
	Storage Storage `json:"storage"`
}

// Storage struct .Value holds the real <body> elemet of HTML and is nested under Body
type Storage struct {
	Value string `json:"value"`
}

// Links struct contains link information
type Links struct {
	Base    string `json:"base"`
	TinyUI  string `json:"tinyui"`
	WebUI   string `json:"webui"`
	Content string `json:"context"`
}
