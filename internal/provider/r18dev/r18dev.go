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
	ContentID      string `json:"content_id"`
	DVDID          string `json:"dvd_id"`
	Title          string `json:"title"`
	TitleJa        string `json:"title_ja"`
	TitleEn        string `json:"title_en"`
	ReleaseDate    string `json:"release_date"`
	RuntimeMins    int    `json:"runtime_mins"`
	RuntimeMinutes int    `json:"runtime_minutes"`
	Maker          struct {
		Name string `json:"name"`
	} `json:"maker"`
	MakerNameJa string `json:"maker_name_ja"`
	MakerNameEn string `json:"maker_name_en"`
	Label       struct {
		Name string `json:"name"`
	} `json:"label"`
	LabelNameJa string `json:"label_name_ja"`
	LabelNameEn string `json:"label_name_en"`
	Series      struct {
		Name string `json:"name"`
	} `json:"series"`
	SeriesNameJa string `json:"series_name_ja"`
	Directors    []struct {
		Name   string `json:"name"`
		NameJa string `json:"name_ja"`
		NameEn string `json:"name_en"`
	} `json:"directors"`
	Director string `json:"director"`
	Actresses []struct {
		Name      string `json:"name"`
		NameKanji string `json:"name_kanji"`
		ImageURL  string `json:"image_url"`
	} `json:"actresses"`
	Categories []struct {
		Name   string `json:"name"`
		NameJa string `json:"name_ja"`
		NameEn string `json:"name_en"`
	} `json:"categories"`
	Images struct {
		JacketImage struct {
			Large  string `json:"large"`
			Large2 string `json:"large2"`
		} `json:"jacket_image"`
		Jackets struct {
			Large  string `json:"large"`
			Medium string `json:"medium"`
			Small  string `json:"small"`
		} `json:"jackets"`
		Gallery []struct {
			Large string `json:"large"`
		} `json:"gallery"`
	} `json:"images"`
	Gallery []struct {
		Large string `json:"large"`
	} `json:"gallery"`
	JacketFullURL  string `json:"jacket_full_url"`
	JacketThumbURL string `json:"jacket_thumb_url"`
	Translations   struct {
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
	if d.ContentID != "" && (d.Translations.Ja.Title == "" || d.TitleJa == "" || d.JacketFullURL == "") {
		combinedURL := fmt.Sprintf("%s/videos/vod/movies/detail/-/combined=%s/json", BaseURL, url.PathEscape(d.ContentID))
		if cReq, err := http.NewRequestWithContext(ctx, http.MethodGet, combinedURL, nil); err == nil {
			cReq.Header.Set("User-Agent", provider.DefaultUserAgent)
			if cResp, err := p.client.Do(cReq); err == nil && cResp.StatusCode == http.StatusOK {
				var comb r18DetailResponse
				if err := json.NewDecoder(cResp.Body).Decode(&comb); err == nil {
					if comb.TitleJa != "" {
						d.TitleJa = comb.TitleJa
					}
					if comb.TitleEn != "" {
						d.TitleEn = comb.TitleEn
					}
					if comb.DVDID != "" {
						d.DVDID = comb.DVDID
					}
					if comb.JacketFullURL != "" {
						d.JacketFullURL = comb.JacketFullURL
					}
					if comb.JacketThumbURL != "" {
						d.JacketThumbURL = comb.JacketThumbURL
					}
					if comb.MakerNameJa != "" {
						d.MakerNameJa = comb.MakerNameJa
					}
					if comb.LabelNameJa != "" {
						d.LabelNameJa = comb.LabelNameJa
					}
					if comb.SeriesNameJa != "" {
						d.SeriesNameJa = comb.SeriesNameJa
					}
					if comb.RuntimeMins > 0 {
						d.RuntimeMins = comb.RuntimeMins
					}
					if len(comb.Actresses) > 0 {
						d.Actresses = comb.Actresses
					}
					if len(comb.Categories) > 0 {
						d.Categories = comb.Categories
					}
					if len(comb.Gallery) > 0 {
						d.Gallery = comb.Gallery
					}
					if len(comb.Directors) > 0 {
						d.Directors = comb.Directors
					}
				}
				cResp.Body.Close()
			}
		}
	}

	title := d.TitleJa
	if title == "" {
		title = d.Translations.Ja.Title
	}
	if title == "" {
		title = d.Title
	}
	if title == "" {
		title = d.TitleEn
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
		if name == "" {
			name = c.NameEn
		}
		if name != "" {
			genres = append(genres, strings.TrimSpace(name))
		}
	}

	var director string
	if d.Director != "" {
		director = d.Director
	} else if len(d.Directors) > 0 {
		director = d.Directors[0].NameJa
		if director == "" {
			director = d.Directors[0].Name
		}
		if director == "" {
			director = d.Directors[0].NameEn
		}
	}

	cover := d.JacketFullURL
	if cover == "" {
		cover = d.Images.JacketImage.Large2
	}
	if cover == "" || strings.TrimSpace(cover) == "" {
		cover = d.Images.Jackets.Large
	}
	if cover == "" || strings.TrimSpace(cover) == "" {
		cover = d.Images.Jackets.Medium
	}
	cover = strings.TrimSpace(cover)

	thumb := d.JacketThumbURL
	if thumb == "" {
		thumb = d.Images.Jackets.Small
	}
	if thumb == "" {
		thumb = cover
	}
	thumb = strings.TrimSpace(thumb)

	var previewImages []string
	for _, g := range d.Gallery {
		if g.Large != "" {
			previewImages = append(previewImages, g.Large)
		}
	}
	if len(previewImages) == 0 {
		for _, g := range d.Images.Gallery {
			if g.Large != "" {
				previewImages = append(previewImages, g.Large)
			}
		}
	}

	maker := d.MakerNameJa
	if maker == "" {
		maker = d.Maker.Name
	}
	if maker == "" {
		maker = d.MakerNameEn
	}

	label := d.LabelNameJa
	if label == "" {
		label = d.Label.Name
	}
	if label == "" {
		label = d.LabelNameEn
	}

	series := d.SeriesNameJa
	if series == "" {
		series = d.Translations.Ja.SeriesName
	}
	if series == "" {
		series = d.Series.Name
	}

	runtime := d.RuntimeMins
	if runtime == 0 {
		runtime = d.RuntimeMinutes
	}

	dvdID := d.DVDID
	if dvdID == "" {
		dvdID = code
	}

	return &model.MovieInfo{
		MovieSearchResult: model.MovieSearchResult{
			ID:          dvdID,
			Provider:    Name,
			Homepage:    fmt.Sprintf("https://www.dmm.co.jp/digital/videoa/-/detail/=/cid=%s/", d.ContentID),
			Number:      dvdID,
			Title:       title,
			CoverURL:    cover,
			ThumbURL:    thumb,
			ReleaseDate: model.Date(releaseDate),
			Actors:      actors,
		},
		Director:      director,
		Maker:         maker,
		Label:         label,
		Series:        series,
		Runtime:       runtime,
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
