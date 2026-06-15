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
        "eid": "ep-abc",
        "title": "AI 时代的投资逻辑",
        "podcast": {"title": "硅谷101"},
        "duration": 3661,
        "playCount": 12345,
        "commentCount": 88,
        "clapCount": 500,
        "pubDate": "2024-02-01"
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
	p, err := c.GetPodcast(context.Background(), "test-id")
	if err != nil {
		t.Fatal(err)
	}
	if p.ID != "test-id" {
		t.Errorf("id = %q, want test-id", p.ID)
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
	ep, err := c.GetEpisode(context.Background(), "ep-abc")
	if err != nil {
		t.Fatal(err)
	}
	if ep.EID != "ep-abc" {
		t.Errorf("eid = %q, want ep-abc", ep.EID)
	}
	if ep.Title != "AI 时代的投资逻辑" {
		t.Errorf("title = %q", ep.Title)
	}
	if ep.PodcastTitle != "硅谷101" {
		t.Errorf("podcast title = %q, want 硅谷101", ep.PodcastTitle)
	}
	// 3661 seconds
	if ep.DurationSecs != 3661 {
		t.Errorf("DurationSecs = %v, want 3661", ep.DurationSecs)
	}
	if ep.PlayCount != 12345 {
		t.Errorf("playCount = %v, want 12345", ep.PlayCount)
	}
	if ep.CommentCount != 88 {
		t.Errorf("commentCount = %v, want 88", ep.CommentCount)
	}
	if ep.PubDate != "2024-02-01" {
		t.Errorf("pubDate = %q, want 2024-02-01", ep.PubDate)
	}
}

func TestListEpisodesReturnsList(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(wrapNextData(podcastJSON)))
	}))
	defer srv.Close()

	c := newTestClient(srv)
	eps, err := c.ListEpisodes(context.Background(), "test-id", 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(eps) != 2 {
		t.Fatalf("got %d episodes, want 2", len(eps))
	}
	if eps[0].EID != "ep001" {
		t.Errorf("eid = %q, want ep001", eps[0].EID)
	}
	if eps[0].DurationSecs != 3661 {
		t.Errorf("DurationSecs = %v, want 3661", eps[0].DurationSecs)
	}
	// url uses BaseURL + /episode/ + eid
	wantURL := srv.URL + "/episode/ep001"
	if eps[0].URL != wantURL {
		t.Errorf("url = %q, want %q", eps[0].URL, wantURL)
	}
}

func TestListEpisodesLimit(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(wrapNextData(podcastJSON)))
	}))
	defer srv.Close()

	c := newTestClient(srv)
	eps, err := c.ListEpisodes(context.Background(), "test-id", 1)
	if err != nil {
		t.Fatal(err)
	}
	if len(eps) != 1 {
		t.Errorf("got %d episodes with limit=1, want 1", len(eps))
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
	_, err := c.GetPodcast(context.Background(), "any")
	if err != nil {
		t.Fatal(err)
	}
	if gotUA == "" {
		t.Error("request carried no User-Agent header")
	}
}

func TestDurationSecsRaw(t *testing.T) {
	cases := []struct {
		secs float64
	}{
		{90},
		{3661},
		{60},
	}
	for _, tc := range cases {
		epJSON := `{"props":{"pageProps":{"episode":{"eid":"id","title":"T","duration":` +
			fmt.Sprintf("%v", tc.secs) + `,"podcast":{"title":"P"}}}}}`
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, _ = w.Write([]byte(wrapNextData(epJSON)))
		}))
		c := newTestClient(srv)
		ep, err := c.GetEpisode(context.Background(), "id")
		srv.Close()
		if err != nil {
			t.Fatalf("secs=%v: %v", tc.secs, err)
		}
		if ep.DurationSecs != tc.secs {
			t.Errorf("secs=%v DurationSecs=%v, want %v", tc.secs, ep.DurationSecs, tc.secs)
		}
	}
}
