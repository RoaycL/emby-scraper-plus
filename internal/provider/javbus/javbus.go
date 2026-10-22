package javbus

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/RoaycL/emby-scraper-plus/internal/model"
	"github.com/RoaycL/emby-scraper-plus/internal/provider"
)

const (
	BaseURL = "https://www.javbus.com"
	Name    = "JAVBUS"
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

func (p *Provider) fetchDoc(ctx context.Context, targetURL string) (*goquery.Document, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, targetURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", provider.DefaultUserAgent)
	// Critical fix from upstream: dv=1 and existmag=all bypass JavBus age confirmation gate
	req.Header.Set("Cookie", "existmag=all; dv=1; age=verified")

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("javbus: 404 not found")
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("javbus returned status: %d", resp.StatusCode)
	}

	return goquery.NewDocumentFromReader(resp.Body)
}

func (p *Provider) GetMovieInfo(ctx context.Context, id string) (*model.MovieInfo, error) {
	cleanID := strings.TrimSpace(id)
	targetURL := fmt.Sprintf("%s/%s", BaseURL, url.PathEscape(cleanID))

	doc, err := p.fetchDoc(ctx, targetURL)
	if err != nil {
		// Fallback to search if direct code URL fails
		results, sErr := p.SearchMovie(ctx, cleanID)
		if sErr == nil && len(results) > 0 {
			targetURL = fmt.Sprintf("%s/%s", BaseURL, url.PathEscape(results[0].ID))
			doc, err = p.fetchDoc(ctx, targetURL)
		}
	}

	if err != nil {
		return nil, err
	}

	title := strings.TrimSpace(doc.Find("h3").First().Text())
	if title == "" {
		return nil, fmt.Errorf("javbus: empty title or page blocked for %s", id)
	}

	cover, _ := doc.Find("a.bigImage").Attr("href")
	if cover != "" && !strings.HasPrefix(cover, "http") {
		cover = BaseURL + cover
	}

	info := &model.MovieInfo{
		MovieSearchResult: model.MovieSearchResult{
			ID:          cleanID,
			Provider:    Name,
			Homepage:    targetURL,
			Number:      cleanID,
			Title:       title,
			CoverURL:    cover,
			ThumbURL:    cover,
		},
		BigCoverURL: cover,
		BigThumbURL: cover,
	}

	// Parse sidebar fields: <span class="header">識別碼:</span>
	doc.Find("div.info p").Each(func(i int, s *goquery.Selection) {
		text := s.Text()
		header := strings.TrimSpace(s.Find("span.header").Text())

		switch {
		case strings.Contains(header, "識別碼") || strings.Contains(header, "识别码"):
			val := strings.TrimSpace(s.Find("span").Last().Text())
			if val != "" {
				info.Number = val
				info.ID = val
			}
		case strings.Contains(header, "發行日期") || strings.Contains(header, "发行日期"):
			dateStr := strings.TrimSpace(strings.ReplaceAll(text, header, ""))
			if t, err := time.Parse("2006-01-02", dateStr); err == nil {
				info.ReleaseDate = t
			}
		case strings.Contains(header, "長度") || strings.Contains(header, "长度"):
			var mins int
			_, _ = fmt.Sscanf(text, "%d分鐘", &mins)
			if mins == 0 {
				_, _ = fmt.Sscanf(text, "%d分钟", &mins)
			}
			info.Runtime = mins
		case strings.Contains(header, "導演") || strings.Contains(header, "导演"):
			info.Director = strings.TrimSpace(s.Find("a").Text())
		case strings.Contains(header, "製作商") || strings.Contains(header, "制作商"):
			info.Maker = strings.TrimSpace(s.Find("a").Text())
		case strings.Contains(header, "發行商") || strings.Contains(header, "发行商"):
			info.Label = strings.TrimSpace(s.Find("a").Text())
		case strings.Contains(header, "系列"):
			info.Series = strings.TrimSpace(s.Find("a").Text())
		}
	})

	// Parse genres
	var genres []string
	doc.Find("span.genre a").Each(func(i int, s *goquery.Selection) {
		g := strings.TrimSpace(s.Text())
		if g != "" {
			genres = append(genres, g)
		}
	})
	info.Genres = genres

	// Parse actresses
	var actors []string
	doc.Find("div.star-name a").Each(func(i int, s *goquery.Selection) {
		name := strings.TrimSpace(s.Text())
		if name != "" {
			actors = append(actors, name)
		}
	})
	info.Actors = actors

	// Parse sample images
	var previewImages []string
	doc.Find("div#sample-waterfall a.sample-box").Each(func(i int, s *goquery.Selection) {
		if href, ok := s.Attr("href"); ok && href != "" {
			if !strings.HasPrefix(href, "http") {
				href = BaseURL + href
			}
			previewImages = append(previewImages, href)
		}
	})
	info.PreviewImages = previewImages

	return info, nil
}

func (p *Provider) SearchMovie(ctx context.Context, keyword string) ([]*model.MovieSearchResult, error) {
	searchURL := fmt.Sprintf("%s/search/%s", BaseURL, url.PathEscape(keyword))
	doc, err := p.fetchDoc(ctx, searchURL)
	if err != nil {
		return nil, err
	}

	var results []*model.MovieSearchResult
	doc.Find("a.movie-box").Each(func(i int, s *goquery.Selection) {
		href, _ := s.Attr("href")
		img, _ := s.Find("img").Attr("src")
		title, _ := s.Find("img").Attr("title")
		num := strings.TrimSpace(s.Find("date").First().Text())
		dateStr := strings.TrimSpace(s.Find("date").Last().Text())

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

		if !strings.HasPrefix(img, "http") {
			img = BaseURL + img
		}

		results = append(results, &model.MovieSearchResult{
			ID:          id,
			Provider:    Name,
			Homepage:    href,
			Number:      num,
			Title:       title,
			CoverURL:    img,
			ThumbURL:    img,
			ReleaseDate: relDate,
		})
	})

	return results, nil
}

func (p *Provider) SearchActor(ctx context.Context, keyword string) ([]*model.ActorSearchResult, error) {
	return nil, nil
}

func (p *Provider) GetActorInfo(ctx context.Context, id string) (*model.ActorInfo, error) {
	return nil, nil
}
