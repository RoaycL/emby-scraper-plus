package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/RoaycL/emby-scraper-plus/config"
	"github.com/RoaycL/emby-scraper-plus/internal/model"
	"github.com/RoaycL/emby-scraper-plus/internal/router"
)

func TestMetaTubeEndpoints(t *testing.T) {
	config.GlobalConfig.Token = "testtoken"
	engine := router.NewEngine()
	r := SetupRouter(engine)

	// 1. Test Search Movie
	req, _ := http.NewRequest(http.MethodGet, "/v1/movies/search?q=SSIS-057", nil)
	req.Header.Set("Authorization", "Bearer testtoken")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	var searchResp model.Response[[]*model.MovieSearchResult]
	if err := json.Unmarshal(w.Body.Bytes(), &searchResp); err != nil {
		t.Fatalf("unmarshal search failed: %v", err)
	}
	if len(searchResp.Data) == 0 {
		t.Fatalf("expected search results, got 0")
	}
	if searchResp.Data[0].Title == "" {
		t.Errorf("search result title should not be empty")
	}

	// 2. Test Get Movie Info (ensure embedded MovieInfo fields are not omitted)
	req2, _ := http.NewRequest(http.MethodGet, "/v1/movies/SMART/SSIS-057", nil)
	req2.Header.Set("Authorization", "Bearer testtoken")
	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, req2)
	if w2.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w2.Code, w2.Body.String())
	}
	bodyStr := w2.Body.String()
	if !containsField(bodyStr, "maker") || !containsField(bodyStr, "genres") || !containsField(bodyStr, "summary") {
		t.Fatalf("MovieInfo fields omitted in JSON response: %s", bodyStr)
	}

	// 3. Test Image Proxy without url param (auto lookup by id)
	req3, _ := http.NewRequest(http.MethodGet, "/v1/images/primary/SMART/SSIS-057", nil)
	w3 := httptest.NewRecorder()
	r.ServeHTTP(w3, req3)
	if w3.Code != http.StatusOK {
		t.Fatalf("expected 200 for image proxy, got %d", w3.Code)
	}
	if w3.Header().Get("Content-Type") != "image/jpeg" {
		t.Errorf("expected image/jpeg, got %s", w3.Header().Get("Content-Type"))
	}

	// 4. Test Translation endpoint
	req4, _ := http.NewRequest(http.MethodGet, "/v1/translate?q=Hello&from=en&to=zh", nil)
	w4 := httptest.NewRecorder()
	r.ServeHTTP(w4, req4)
	if w4.Code != http.StatusOK {
		t.Fatalf("expected 200 for translate, got %d", w4.Code)
	}
	var transResp struct {
		Data struct {
			From           string `json:"from"`
			To             string `json:"to"`
			TranslatedText string `json:"translated_text"`
		} `json:"data"`
	}
	if err := json.Unmarshal(w4.Body.Bytes(), &transResp); err != nil {
		t.Fatalf("unmarshal translate response failed: %v", err)
	}
	if transResp.Data.TranslatedText != "Hello" {
		t.Errorf("expected translated_text Hello, got %s", transResp.Data.TranslatedText)
	}

	// 5. Test double slash handling
	req5, _ := http.NewRequest(http.MethodGet, "http://localhost//v1/movies/SMART/SSIS-057", nil)
	req5.Header.Set("Authorization", "Bearer testtoken")
	w5 := httptest.NewRecorder()
	r.ServeHTTP(w5, req5)
	if w5.Code != http.StatusOK {
		t.Fatalf("expected 200 for double slash, got %d", w5.Code)
	}
}

func containsField(jsonStr, field string) bool {
	var m map[string]interface{}
	_ = json.Unmarshal([]byte(jsonStr), &m)
	data, ok := m["data"].(map[string]interface{})
	if !ok {
		return false
	}
	_, exists := data[field]
	return exists
}
