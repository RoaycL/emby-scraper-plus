package nfo

import (
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/RoaycL/emby-scraper-plus/internal/model"
)

type MovieNfo struct {
	XMLName       xml.Name   `xml:"movie"`
	Title         string     `xml:"title"`
	OriginalTitle string     `xml:"originaltitle,omitempty"`
	SortTitle     string     `xml:"sorttitle,omitempty"`
	Studio        string     `xml:"studio,omitempty"`
	Set           string     `xml:"set,omitempty"`
	Plot          string     `xml:"plot,omitempty"`
	Outline       string     `xml:"outline,omitempty"`
	Premiered     string     `xml:"premiered,omitempty"`
	ReleaseDate   string     `xml:"releasedate,omitempty"`
	Year          string     `xml:"year,omitempty"`
	Director      string     `xml:"director,omitempty"`
	Runtime       int        `xml:"runtime,omitempty"`
	Poster        string     `xml:"art>poster,omitempty"`
	Fanart        string     `xml:"art>fanart,omitempty"`
	Genres        []string   `xml:"genre"`
	Actors        []NfoActor `xml:"actor"`
	UniqueID      []NfoID    `xml:"uniqueid"`
}

type NfoActor struct {
	Name string `xml:"name"`
	Role string `xml:"role,omitempty"`
}

type NfoID struct {
	Type    string `xml:"type,attr"`
	Default bool   `xml:"default,attr"`
	Value   string `xml:",chardata"`
}

// GenerateFromMovieInfo converts a MovieInfo model into standard NFO XML structure.
func GenerateFromMovieInfo(info *model.MovieInfo) *MovieNfo {
	var premiered string
	var year string
	if !info.ReleaseDate.IsZero() {
		premiered = info.ReleaseDate.Format("2006-01-02")
		year = info.ReleaseDate.Format("2006")
	}

	actors := make([]NfoActor, 0, len(info.Actors))
	for _, a := range info.Actors {
		actors = append(actors, NfoActor{Name: a})
	}

	return &MovieNfo{
		Title:         fmt.Sprintf("[%s] %s", info.Number, info.Title),
		OriginalTitle: info.Title,
		SortTitle:     info.Number,
		Studio:        info.Maker,
		Set:           info.Series,
		Plot:          info.Summary,
		Premiered:     premiered,
		ReleaseDate:   premiered,
		Year:          year,
		Director:      info.Director,
		Runtime:       info.Runtime,
		Poster:        info.CoverURL,
		Fanart:        info.BigCoverURL,
		Genres:        info.Genres,
		Actors:        actors,
		UniqueID: []NfoID{
			{Type: "num", Default: true, Value: info.Number},
			{Type: info.Provider, Default: false, Value: info.ID},
		},
	}
}

// WriteNFO serializes the MovieInfo into an .nfo file in the target directory.
func WriteNFO(info *model.MovieInfo, targetDir string, baseFilename string) error {
	nfoData := GenerateFromMovieInfo(info)

	out, err := xml.MarshalIndent(nfoData, "", "  ")
	if err != nil {
		return err
	}

	content := append([]byte(xml.Header), out...)
	destPath := filepath.Join(targetDir, baseFilename+".nfo")

	return os.WriteFile(destPath, content, 0644)
}

// WriteXML outputs the NFO to an io.Writer.
func WriteXML(w io.Writer, info *model.MovieInfo) error {
	nfoData := GenerateFromMovieInfo(info)
	out, err := xml.MarshalIndent(nfoData, "", "  ")
	if err != nil {
		return err
	}
	if _, err := w.Write([]byte(xml.Header)); err != nil {
		return err
	}
	_, err = w.Write(out)
	return err
}
