package server

import (
	"net/http"
	"strings"

	"github.com/RoaycL/emby-scraper-plus/internal/model"
	"github.com/RoaycL/emby-scraper-plus/internal/nfo"
	"github.com/RoaycL/emby-scraper-plus/internal/router"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	engine *router.Engine
}

func NewHandler(engine *router.Engine) *Handler {
	return &Handler{engine: engine}
}

func (h *Handler) SearchMovie(c *gin.Context) {
	q := strings.TrimSpace(c.Query("q"))
	if q == "" {
		c.JSON(http.StatusOK, model.Response[[]*model.MovieSearchResult]{Data: []*model.MovieSearchResult{}})
		return
	}

	provName := strings.TrimSpace(c.Query("provider"))
	if provName != "" && strings.ToUpper(provName) != "SMART" {
		if p, ok := h.engine.GetProvider(provName); ok {
			results, err := p.SearchMovie(c.Request.Context(), q)
			if err != nil {
				c.JSON(http.StatusOK, model.Response[[]*model.MovieSearchResult]{
					Error: &model.ErrorInfo{Code: 404, Message: err.Error()},
				})
				return
			}
			c.JSON(http.StatusOK, model.Response[[]*model.MovieSearchResult]{Data: results})
			return
		}
	}

	// Smart routing search
	results, err := h.engine.SmartSearch(c.Request.Context(), q)
	if err != nil {
		c.JSON(http.StatusOK, model.Response[[]*model.MovieSearchResult]{
			Error: &model.ErrorInfo{Code: 404, Message: err.Error()},
		})
		return
	}

	c.JSON(http.StatusOK, model.Response[[]*model.MovieSearchResult]{Data: results})
}

func (h *Handler) GetMovie(c *gin.Context) {
	provName := strings.TrimSpace(c.Param("provider"))
	id := strings.TrimSpace(c.Param("id"))

	if id == "" {
		c.JSON(http.StatusBadRequest, model.Response[*model.MovieInfo]{
			Error: &model.ErrorInfo{Code: 400, Message: "id parameter required"},
		})
		return
	}

	var info *model.MovieInfo
	var err error

	if provName != "" && strings.ToUpper(provName) != "SMART" {
		if p, ok := h.engine.GetProvider(provName); ok {
			info, err = p.GetMovieInfo(c.Request.Context(), id)
		} else {
			info, err = h.engine.SmartGetMovieInfo(c.Request.Context(), id)
		}
	} else {
		info, err = h.engine.SmartGetMovieInfo(c.Request.Context(), id)
	}

	if err != nil || info == nil {
		msg := "movie not found"
		if err != nil {
			msg = err.Error()
		}
		c.JSON(http.StatusNotFound, model.Response[*model.MovieInfo]{
			Error: &model.ErrorInfo{Code: 404, Message: msg},
		})
		return
	}

	c.JSON(http.StatusOK, model.Response[*model.MovieInfo]{Data: info})
}

func (h *Handler) SearchActor(c *gin.Context) {
	q := strings.TrimSpace(c.Query("q"))
	if q == "" {
		c.JSON(http.StatusOK, model.Response[[]*model.ActorSearchResult]{Data: []*model.ActorSearchResult{}})
		return
	}

	avatar := h.engine.GetActorAvatar(q)
	var images []string
	if avatar != "" {
		images = append(images, avatar)
	}

	result := []*model.ActorSearchResult{
		{
			ID:       q,
			Provider: "GFRIENDS",
			Name:     q,
			Images:   images,
		},
	}

	c.JSON(http.StatusOK, model.Response[[]*model.ActorSearchResult]{Data: result})
}

func (h *Handler) GetActor(c *gin.Context) {
	id := strings.TrimSpace(c.Param("id"))
	avatar := h.engine.GetActorAvatar(id)

	var images []string
	if avatar != "" {
		images = append(images, avatar)
	}

	c.JSON(http.StatusOK, model.Response[*model.ActorInfo]{
		Data: &model.ActorInfo{
			ActorSearchResult: model.ActorSearchResult{
				ID:       id,
				Provider: "GFRIENDS",
				Name:     id,
				Images:   images,
			},
		},
	})
}

func (h *Handler) DownloadNFO(c *gin.Context) {
	provName := strings.TrimSpace(c.Param("provider"))
	id := strings.TrimSpace(c.Param("id"))

	var info *model.MovieInfo
	var err error

	if provName != "" && strings.ToUpper(provName) != "SMART" {
		if p, ok := h.engine.GetProvider(provName); ok {
			info, err = p.GetMovieInfo(c.Request.Context(), id)
		} else {
			info, err = h.engine.SmartGetMovieInfo(c.Request.Context(), id)
		}
	} else {
		info, err = h.engine.SmartGetMovieInfo(c.Request.Context(), id)
	}

	if err != nil || info == nil {
		c.String(http.StatusNotFound, "not found")
		return
	}

	c.Header("Content-Type", "application/xml; charset=utf-8")
	c.Header("Content-Disposition", `attachment; filename="`+info.Number+`.nfo"`)
	_ = nfo.WriteXML(c.Writer, info)
}
