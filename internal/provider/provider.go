package provider

import (
	"context"
	"crypto/tls"
	"net/http"
	"net/url"
	"time"

	"github.com/RoaycL/emby-scraper-plus/config"
	"github.com/RoaycL/emby-scraper-plus/internal/model"
)

// Provider defines the standard metadata provider interface.
type Provider interface {
	Name() string
	SearchMovie(ctx context.Context, keyword string) ([]*model.MovieSearchResult, error)
	GetMovieInfo(ctx context.Context, id string) (*model.MovieInfo, error)
	SearchActor(ctx context.Context, keyword string) ([]*model.ActorSearchResult, error)
	GetActorInfo(ctx context.Context, id string) (*model.ActorInfo, error)
}

// NewHTTPClient creates an HTTP client configured with timeout and optional proxy.
func NewHTTPClient(timeout time.Duration) *http.Client {
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	if config.GlobalConfig.Proxy != "" {
		if proxyURL, err := url.Parse(config.GlobalConfig.Proxy); err == nil {
			transport.Proxy = http.ProxyURL(proxyURL)
		}
	}

	return &http.Client{
		Timeout:   timeout,
		Transport: transport,
	}
}

const DefaultUserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/126.0.0.0 Safari/537.36"
