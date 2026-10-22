package r18dev

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/RoaycL/emby-scraper-plus/internal/model"
	"github.com/RoaycL/emby-scraper-plus/internal/provider"
)

const (
	BaseURL = "https://r18.dev"
	Name    = "R18DEV"
)

type Provider struct {
	client *http.Client
}

func New() *Provider {
	return &Provider{
		client: provider.NewHTTPClient(15 * time.Second),
	}
}

func (p *Provider) Name() string {
	return Name
}

type r18DetailResponse struct {
	ContentID   string `json:"content_id"`
	DVDID       string `json:"dvd_id"`
	Title       string `json:"title"`
	ReleaseDate string `json:"release_date"`
	Runtime     int    `json:"runtime_mins"`
	Maker       struct {
		Name string `json:"name"`
	} `json:"maker"`
	Label struct {
		Name string `json:"name"`
	} `json:"label"`
	Series struct {
		Name string `json:"name"`
	} `json:"series"`
	Directors []struct {
		Name string `json:"name"`
	} `json:"directors"`
	Actresses []struct {
		Name      string `json:"name"`
		NameKanji string `json:"name_kanji"`
		ImageURL  string `json:"image_url"`
	} `json:"actresses"`
	Categories []struct {
		Name   string `json:"name"`
		NameJa string `json:"name_ja"`
	} `json:"categories"`
	Images struct {
		Jackets struct {
			Large  string `json:"large"`
			Medium string `json:"medium"`
			Small  string `json:"small"`
		} `json:"jackets"`
		Gallery []struct {
			Large string `json:"large"`
		} `json:"gallery"`
	} `json:"images"`
	Translations struct {
		Ja struct {
			Title      string `json:"title"`
			SeriesName string `json:"series_name"`
			LabelName  string `json:"label_name"`
		} `json:"ja"`
	} `json:"translations"`
}

func (p *Provider) GetMovieInfo(ctx context.Context, id string) (*model.MovieInfo, error) {
	code := strings.TrimSpace(id)
	// Try fetching by dvd_id first
	reqURL := fmt.Sprintf("%s/videos/vod/movies/detail/-/dvd_id=%s/json", BaseURL, url.PathEscape(code))
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", provider.DefaultUserAgent)

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("r18dev: movie not found: %s", id)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("r18dev returned status: %d", resp.StatusCode)
	}

	var d r18DetailResponse
	if err := json.NewDecoder(resp.Body).Decode(&d); err != nil {
		return nil, err
	}

	// Try fetching combined version for Japanese Kanji text if needed
	if d.ContentID != "" && d.Translations.Ja.Title == "" {
		combinedURL := fmt.Sprintf("%s/videos/vod/movies/detail/-/combined=%s/json", BaseURL, url.PathEscape(d.ContentID))
		if cReq, err := http.NewRequestWithContext(ctx, http.MethodGet, combinedURL, nil); err == nil {
			cReq.Header.Set("User-Agent", provider.DefaultUserAgent)
			if cResp, err := p.client.Do(cReq); err == nil && cResp.StatusCode == http.StatusOK {
				var comb r18DetailResponse
				if err := json.NewDecoder(cResp.Body).Decode(&comb); err == nil {
					d = comb
				}
				cResp.Body.Close()
			}
		}
	}

	title := d.Title
	if d.Translations.Ja.Title != "" {
		title = d.Translations.Ja.Title
	}

	var releaseDate time.Time
	if d.ReleaseDate != "" {
		if t, err := time.Parse("2006-01-02", d.ReleaseDate); err == nil {
			releaseDate = t
		}
	}

	var actors []string
	for _, a := range d.Actresses {
		name := a.NameKanji
		if name == "" {
			name = a.Name
		}
		if name != "" {
			actors = append(actors, strings.TrimSpace(name))
		}
	}

	var genres []string
	for _, c := range d.Categories {
		name := c.NameJa
		if name == "" {
			name = c.Name
		}
		if name != "" {
			genres = append(genres, strings.TrimSpace(name))
		}
	}

	var director string
	if len(d.Directors) > 0 {
		director = d.Directors[0].Name
	}

	cover := d.Images.Jackets.Large
	if cover == "" {
		cover = d.Images.Jackets.Medium
	}

	var previewImages []string
	for _, g := range d.Images.Gallery {
		if g.Large != "" {
			previewImages = append(previewImages, g.Large)
		}
	}

	series := d.Series.Name
	if d.Translations.Ja.SeriesName != "" {
		series = d.Translations.Ja.SeriesName
	}

	return &model.MovieInfo{
		MovieSearchResult: model.MovieSearchResult{
			ID:          d.DVDID,
			Provider:    Name,
			Homepage:    fmt.Sprintf("https://www.dmm.co.jp/digital/videoa/-/detail/=/cid=%s/", d.ContentID),
			Number:      d.DVDID,
			Title:       title,
			CoverURL:    cover,
			ThumbURL:    d.Images.Jackets.Small,
			ReleaseDate: releaseDate,
			Actors:      actors,
		},
		Director:      director,
		Maker:         d.Maker.Name,
		Label:         d.Label.Name,
		Series:        series,
		Runtime:       d.Runtime,
		Genres:        genres,
		BigCoverURL:   cover,
		BigThumbURL:   cover,
		PreviewImages: previewImages,
	}, nil
}

func (p *Provider) SearchMovie(ctx context.Context, keyword string) ([]*model.MovieSearchResult, error) {
	// Directly try GetMovieInfo since r18.dev indexes by exact DVD ID
	info, err := p.GetMovieInfo(ctx, keyword)
	if err != nil {
		return nil, err
	}
	return []*model.MovieSearchResult{&info.MovieSearchResult}, nil
}

func (p *Provider) SearchActor(ctx context.Context, keyword string) ([]*model.ActorSearchResult, error) {
	return nil, nil
}

func (p *Provider) GetActorInfo(ctx context.Context, id string) (*model.ActorInfo, error) {
	return nil, nil
}
