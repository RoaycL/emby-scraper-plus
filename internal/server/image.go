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

func proxyImageHandler(c *gin.Context) {
	rawURL := c.Query("url")
	if rawURL == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing url query parameter"})
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
