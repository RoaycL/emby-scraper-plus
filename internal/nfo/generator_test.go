package nfo

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/RoaycL/emby-scraper-plus/internal/model"
)

func TestWriteXML(t *testing.T) {
	relDate, _ := time.Parse("2006-01-02", "2026-05-20")
	info := &model.MovieInfo{
		MovieSearchResult: model.MovieSearchResult{
			ID:          "SSIS-057",
			Number:      "SSIS-057",
			Title:       "Test Movie Title",
			Provider:    "DMM",
			CoverURL:    "https://example.com/cover.jpg",
			ReleaseDate: model.Date(relDate),
			Actors:      []string{"Actor A", "Actor B"},
		},
		Summary:     "Test overview plot summary.",
		Director:    "Test Director",
		Maker:       "S1 NO.1 STYLE",
		BigCoverURL: "https://example.com/big_cover.jpg",
		Genres:      []string{"Drama", "Feature"},
	}

	var buf bytes.Buffer
	err := WriteXML(&buf, info)
	if err != nil {
		t.Fatalf("WriteXML failed: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "<title>[SSIS-057] Test Movie Title</title>") {
		t.Errorf("expected title in NFO XML, got: %s", output)
	}
	if !strings.Contains(output, "<name>Actor A</name>") {
		t.Errorf("expected Actor A in NFO XML, got: %s", output)
	}
	if !strings.Contains(output, "<studio>S1 NO.1 STYLE</studio>") {
		t.Errorf("expected studio in NFO XML, got: %s", output)
	}
}
