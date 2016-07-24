package plexapi

import "encoding/xml"

// Video struct
type Video struct {
	Server                *Server   `xml:"-"`
	Keys                  []KeyInfo `xml:"-"`
	Paths                 []string  `xml:"-"`
	FID                   string    `xml:"-"`
	XMLName               xml.Name  `xml:"Video"`
	GUID                  string    `xml:"guid,attr"`
	RatingKey             string    `xml:"ratingKey,attr"`
	Key                   string    `xml:"key,attr"`
	Studio                string    `xml:"studio,attr"`
	Type                  string    `xml:"type,attr"`
	Title                 string    `xml:"title,attr"`
	TitleSort             string    `xml:"titleSort,attr"`
	ContentRating         string    `xml:"contentRating,attr"`
	Summary               string    `xml:"summary,attr"`
	Rating                string    `xml:"rating,attr"`
	ViewCount             string    `xml:"viewCount,attr"`
	ViewOffset            string    `xml:"viewOffset,attr"`
	LastViewedAt          string    `xml:"lastViewedAt,attr"`
	Year                  string    `xml:"year,attr"`
	Tagline               string    `xml:"tagline,attr"`
	Thumb                 string    `xml:"thumb,attr"`
	Art                   string    `xml:"art,attr"`
	Duration              string    `xml:"duration,attr"`
	OriginallyAvailableAt string    `xml:"originallyAvailableAt,attr"`
	AddedAt               string    `xml:"addedAt,attr"`
	UpdatedAt             string    `xml:"updatedAt,attr"`
	ChapterSource         string    `xml:"chapterSource,attr"`
	Media                 Media     `xml:"Media"`
	Genre                 Genre     `xml:"Genre"`
	Writer                Writer    `xml:"Writer"`
	Country               Country   `xml:"Country"`
	Role                  Role      `xml:"Role"`
	Director              Director  `xml:"Director"`
}

// Media struct
type Media struct {
	XMLName         xml.Name `xml:"Media"`
	VideoResolution string   `xml:"videoResolution,attr"`
	ID              string   `xml:"id,attr"`
	Duration        string   `xml:"duration,attr"`
	Bitrate         string   `xml:"bitrate,attr"`
	Width           string   `xml:"width,attr"`
	Height          string   `xml:"height,attr"`
	AspectRatio     string   `xml:"aspectRatio,attr"`
	AudioChannels   string   `xml:"audioChannels,attr"`
	AudioCodec      string   `xml:"audioCodec,attr"`
	VideoCodec      string   `xml:"videoCodec,attr"`
	Container       string   `xml:"container,attr"`
	VideoFrameRate  string   `xml:"videoFrameRate,attr"`
	VideoProfile    string   `xml:"videoProfile,attr"`
	Parts           []Part   `xml:"Part"`
}

// Part struct
type Part struct {
	XMLName             xml.Name `xml:"Part"`
	ID                  string   `xml:"id,attr"`
	Key                 string   `xml:"key,attr"`
	Duration            string   `xml:"duration,attr"`
	File                string   `xml:"file,attr"`
	Sizecontainer       string   `xml:"sizecontainer,attr"`
	DeepAnalysisVersion string   `xml:"deepAnalysisVersion,attr"`
	RequiredBandwidths  string   `xml:"requiredBandwidths,attr"`
	VideoProfile        string   `xml:"videoProfile,attr"`
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

// GetStatus func
func (v *Video) GetStatus() *ViewStatus {
	return &ViewStatus{
		LastViewedAt: v.LastViewedAt,
		ViewCount:    v.ViewCount,
		ViewOffset:   v.ViewOffset,
	}
}
