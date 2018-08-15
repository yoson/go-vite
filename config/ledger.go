package config

type Ledger struct {
	SendExplorer         bool     `json:"SendExplorer"`
	SendExplorerAddrs    []string `json:"SendExplorerAddrs"`
	SendExplorerFilename string   `json:"SendExplorerFilename"`
}
