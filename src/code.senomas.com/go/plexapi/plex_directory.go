package plexapi

import "encoding/xml"

// Directory struct
type Directory struct {
	Keys       []KeyInfo  `xml:"-"`
	Paths      []string   `xml:"-"`
	XMLName    xml.Name   `xml:"Directory"`
	Locations  []Location `xml:"Location"`
	Count      int        `xml:"count,attr"`
	Key        string     `xml:"key,attr"`
	Title      string     `xml:"title,attr"`
	Art        string     `xml:"art,attr"`
	Composite  string     `xml:"composite,attr"`
	Filters    int        `xml:"filters,attr"`
	Refreshing int        `xml:"refreshing,attr"`
	Thumb      string     `xml:"thumb,attr"`
	Type       string     `xml:"type,attr"`
	Agent      string     `xml:"agent,attr"`
	Scanner    string     `xml:"scanner,attr"`
	Language   string     `xml:"language,attr"`
	UUID       string     `xml:"uuid,attr"`
	UpdatedAt  string     `xml:"updatedAt,attr"`
	CreatedAt  string     `xml:"createdAt,attr"`
	AllowSync  int        `xml:"allowSync,attr"`
}

// Location struct
type Location struct {
	XMLName xml.Name `xml:"Location"`
	ID      int      `xml:"id,attr"`
	Path    string   `xml:"path,attr"`
}
