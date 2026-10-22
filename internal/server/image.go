package server

import (
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/RoaycL/emby-scraper-plus/internal/provider"
	"github.com/gin-gonic/gin"
)

var imageClient = provider.NewHTTPClient(20 * time.Second)

func (h *Handler) ProxyImage(c *gin.Context) {
	rawURL := strings.TrimSpace(c.Query("url"))
	provName := strings.TrimSpace(c.Param("provider"))
	id := strings.TrimSpace(c.Param("id"))
	path := c.Request.URL.Path

	if rawURL == "" && id != "" {
		if strings.EqualFold(provName, "GFRIENDS") {
			rawURL = h.engine.GetActorAvatar(id)
		} else {
			info, err := h.engine.SmartGetMovieInfo(c.Request.Context(), id)
			if err == nil && info != nil {
				if strings.Contains(path, "/primary") {
					rawURL = info.CoverURL
					if rawURL == "" {
						rawURL = info.BigCoverURL
					}
				} else if strings.Contains(path, "/thumb") {
					rawURL = info.ThumbURL
					if rawURL == "" {
						rawURL = info.BigThumbURL
					}
					if rawURL == "" {
						rawURL = info.CoverURL
					}
				} else if strings.Contains(path, "/backdrop") {
					rawURL = info.BigCoverURL
					if rawURL == "" {
						rawURL = info.CoverURL
					}
					if rawURL == "" && len(info.PreviewImages) > 0 {
						rawURL = info.PreviewImages[0]
					}
				}
			}
		}
	}

	if rawURL == "" {
		c.JSON(http.StatusNotFound, gin.H{"error": "image not found"})
		return
	}

	req, err := http.NewRequestWithContext(c.Request.Context(), http.MethodGet, rawURL, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	req.Header.Set("User-Agent", provider.DefaultUserAgent)
	if strings.Contains(rawURL, "dmm.co.jp") || strings.Contains(rawURL, "dmm.com") {
		req.Header.Set("Referer", "https://video.dmm.co.jp/")
	}

	resp, err := imageClient.Do(req)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		c.Status(resp.StatusCode)
		return
	}

	contentType := resp.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "image/jpeg"
	}

	c.Header("Content-Type", contentType)
	c.Header("Cache-Control", "public, max-age=604800")
	c.Status(http.StatusOK)

	_, _ = io.Copy(c.Writer, resp.Body)
}
