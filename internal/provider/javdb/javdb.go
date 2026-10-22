package javdb

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/RoaycL/emby-scraper-plus/internal/model"
	"github.com/RoaycL/emby-scraper-plus/internal/provider"
)

const (
	BaseURL = "https://javdb.com"
	Name    = "JAVDB"
)

type Provider struct {
	client   *http.Client
	mu       sync.Mutex
	lastCall time.Time
	interval time.Duration
}

func New() *Provider {
	return &Provider{
		client:   provider.NewHTTPClient(15 * time.Second),
		interval: 3 * time.Second, // Peach rate limiter standard to avoid IP ban
	}
}

func (p *Provider) Name() string {
	return Name
}

func (p *Provider) waitRateLimit() {
	p.mu.Lock()
	defer p.mu.Unlock()

	elapsed := time.Since(p.lastCall)
	if elapsed < p.interval {
		time.Sleep(p.interval - elapsed)
	}
	p.lastCall = time.Now()
}

func (p *Provider) fetchDoc(ctx context.Context, targetURL string) (*goquery.Document, error) {
	p.waitRateLimit()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, targetURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", provider.DefaultUserAgent)
	req.Header.Set("Cookie", "over18=1; locale=zh")

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("javdb: 404 not found")
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("javdb returned status: %d", resp.StatusCode)
	}

	return goquery.NewDocumentFromReader(resp.Body)
}

func (p *Provider) SearchMovie(ctx context.Context, keyword string) ([]*model.MovieSearchResult, error) {
	searchURL := fmt.Sprintf("%s/search?q=%s&f=all", BaseURL, url.QueryEscape(keyword))
	doc, err := p.fetchDoc(ctx, searchURL)
	if err != nil {
		return nil, err
	}

	var results []*model.MovieSearchResult
	doc.Find("div.movie-list div.item").Each(func(i int, s *goquery.Selection) {
		box := s.Find("a.box")
		href, _ := box.Attr("href")
		img, _ := s.Find("img").Attr("src")
		num := strings.TrimSpace(s.Find("div.video-title strong").Text())
		title := strings.TrimSpace(s.Find("div.video-title").Text())
		dateStr := strings.TrimSpace(s.Find("div.meta").Text())

		var relDate time.Time
		if t, err := time.Parse("2006-01-02", dateStr); err == nil {
			relDate = t
		}

		id := ""
		if href != "" {
			parts := strings.Split(strings.TrimRight(href, "/"), "/")
			id = parts[len(parts)-1]
		}
		if num == "" {
			num = id
		}

		results = append(results, &model.MovieSearchResult{
			ID:          id,
			Provider:    Name,
			Homepage:    BaseURL + href,
			Number:      num,
			Title:       title,
			CoverURL:    img,
			ThumbURL:    img,
			ReleaseDate: model.Date(relDate),
		})
	})

	return results, nil
}

func (p *Provider) GetMovieInfo(ctx context.Context, id string) (*model.MovieInfo, error) {
	detailURL := ""
	if strings.HasPrefix(id, "v/") || strings.HasPrefix(id, "/") {
		detailURL = BaseURL + "/" + strings.TrimLeft(id, "/") + "?locale=zh"
	} else {
		// First search to get the item detail path
		results, err := p.SearchMovie(ctx, id)
		if err != nil || len(results) == 0 {
			return nil, fmt.Errorf("javdb: movie not found: %s", id)
		}
		detailURL = fmt.Sprintf("%s/v/%s?locale=zh", BaseURL, results[0].ID)
	}

	doc, err := p.fetchDoc(ctx, detailURL)
	if err != nil {
		return nil, err
	}

	title := strings.TrimSpace(doc.Find("strong.current-title").Text())
	cover, _ := doc.Find("img.video-cover").Attr("src")

	info := &model.MovieInfo{
		MovieSearchResult: model.MovieSearchResult{
			ID:          id,
			Provider:    Name,
			Homepage:    detailURL,
			Number:      id,
			Title:       title,
			CoverURL:    cover,
			ThumbURL:    cover,
		},
		BigCoverURL: cover,
		BigThumbURL: cover,
	}

	doc.Find("div.panel-block").Each(func(i int, s *goquery.Selection) {
		label := strings.TrimSpace(s.Find("strong").Text())
		val := strings.TrimSpace(s.Find("span.value").Text())

		switch {
		case strings.Contains(label, "番號") || strings.Contains(label, "番号"):
			info.Number = val
		case strings.Contains(label, "日期"):
			if t, err := time.Parse("2006-01-02", val); err == nil {
				info.ReleaseDate = model.Date(t)
			}
		case strings.Contains(label, "時長") || strings.Contains(label, "时长"):
			var mins int
			_, _ = fmt.Sscanf(val, "%d 分鐘", &mins)
			if mins == 0 {
				_, _ = fmt.Sscanf(val, "%d 分钟", &mins)
			}
			info.Runtime = mins
		case strings.Contains(label, "導演") || strings.Contains(label, "导演"):
			info.Director = s.Find("span.value a").Text()
		case strings.Contains(label, "片商"):
			info.Maker = s.Find("span.value a").Text()
		case strings.Contains(label, "發行") || strings.Contains(label, "发行"):
			info.Label = s.Find("span.value a").Text()
		case strings.Contains(label, "系列"):
			info.Series = s.Find("span.value a").Text()
		case strings.Contains(label, "演員") || strings.Contains(label, "演员"):
			var actors []string
			s.Find("span.value a").Each(func(j int, a *goquery.Selection) {
				name := strings.TrimSpace(a.Text())
				if name != "" {
					actors = append(actors, name)
				}
			})
			info.Actors = actors
		case strings.Contains(label, "類別") || strings.Contains(label, "类别"):
			var genres []string
			s.Find("span.value a").Each(func(j int, g *goquery.Selection) {
				genre := strings.TrimSpace(g.Text())
				if genre != "" {
					genres = append(genres, genre)
				}
			})
			info.Genres = genres
		}
	})

	return info, nil
}

func (p *Provider) SearchActor(ctx context.Context, keyword string) ([]*model.ActorSearchResult, error) {
	return nil, nil
}

func (p *Provider) GetActorInfo(ctx context.Context, id string) (*model.ActorInfo, error) {
	return nil, nil
}
