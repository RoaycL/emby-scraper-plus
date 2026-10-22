package router

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/RoaycL/emby-scraper-plus/internal/model"
	"github.com/RoaycL/emby-scraper-plus/internal/provider"
	"github.com/RoaycL/emby-scraper-plus/internal/provider/avbase"
	"github.com/RoaycL/emby-scraper-plus/internal/provider/dmm"
	"github.com/RoaycL/emby-scraper-plus/internal/provider/gfriends"
	"github.com/RoaycL/emby-scraper-plus/internal/provider/r18dev"
)

type Engine struct {
	providers map[string]provider.Provider
	gfriends  *gfriends.Provider
	cache     map[string]*cacheEntry
	cacheMu   sync.RWMutex
}

type cacheEntry struct {
	data      *model.MovieInfo
	expiresAt time.Time
}

func NewEngine() *Engine {
	gf := gfriends.New()
	eng := &Engine{
		providers: make(map[string]provider.Provider),
		gfriends:  gf,
		cache:     make(map[string]*cacheEntry),
	}

	eng.Register(dmm.New())
	eng.Register(r18dev.New())
	eng.Register(avbase.New())
	eng.Register(gf)

	return eng
}

func (e *Engine) Register(p provider.Provider) {
	e.providers[strings.ToUpper(p.Name())] = p
}

func (e *Engine) GetProvider(name string) (provider.Provider, bool) {
	p, ok := e.providers[strings.ToUpper(name)]
	return p, ok
}

// GetRouteChain returns the prioritized provider chain for the detected content type.
func (e *Engine) GetRouteChain(t ContentType) []string {
	switch t {
	case TypeCensored:
		return []string{"DMM", "R18DEV", "AVBASE"}
	case TypeAmateur:
		return []string{"AVBASE", "DMM", "R18DEV"}
	case TypeUncensored:
		return []string{"AVBASE"}
	case TypeFC2:
		return []string{"AVBASE"}
	default:
		return []string{"DMM", "R18DEV", "AVBASE"}
	}
}

// SmartGetMovieInfo retrieves movie metadata by dynamically selecting the best source chain.
func (e *Engine) SmartGetMovieInfo(ctx context.Context, code string) (*model.MovieInfo, error) {
	clean := strings.TrimSpace(code)

	e.cacheMu.RLock()
	if entry, exists := e.cache[strings.ToUpper(clean)]; exists && time.Now().Before(entry.expiresAt) {
		e.cacheMu.RUnlock()
		return entry.data, nil
	}
	e.cacheMu.RUnlock()

	cType := ClassifyNumber(clean)
	chain := e.GetRouteChain(cType)

	var lastErr error
	var finalInfo *model.MovieInfo

	for _, provName := range chain {
		prov, exists := e.GetProvider(provName)
		if !exists {
			continue
		}

		info, err := prov.GetMovieInfo(ctx, clean)
		if err != nil {
			lastErr = err
			continue
		}

		if info != nil && info.Title != "" {
			finalInfo = info
			// If essential fields are satisfied, short-circuit
			if info.CoverURL != "" && len(info.Actors) > 0 && info.Maker != "" {
				break
			}
		}
	}

	if finalInfo == nil {
		if lastErr != nil {
			return nil, lastErr
		}
		return nil, fmt.Errorf("no metadata found across providers for code: %s", code)
	}

	// Cache result for 24 hours
	e.cacheMu.Lock()
	e.cache[strings.ToUpper(clean)] = &cacheEntry{
		data:      finalInfo,
		expiresAt: time.Now().Add(24 * time.Hour),
	}
	e.cacheMu.Unlock()

	return finalInfo, nil
}

// SmartSearch searches movies using the smart routing chain.
func (e *Engine) SmartSearch(ctx context.Context, keyword string) ([]*model.MovieSearchResult, error) {
	clean := strings.TrimSpace(keyword)
	cType := ClassifyNumber(clean)
	chain := e.GetRouteChain(cType)

	for _, provName := range chain {
		prov, exists := e.GetProvider(provName)
		if !exists {
			continue
		}

		results, err := prov.SearchMovie(ctx, clean)
		if err == nil && len(results) > 0 {
			return results, nil
		}
	}

	return nil, fmt.Errorf("no search results for keyword: %s", keyword)
}

// GetActorAvatar returns avatar URL for a given actor name from Gfriends.
func (e *Engine) GetActorAvatar(name string) string {
	if e.gfriends != nil {
		return e.gfriends.GetActorAvatarURL(name)
	}
	return ""
}
