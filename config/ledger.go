package config

type Ledger struct {
	SendExplorer         bool     `json:"SendExplorer"`
	SendExplorerAddrs    []string `json:"SendExplorerAddrs"`
	SendExplorerFilename string   `json:"SendExplorerFilename"`
	SendExplorerTopic    string   `json:"SendExplorerTopic"`

	IsDownload bool `json:"IsDownload"`
}
