package models

type Patent struct {
	PatentID   string   `json:"patent_id"`
	Title      string   `json:"title"`
	Abstract   string   `json:"abstract"`
	FilingDate string   `json:"filing_date"`
	GrantDate  string   `json:"grant_date"`
	Inventors  []string `json:"inventors"`
	Assignee   string   `json:"assignee"`
	Country    string   `json:"country"`
	URL        string   `json:"url"`
	Claims     []string `json:"claims,omitempty"`
	CPCCodes   []string `json:"cpc_codes,omitempty"`
	Citations  []string `json:"citations,omitempty"`
}
