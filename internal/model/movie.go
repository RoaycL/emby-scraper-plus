package model

import (
	"fmt"
	"strings"
	"time"
)

// Date represents a date serialized as YYYY-MM-DD or null in JSON.
type Date time.Time

func (d Date) MarshalJSON() ([]byte, error) {
	t := time.Time(d)
	if t.IsZero() {
		return []byte("null"), nil
	}
	return []byte(fmt.Sprintf("%q", t.Format("2006-01-02"))), nil
}

func (d *Date) UnmarshalJSON(b []byte) error {
	s := strings.Trim(string(b), "\"")
	if s == "null" || s == "" {
		return nil
	}
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		return err
	}
	*d = Date(t)
	return nil
}

func (d Date) Time() time.Time {
	return time.Time(d)
}

func (d Date) IsZero() bool {
	return time.Time(d).IsZero()
}

func (d Date) Format(layout string) string {
	return time.Time(d).Format(layout)
}

// MovieSearchResult represents the search result item returned to MetaTube plugin/client.
type MovieSearchResult struct {
	ID          string   `json:"id"`
	Provider    string   `json:"provider"`
	Homepage    string   `json:"homepage,omitempty"`
	Number      string   `json:"number"`
	Title       string   `json:"title"`
	CoverURL    string   `json:"cover_url"`
	ThumbURL    string   `json:"thumb_url"`
	Score       float64  `json:"score"`
	ReleaseDate Date     `json:"release_date"`
	Actors      []string `json:"actors"`
}

// MovieInfo represents full movie metadata.
type MovieInfo struct {
	MovieSearchResult
	Summary            string   `json:"summary"`
	Director           string   `json:"director"`
	Maker              string   `json:"maker"`
	Label              string   `json:"label"`
	Series             string   `json:"series"`
	Runtime            int      `json:"runtime"`
	Genres             []string `json:"genres"`
	BigCoverURL        string   `json:"big_cover_url"`
	BigThumbURL        string   `json:"big_thumb_url"`
	PreviewImages      []string `json:"preview_images"`
	PreviewVideoURL    string   `json:"preview_video_url,omitempty"`
	PreviewVideoHlsURL string   `json:"preview_video_hls_url,omitempty"`
}
