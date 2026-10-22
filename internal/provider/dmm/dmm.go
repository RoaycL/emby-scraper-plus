package dmm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/RoaycL/emby-scraper-plus/internal/model"
	"github.com/RoaycL/emby-scraper-plus/internal/provider"
)

const (
	GraphQLEndpoint = "https://api.video.dmm.co.jp/graphql"
	Name            = "DMM"
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

type graphQLRequest struct {
	Query     string                 `json:"query"`
	Variables map[string]interface{} `json:"variables"`
}

type graphQLResponse struct {
	Data struct {
		PpvContent *contentDetail `json:"ppvContent"`
		LegacySearchPPV *struct {
			Contents []searchItem `json:"contents"`
		} `json:"legacySearchPPV"`
	} `json:"data"`
	Errors []struct {
		Message string `json:"message"`
	} `json:"errors"`
}

type searchItem struct {
	ID              string `json:"id"`
	Title           string `json:"title"`
	MakerContentID  string `json:"makerContentId"`
	MakerReleasedAt string `json:"makerReleasedAt"`
	PackageImage    struct {
		LargeURL  string `json:"largeUrl"`
		MediumURL string `json:"mediumUrl"`
	} `json:"packageImage"`
	Actresses []struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"actresses"`
}

type contentDetail struct {
	ID              string `json:"id"`
	Title           string `json:"title"`
	Description     string `json:"description"`
	MakerContentID  string `json:"makerContentId"`
	MakerReleasedAt string `json:"makerReleasedAt"`
	Duration        int    `json:"duration"`
	Maker           struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"maker"`
	Label struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"label"`
	Series struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"series"`
	Actresses []struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"actresses"`
	Directors []struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"directors"`
	Genres []struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"genres"`
	PackageImage struct {
		LargeURL  string `json:"largeUrl"`
		MediumURL string `json:"mediumUrl"`
	} `json:"packageImage"`
	SampleImages []struct {
		ImageURL string `json:"imageUrl"`
	} `json:"sampleImages"`
}

var codeRegex = regexp.MustCompile(`(?i)^([a-z]+)[-_]?0*(\d+)$`)

// CodeToCid converts a standard code (e.g. SSIS-057) to standard DMM content ID (e.g. ssis00057).
func CodeToCid(code string) string {
	clean := strings.TrimSpace(code)
	if matches := codeRegex.FindStringSubmatch(clean); len(matches) == 3 {
		letters := strings.ToLower(matches[1])
		num, _ := strconv.Atoi(matches[2])
		return fmt.Sprintf("%s%05d", letters, num)
	}
	return strings.ToLower(clean)
}

func (p *Provider) queryGraphQL(ctx context.Context, query string, vars map[string]interface{}) (*graphQLResponse, error) {
	reqBody, err := json.Marshal(graphQLRequest{
		Query:     query,
		Variables: vars,
	})
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, GraphQLEndpoint, bytes.NewReader(reqBody))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", provider.DefaultUserAgent)

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("dmm graphql returned status: %d", resp.StatusCode)
	}

	var gqlResp graphQLResponse
	if err := json.NewDecoder(resp.Body).Decode(&gqlResp); err != nil {
		return nil, err
	}

	return &gqlResp, nil
}

func (p *Provider) GetMovieInfo(ctx context.Context, id string) (*model.MovieInfo, error) {
	cid := CodeToCid(id)

	query := `query Content($id: ID!) {
		ppvContent(id: $id) {
			id
			title
			description
			makerContentId
			makerReleasedAt
			duration
			maker { id name }
			label { id name }
			series { id name }
			actresses { id name }
			directors { id name }
			genres { id name }
			packageImage { largeUrl mediumUrl }
			sampleImages { imageUrl }
		}
	}`

	resp, err := p.queryGraphQL(ctx, query, map[string]interface{}{"id": cid})
	if err != nil || resp.Data.PpvContent == nil {
		// Try search fallback
		searchResult, searchErr := p.SearchMovie(ctx, id)
		if searchErr == nil && len(searchResult) > 0 {
			cid = searchResult[0].ID
			resp, err = p.queryGraphQL(ctx, query, map[string]interface{}{"id": cid})
		}
	}

	if err != nil {
		return nil, err
	}
	if resp.Data.PpvContent == nil {
		return nil, fmt.Errorf("movie not found on dmm: %s", id)
	}

	c := resp.Data.PpvContent
	number := c.MakerContentID
	if number == "" {
		number = strings.ToUpper(id)
	}

	var releaseDate time.Time
	if c.MakerReleasedAt != "" {
		if t, err := time.Parse(time.RFC3339, c.MakerReleasedAt); err == nil {
			// Convert UTC to JST
			releaseDate = t.In(time.FixedZone("JST", 9*3600))
		}
	}

	var actors []string
	for _, a := range c.Actresses {
		if name := strings.TrimSpace(a.Name); name != "" {
			actors = append(actors, name)
		}
	}

	var genres []string
	for _, g := range c.Genres {
		if name := strings.TrimSpace(g.Name); name != "" {
			genres = append(genres, name)
		}
	}

	var director string
	if len(c.Directors) > 0 {
		director = c.Directors[0].Name
	}

	var previewImages []string
	for _, s := range c.SampleImages {
		if s.ImageURL != "" {
			previewImages = append(previewImages, s.ImageURL)
		}
	}

	cover := c.PackageImage.LargeURL
	if cover == "" {
		cover = c.PackageImage.MediumURL
	}

	return &model.MovieInfo{
		MovieSearchResult: model.MovieSearchResult{
			ID:          c.ID,
			Provider:    Name,
			Homepage:    fmt.Sprintf("https://video.dmm.co.jp/av/content/?id=%s", c.ID),
			Number:      number,
			Title:       c.Title,
			CoverURL:    cover,
			ThumbURL:    c.PackageImage.MediumURL,
			ReleaseDate: model.Date(releaseDate),
			Actors:      actors,
		},
		Summary:       c.Description,
		Director:      director,
		Maker:         c.Maker.Name,
		Label:         c.Label.Name,
		Series:        c.Series.Name,
		Runtime:       c.Duration / 60,
		Genres:        genres,
		BigCoverURL:   cover,
		BigThumbURL:   cover,
		PreviewImages: previewImages,
	}, nil
}

func (p *Provider) SearchMovie(ctx context.Context, keyword string) ([]*model.MovieSearchResult, error) {
	// Fast path: if keyword is a movie ID or number, fetch direct info
	if info, err := p.GetMovieInfo(ctx, keyword); err == nil && info != nil && info.Title != "" {
		return []*model.MovieSearchResult{&info.MovieSearchResult}, nil
	}

	query := `query Search($query: String!) {
		legacySearchPPV(query: $query, limit: 10) {
			contents {
				id
				title
				makerContentId
				makerReleasedAt
				packageImage { largeUrl mediumUrl }
				actresses { id name }
			}
		}
	}`

	resp, err := p.queryGraphQL(ctx, query, map[string]interface{}{"query": keyword})
	if err != nil {
		return nil, err
	}

	if resp.Data.LegacySearchPPV == nil {
		return nil, nil
	}

	var results []*model.MovieSearchResult
	for _, item := range resp.Data.LegacySearchPPV.Contents {
		number := item.MakerContentID
		if number == "" {
			number = item.ID
		}

		var releaseDate time.Time
		if item.MakerReleasedAt != "" {
			if t, err := time.Parse(time.RFC3339, item.MakerReleasedAt); err == nil {
				releaseDate = t.In(time.FixedZone("JST", 9*3600))
			}
		}

		var actors []string
		for _, a := range item.Actresses {
			if a.Name != "" {
				actors = append(actors, a.Name)
			}
		}

		cover := item.PackageImage.LargeURL
		if cover == "" {
			cover = item.PackageImage.MediumURL
		}

		results = append(results, &model.MovieSearchResult{
			ID:          item.ID,
			Provider:    Name,
			Homepage:    fmt.Sprintf("https://video.dmm.co.jp/av/content/?id=%s", item.ID),
			Number:      number,
			Title:       item.Title,
			CoverURL:    cover,
			ThumbURL:    item.PackageImage.MediumURL,
			ReleaseDate: model.Date(releaseDate),
			Actors:      actors,
		})
	}

	return results, nil
}

func (p *Provider) SearchActor(ctx context.Context, keyword string) ([]*model.ActorSearchResult, error) {
	return nil, nil
}

func (p *Provider) GetActorInfo(ctx context.Context, id string) (*model.ActorInfo, error) {
	return nil, nil
}
