package gfriends

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/RoaycL/emby-scraper-plus/internal/model"
	"github.com/RoaycL/emby-scraper-plus/internal/provider"
)

const (
	FiletreeURL = "https://raw.githubusercontent.com/gfriends/gfriends/master/Filetree.json"
	ContentBase = "https://raw.githubusercontent.com/gfriends/gfriends/master/Content"
	Name        = "GFRIENDS"
)

type Provider struct {
	client    *http.Client
	filetree  map[string]string // Name -> Relative Path
	mu        sync.RWMutex
	loadedAt  time.Time
}

func New() *Provider {
	p := &Provider{
		client:   provider.NewHTTPClient(10 * time.Second),
		filetree: make(map[string]string),
	}
	go p.refreshFiletree(context.Background())
	return p
}

func (p *Provider) Name() string {
	return Name
}

func (p *Provider) refreshFiletree(ctx context.Context) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, FiletreeURL, nil)
	if err != nil {
		return err
	}

	resp, err := p.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("gfriends returned status: %d", resp.StatusCode)
	}

	var rawTree map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&rawTree); err != nil {
		return err
	}

	tree := make(map[string]string)
	if content, ok := rawTree["Content"].(map[string]interface{}); ok {
		for folder, files := range content {
			if fileMap, ok := files.(map[string]interface{}); ok {
				for filename, rel := range fileMap {
					if relStr, ok := rel.(string); ok {
						tree[filename] = fmt.Sprintf("%s/%s", folder, relStr)
					}
				}
			}
		}
	}

	p.filetree = tree
	p.loadedAt = time.Now()
	return nil
}

func (p *Provider) GetActorAvatarURL(name string) string {
	p.mu.RLock()
	defer p.mu.RUnlock()

	target := fmt.Sprintf("%s.jpg", name)
	if rel, exists := p.filetree[target]; exists {
		return fmt.Sprintf("%s/%s", ContentBase, url.PathEscape(rel))
	}
	return ""
}

func (p *Provider) SearchActor(ctx context.Context, keyword string) ([]*model.ActorSearchResult, error) {
	avatar := p.GetActorAvatarURL(keyword)
	if avatar == "" {
		return nil, nil
	}

	return []*model.ActorSearchResult{
		{
			ID:       keyword,
			Provider: Name,
			Name:     keyword,
			Images:   []string{avatar},
		},
	}, nil
}

func (p *Provider) GetActorInfo(ctx context.Context, id string) (*model.ActorInfo, error) {
	avatar := p.GetActorAvatarURL(id)
	return &model.ActorInfo{
		ActorSearchResult: model.ActorSearchResult{
			ID:       id,
			Provider: Name,
			Name:     id,
			Images:   []string{avatar},
		},
	}, nil
}

func (p *Provider) SearchMovie(ctx context.Context, keyword string) ([]*model.MovieSearchResult, error) {
	return nil, nil
}

func (p *Provider) GetMovieInfo(ctx context.Context, id string) (*model.MovieInfo, error) {
	return nil, nil
}
