package fc2

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/RoaycL/emby-scraper-plus/internal/model"
	"github.com/RoaycL/emby-scraper-plus/internal/provider"
)

const (
	BaseURL  = "https://adult.contents.fc2.com"
	CmaBase  = "https://fc2cmadb.com"
	Name     = "FC2"
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

var fc2NumRegex = regexp.MustCompile(`(?i)(?:FC2[-_]?(?:PPV[-_]?)?|FC-)?(\d+)`)

func CleanFC2ID(id string) string {
	matches := fc2NumRegex.FindStringSubmatch(strings.TrimSpace(id))
	if len(matches) == 2 {
		return matches[1]
	}
	return id
}

func (p *Provider) GetMovieInfo(ctx context.Context, id string) (*model.MovieInfo, error) {
	num := CleanFC2ID(id)
	stdNumber := fmt.Sprintf("FC2-PPV-%s", num)

	// Step 1: Try official FC2 product page
	officialURL := fmt.Sprintf("%s/article/%s/", BaseURL, num)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, officialURL, nil)
	if err == nil {
		req.Header.Set("User-Agent", provider.DefaultUserAgent)
		req.Header.Set("Cookie", "age_confirmed=1")

		if resp, err := p.client.Do(req); err == nil && resp.StatusCode == http.StatusOK {
			doc, err := goquery.NewDocumentFromReader(resp.Body)
			resp.Body.Close()
			if err == nil {
				title := strings.TrimSpace(doc.Find("h3").First().Text())
				if title != "" {
					cover, _ := doc.Find("div.items_article_MainitemThumb img").Attr("src")
					if cover != "" && !strings.HasPrefix(cover, "http") {
						cover = "https:" + cover
					}

					var genres []string
					doc.Find("section.items_article_TagArea a.tag").Each(func(i int, s *goquery.Selection) {
						g := strings.TrimSpace(s.Text())
						if g != "" {
							genres = append(genres, g)
						}
					})

					seller := strings.TrimSpace(doc.Find("div.items_article_headerInfo ul li a").First().Text())

					return &model.MovieInfo{
						MovieSearchResult: model.MovieSearchResult{
							ID:          stdNumber,
							Provider:    Name,
							Homepage:    officialURL,
							Number:      stdNumber,
							Title:       title,
							CoverURL:    cover,
							ThumbURL:    cover,
						},
						BigCoverURL: cover,
						BigThumbURL: cover,
						Maker:       seller,
						Genres:      genres,
					}, nil
				}
			}
		} else if resp != nil {
			resp.Body.Close()
		}
	}

	// Step 2: Fallback to FC2CMADB (mirror for deleted or unlisted FC2 items)
	cmaURL := fmt.Sprintf("%s/articles/%s", CmaBase, num)
	cReq, err := http.NewRequestWithContext(ctx, http.MethodGet, cmaURL, nil)
	if err != nil {
		return nil, err
	}
	cReq.Header.Set("User-Agent", provider.DefaultUserAgent)

	cResp, err := p.client.Do(cReq)
	if err != nil {
		return nil, fmt.Errorf("fc2: movie not found: %s", id)
	}
	defer cResp.Body.Close()

	if cResp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("fc2: movie not found on official or cmadb: %s", id)
	}

	doc, err := goquery.NewDocumentFromReader(cResp.Body)
	if err != nil {
		return nil, err
	}

	appJSON := doc.Find("div#app").AttrOr("data-page", "")
	if appJSON == "" {
		return nil, fmt.Errorf("fc2: unable to parse cmadb page")
	}

	var pageData struct {
		Props struct {
			Article struct {
				Title       string   `json:"title"`
				Description string   `json:"description"`
				SellerName  string   `json:"seller_name"`
				ActressName string   `json:"actress_name"`
				ReleaseDate string   `json:"release_date"`
				MainImage   string   `json:"main_image"`
				Tags        []string `json:"tags"`
			} `json:"article"`
		} `json:"props"`
	}

	if err := json.Unmarshal([]byte(appJSON), &pageData); err != nil {
		return nil, err
	}

	art := pageData.Props.Article
	var releaseDate time.Time
	if art.ReleaseDate != "" {
		if t, err := time.Parse("2006-01-02", art.ReleaseDate); err == nil {
			releaseDate = t
		}
	}

	var actors []string
	if art.ActressName != "" {
		actors = append(actors, art.ActressName)
	}

	return &model.MovieInfo{
		MovieSearchResult: model.MovieSearchResult{
			ID:          stdNumber,
			Provider:    Name,
			Homepage:    cmaURL,
			Number:      stdNumber,
			Title:       art.Title,
			CoverURL:    art.MainImage,
			ThumbURL:    art.MainImage,
			ReleaseDate: releaseDate,
			Actors:      actors,
		},
		BigCoverURL: art.MainImage,
		BigThumbURL: art.MainImage,
		Summary:     art.Description,
		Maker:       art.SellerName,
		Genres:      art.Tags,
	}, nil
}

func (p *Provider) SearchMovie(ctx context.Context, keyword string) ([]*model.MovieSearchResult, error) {
	num := CleanFC2ID(keyword)
	info, err := p.GetMovieInfo(ctx, num)
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
