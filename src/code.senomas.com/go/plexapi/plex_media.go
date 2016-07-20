package plexapi

import "encoding/xml"

// Video struct
type Video struct {
	Server                *Server
	XMLName               xml.Name `xml:"Video"`
	GUID                  string   `xml:"guid,attr"`
	RatingKey             string   `xml:"ratingKey,attr"`
	Key                   string   `xml:"key,attr"`
	Studio                string   `xml:"studio,attr"`
	Type                  string   `xml:"type,attr"`
	Title                 string   `xml:"title,attr"`
	TitleSort             string   `xml:"titleSort,attr"`
	ContentRating         string   `xml:"contentRating,attr"`
	Summary               string   `xml:"summary,attr"`
	Rating                string   `xml:"rating,attr"`
	ViewCount             string   `xml:"viewCount,attr"`
	ViewOffset            string   `xml:"viewOffset,attr"`
	LastViewedAt          string   `xml:"lastViewedAt,attr"`
	Year                  string   `xml:"year,attr"`
	Tagline               string   `xml:"tagline,attr"`
	Thumb                 string   `xml:"thumb,attr"`
	Art                   string   `xml:"art,attr"`
	Duration              string   `xml:"duration,attr"`
	OriginallyAvailableAt string   `xml:"originallyAvailableAt,attr"`
	AddedAt               string   `xml:"addedAt,attr"`
	UpdatedAt             string   `xml:"updatedAt,attr"`
	ChapterSource         string   `xml:"chapterSource,attr"`
}

// Genre struct
type Genre struct {
	XMLName xml.Name `xml:"Genre"`
	Tag     string   `xml:"tag,attr"`
}

// Writer struct
type Writer struct {
	XMLName xml.Name `xml:"Writer"`
	Tag     string   `xml:"tag,attr"`
}

// Country struct
type Country struct {
	XMLName xml.Name `xml:"Country"`
	Tag     string   `xml:"tag,attr"`
}

// Role struct
type Role struct {
	XMLName xml.Name `xml:"Role"`
	Tag     string   `xml:"tag,attr"`
}

// Director struct
type Director struct {
	XMLName xml.Name `xml:"Director"`
	Tag     string   `xml:"tag,attr"`
}
