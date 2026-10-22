package server

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/RoaycL/emby-scraper-plus/config"
	"github.com/RoaycL/emby-scraper-plus/internal/model"
	"github.com/RoaycL/emby-scraper-plus/internal/router"
	"github.com/RoaycL/emby-scraper-plus/internal/web"
	"github.com/gin-gonic/gin"
)

func SetupRouter(engine *router.Engine) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()
	r.RemoveExtraSlash = true

	// CORS middleware
	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Token")
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
		})
	})

	h := NewHandler(engine)

	v1 := r.Group("/v1")
	// Token authentication middleware
	v1.Use(func(c *gin.Context) {
		serverToken := strings.TrimSpace(config.GlobalConfig.Token)
		if serverToken == "" {
			c.Next()
			return
		}

		// Skip image proxy and translate endpoints from strict auth
		if strings.HasPrefix(c.Request.URL.Path, "/v1/images/") || strings.HasPrefix(c.Request.URL.Path, "/v1/translate") {
			c.Next()
			return
		}

		clientToken := ""
		authHeader := c.GetHeader("Authorization")
		if authHeader != "" {
			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) == 2 && strings.EqualFold(parts[0], "Bearer") {
				clientToken = strings.TrimSpace(parts[1])
			} else {
				clientToken = strings.TrimSpace(authHeader)
			}
		}
		if clientToken == "" {
			clientToken = strings.TrimSpace(c.GetHeader("X-Token"))
		}
		if clientToken == "" {
			clientToken = strings.TrimSpace(c.Query("token"))
		}

		if clientToken != serverToken {
			c.AbortWithStatusJSON(http.StatusUnauthorized, model.Response[any]{
				Error: &model.ErrorInfo{
					Code:    http.StatusUnauthorized,
					Message: "unauthorized: invalid or missing token",
				},
			})
			return
		}
		c.Next()
	})

	{
		// Movie endpoints (MetaTube spec)
		v1.GET("/movies/search", h.SearchMovie)
		v1.GET("/movies/:provider/:id", h.GetMovie)

		// Actor endpoints (MetaTube spec)
		v1.GET("/actors/search", h.SearchActor)
		v1.GET("/actors/:provider/:id", h.GetActor)

		// Image Proxy endpoints (support both GET and HEAD)
		v1.Match([]string{http.MethodGet, http.MethodHead}, "/images/primary/:provider/:id", h.ProxyImage)
		v1.Match([]string{http.MethodGet, http.MethodHead}, "/images/thumb/:provider/:id", h.ProxyImage)
		v1.Match([]string{http.MethodGet, http.MethodHead}, "/images/backdrop/:provider/:id", h.ProxyImage)

		// Direct NFO download endpoint
		v1.GET("/nfo/:provider/:id", h.DownloadNFO)

		// Translation passthrough (MetaTube TranslationInfo format)
		v1.GET("/translate", func(c *gin.Context) {
			q := c.Query("q")
			if q == "" {
				q = c.Query("text")
			}
			from := c.Query("from")
			to := c.Query("to")
			c.JSON(http.StatusOK, gin.H{
				"data": gin.H{
					"from":            from,
					"to":              to,
					"translated_text": q,
				},
			})
		})
	}

	return r
}

func StartServer(engine *router.Engine) error {
	r := SetupRouter(engine)
	addr := fmt.Sprintf("%s:%d", config.GlobalConfig.Host, config.GlobalConfig.Port)
	return r.Run(addr)
}
