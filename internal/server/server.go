package server

import (
	"fmt"
	"net/http"

	"github.com/RoaycL/emby-scraper-plus/config"
	"github.com/RoaycL/emby-scraper-plus/internal/router"
	"github.com/RoaycL/emby-scraper-plus/internal/web"
	"github.com/gin-gonic/gin"
)

func SetupRouter(engine *router.Engine) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()

	// CORS middleware
	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	})

	// Web Dashboard
	r.GET("/", func(c *gin.Context) {
		c.Data(http.StatusOK, "text/html; charset=utf-8", web.IndexHTML)
	})

	r.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
			"time":   c.GetString("time"),
		})
	})

	h := NewHandler(engine)

	v1 := r.Group("/v1")
	{
		// Movie endpoints (MetaTube spec)
		v1.GET("/movies/search", h.SearchMovie)
		v1.GET("/movies/:provider/:id", h.GetMovie)

		// Actor endpoints (MetaTube spec)
		v1.GET("/actors/search", h.SearchActor)
		v1.GET("/actors/:provider/:id", h.GetActor)

		// Image Proxy endpoints
		v1.GET("/images/primary/:provider/:id", proxyImageHandler)
		v1.GET("/images/thumb/:provider/:id", proxyImageHandler)
		v1.GET("/images/backdrop/:provider/:id", proxyImageHandler)

		// Direct NFO download endpoint
		v1.GET("/nfo/:provider/:id", h.DownloadNFO)

		// Translation passthrough
		v1.GET("/translate", func(c *gin.Context) {
			text := c.Query("text")
			c.JSON(http.StatusOK, gin.H{"data": text})
		})
	}

	return r
}

func StartServer(engine *router.Engine) error {
	r := SetupRouter(engine)
	addr := fmt.Sprintf("%s:%d", config.GlobalConfig.Host, config.GlobalConfig.Port)
	return r.Run(addr)
}
