package model

import (
	"encoding/json"
	"time"
)

// MovieSearchResult represents the search result item returned to MetaTube plugin/client.
type MovieSearchResult struct {
	ID          string    `json:"id"`
	Provider    string    `json:"provider"`
	Homepage    string    `json:"homepage,omitempty"`
	Number      string    `json:"number"`
	Title       string    `json:"title"`
	CoverURL    string    `json:"cover_url"`
	ThumbURL    string    `json:"thumb_url"`
	Score       float64   `json:"score"`
	ReleaseDate time.Time `json:"release_date"`
	Actors      []string  `json:"actors"`
}

// MovieInfo represents full movie metadata.
type MovieInfo struct {
	MovieSearchResult
	Summary             string   `json:"summary"`
	Director            string   `json:"director"`
	Maker               string   `json:"maker"`
	Label               string   `json:"label"`
	Series              string   `json:"series"`
	Runtime             int      `json:"runtime"`
	Genres              []string `json:"genres"`
	BigCoverURL         string   `json:"big_cover_url"`
	BigThumbURL         string   `json:"big_thumb_url"`
	PreviewImages       []string `json:"preview_images"`
	PreviewVideoURL     string   `json:"preview_video_url,omitempty"`
	PreviewVideoHlsURL  string   `json:"preview_video_hls_url,omitempty"`
}

// MarshalJSON customizes release_date serialization to ISO format if needed.
func (m MovieSearchResult) MarshalJSON() ([]byte, error) {
	type Alias MovieSearchResult
	var dateStr string
	if !m.ReleaseDate.IsZero() {
		dateStr = m.ReleaseDate.Format("2006-01-02")
	}
	return json.Marshal(&struct {
		Alias
		ReleaseDate string `json:"release_date,omitempty"`
	}{
		Alias:       Alias(m),
		ReleaseDate: dateStr,
	})
}
