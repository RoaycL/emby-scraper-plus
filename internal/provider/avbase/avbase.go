package avbase

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/RoaycL/emby-scraper-plus/internal/model"
	"github.com/RoaycL/emby-scraper-plus/internal/provider"
)

const (
	BaseURL = "https://www.avbase.net"
	Name    = "AVBASE"
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

type nextDataWrapper struct {
	Props struct {
		PageProps struct {
			Works []struct {
				WorkID    string   `json:"work_id"`
				Title     string   `json:"title"`
				Actors    []string `json:"actors"`
				WorkImage string   `json:"work_image"`
				Products  []struct {
					Source   string `json:"source"`
					URL      string `json:"url"`
					ImageURL string `json:"image_url"`
					Maker    string `json:"maker"`
					Label    string `json:"label"`
					Series   string `json:"series"`
					Date     string `json:"date"`
					ItemInfo struct {
						Director string `json:"director"`
					} `json:"iteminfo"`
				} `json:"products"`
			} `json:"works"`
		} `json:"pageProps"`
	} `json:"props"`
}

func (p *Provider) SearchMovie(ctx context.Context, keyword string) ([]*model.MovieSearchResult, error) {
	reqURL := fmt.Sprintf("%s/works?q=%s", BaseURL, url.QueryEscape(keyword))
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

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("avbase returned status: %d", resp.StatusCode)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, err
	}

	nextDataScript := doc.Find("script#__NEXT_DATA__").Text()
	if nextDataScript == "" {
		return nil, fmt.Errorf("avbase: __NEXT_DATA__ script not found")
	}

	var data nextDataWrapper
	if err := json.Unmarshal([]byte(nextDataScript), &data); err != nil {
		return nil, err
	}

	var results []*model.MovieSearchResult
	for _, w := range data.Props.PageProps.Works {
		var cover string
		var releaseDate time.Time

		if len(w.Products) > 0 {
			prod := w.Products[0]
			cover = prod.ImageURL
			if prod.Date != "" {
				// Format: Fri May 26 2023 09:00:00 GMT+0900
				if t, err := time.Parse("Mon Jan 02 2006 15:04:05 MST-0700", prod.Date); err == nil {
					releaseDate = t
				} else if t, err := time.Parse("2006-01-02", prod.Date); err == nil {
					releaseDate = t
				}
			}
		}
		if cover == "" {
			cover = w.WorkImage
		}

		results = append(results, &model.MovieSearchResult{
			ID:          w.WorkID,
			Provider:    Name,
			Homepage:    fmt.Sprintf("%s/works/%s", BaseURL, w.WorkID),
			Number:      w.WorkID,
			Title:       w.Title,
			CoverURL:    cover,
			ThumbURL:    cover,
			ReleaseDate: model.Date(releaseDate),
			Actors:      w.Actors,
		})
	}

	return results, nil
}

func (p *Provider) GetMovieInfo(ctx context.Context, id string) (*model.MovieInfo, error) {
	// AVBase details can be directly obtained via searching work_id
	results, err := p.SearchMovie(ctx, id)
	if err != nil {
		return nil, err
	}

	if len(results) == 0 {
		return nil, fmt.Errorf("avbase: movie not found: %s", id)
	}

	item := results[0]
	return &model.MovieInfo{
		MovieSearchResult: *item,
		BigCoverURL:       item.CoverURL,
		BigThumbURL:       item.CoverURL,
	}, nil
}

func (p *Provider) SearchActor(ctx context.Context, keyword string) ([]*model.ActorSearchResult, error) {
	return nil, nil
}

func (p *Provider) GetActorInfo(ctx context.Context, id string) (*model.ActorInfo, error) {
	return nil, nil
}
