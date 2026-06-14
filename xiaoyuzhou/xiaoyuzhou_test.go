package xiaoyuzhou_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/tamnd/xiaoyuzhou-cli/xiaoyuzhou"
)

// wrapNextData wraps JSON in an HTML page with __NEXT_DATA__.
func wrapNextData(json string) string {
	return `<!DOCTYPE html><html><head>` +
		`<script id="__NEXT_DATA__" type="application/json">` + json + `</script>` +
		`</head><body></body></html>`
}

const podcastJSON = `{
  "props": {
    "pageProps": {
      "podcast": {
        "title": "硅谷101",
        "author": "翻翻",
        "subscriptionCount": 123456,
        "episodeCount": 88,
        "description": "科技创投播客",
        "episodes": [
          {"eid": "ep001", "title": "First Episode", "duration": 3661, "playCount": 5000, "pubDate": "2024-01-15"},
          {"eid": "ep002", "title": "Second Episode", "duration": 1800, "playCount": 3200, "pubDate": "2024-01-08"}
        ]
      }
    }
  }
}`

const episodeJSON = `{
  "props": {
    "pageProps": {
      "episode": {
        "title": "AI 时代的投资逻辑",
        "podcast": {"title": "硅谷101"},
        "duration": 3661,
        "playCount": 12345,
        "commentCount": 88,
        "clapCount": 500
      }
    }
  }
}`

type Client = xiaoyuzhou.Client

func DefaultConfig() xiaoyuzhou.Config { return xiaoyuzhou.DefaultConfig() }
func NewClient(cfg xiaoyuzhou.Config) *xiaoyuzhou.Client { return xiaoyuzhou.NewClient(cfg) }

func newTestClient(ts *httptest.Server) *Client {
	cfg := DefaultConfig()
	cfg.BaseURL = ts.URL
	cfg.Rate = 0
	return NewClient(cfg)
}

func TestPodcastParsesFields(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(wrapNextData(podcastJSON)))
	}))
	defer srv.Close()

	c := newTestClient(srv)
	p, err := c.Podcast(context.Background(), "test-id")
	if err != nil {
		t.Fatal(err)
	}
	if p.Title != "硅谷101" {
		t.Errorf("title = %q, want 硅谷101", p.Title)
	}
	if p.Author != "翻翻" {
		t.Errorf("author = %q, want 翻翻", p.Author)
	}
	if p.SubscriptionCount != 123456 {
		t.Errorf("subscriptionCount = %v, want 123456", p.SubscriptionCount)
	}
	if p.EpisodeCount != 88 {
		t.Errorf("episodeCount = %v, want 88", p.EpisodeCount)
	}
	if p.Description != "科技创投播客" {
		t.Errorf("description = %q", p.Description)
	}
}

func TestEpisodeParsesFields(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(wrapNextData(episodeJSON)))
	}))
	defer srv.Close()

	c := newTestClient(srv)
	ep, err := c.Episode(context.Background(), "ep-abc")
	if err != nil {
		t.Fatal(err)
	}
	if ep.Title != "AI 时代的投资逻辑" {
		t.Errorf("title = %q", ep.Title)
	}
	if ep.PodcastTitle != "硅谷101" {
		t.Errorf("podcast title = %q, want 硅谷101", ep.PodcastTitle)
	}
	// 3661 seconds = 61 minutes 1 second → "61:01"
	if ep.Duration != "61:01" {
		t.Errorf("duration = %q, want 61:01", ep.Duration)
	}
	if ep.PlayCount != 12345 {
		t.Errorf("playCount = %v, want 12345", ep.PlayCount)
	}
	if ep.CommentCount != 88 {
		t.Errorf("commentCount = %v, want 88", ep.CommentCount)
	}
}

func TestEpisodesReturnsList(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(wrapNextData(podcastJSON)))
	}))
	defer srv.Close()

	c := newTestClient(srv)
	eps, err := c.Episodes(context.Background(), "test-id")
	if err != nil {
		t.Fatal(err)
	}
	if len(eps) != 2 {
		t.Fatalf("got %d episodes, want 2", len(eps))
	}
	if eps[0].EID != "ep001" {
		t.Errorf("eid = %q, want ep001", eps[0].EID)
	}
	// 3661s = 61:01
	if eps[0].Duration != "61:01" {
		t.Errorf("duration = %q, want 61:01", eps[0].Duration)
	}
	// url uses BaseURL + /episode/ + eid
	wantURL := srv.URL + "/episode/ep001"
	if eps[0].URL != wantURL {
		t.Errorf("url = %q, want %q", eps[0].URL, wantURL)
	}
}

func TestClientSendsUserAgent(t *testing.T) {
	var gotUA string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotUA = r.Header.Get("User-Agent")
		_, _ = w.Write([]byte(wrapNextData(podcastJSON)))
	}))
	defer srv.Close()

	c := newTestClient(srv)
	_, err := c.Podcast(context.Background(), "any")
	if err != nil {
		t.Fatal(err)
	}
	if gotUA == "" {
		t.Error("request carried no User-Agent header")
	}
}

func TestDurationViaEpisode(t *testing.T) {
	cases := []struct {
		secs float64
		want string
	}{
		{90, "1:30"},
		{3661, "61:01"},
		{60, "1:00"},
	}
	for _, tc := range cases {
		epJSON := `{"props":{"pageProps":{"episode":{"title":"T","duration":` +
			formatFloat(tc.secs) + `,"podcast":{"title":"P"}}}}}`
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, _ = w.Write([]byte(wrapNextData(epJSON)))
		}))
		c := newTestClient(srv)
		ep, err := c.Episode(context.Background(), "id")
		srv.Close()
		if err != nil {
			t.Fatalf("secs=%v: %v", tc.secs, err)
		}
		if ep.Duration != tc.want {
			t.Errorf("secs=%v duration=%q, want %q", tc.secs, ep.Duration, tc.want)
		}
	}
}

func formatFloat(f float64) string {
	return fmt.Sprintf("%v", f)
}
