// Package xiaoyuzhou is the library behind the xiaoyuzhou command: the HTTP
// client, request shaping, and typed data models for Xiaoyuzhou (小宇宙).
//
// The Client fetches pages from https://www.xiaoyuzhoufm.com, extracts the
// __NEXT_DATA__ JSON blob embedded in each HTML page, and returns typed
// structs. No API key or authentication is required.
package xiaoyuzhou

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"
)

// Host is the canonical hostname for Xiaoyuzhou.
const Host = "www.xiaoyuzhoufm.com"

// baseURL is the URL prefix used to build page URLs.
const baseURL = "https://www.xiaoyuzhoufm.com"

// DefaultUserAgent is the browser User-Agent sent with each request.
// Xiaoyuzhou requires a real User-Agent; without one the server returns 403.
const DefaultUserAgent = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36"

// ErrNotFound is returned when a podcast or episode is not found in pageProps.
var ErrNotFound = errors.New("not found")

// Config holds constructor parameters for the Client.
type Config struct {
	BaseURL   string
	UserAgent string
	Rate      time.Duration
	Timeout   time.Duration
	Retries   int
}

// DefaultConfig returns sensible defaults.
func DefaultConfig() Config {
	return Config{
		BaseURL:   baseURL,
		UserAgent: DefaultUserAgent,
		Rate:      300 * time.Millisecond,
		Timeout:   20 * time.Second,
		Retries:   3,
	}
}

// Client talks to Xiaoyuzhou over HTTP.
type Client struct {
	cfg        Config
	httpClient *http.Client
	mu         sync.Mutex
	last       time.Time
}

// NewClient returns a Client configured with cfg.
func NewClient(cfg Config) *Client {
	return &Client{
		cfg:        cfg,
		httpClient: &http.Client{Timeout: cfg.Timeout},
	}
}

// fetchPage GETs url with the configured User-Agent and extracts pageProps
// from the embedded __NEXT_DATA__ JSON.
func (c *Client) fetchPage(ctx context.Context, url string) (map[string]any, error) {
	body, err := c.get(ctx, url)
	if err != nil {
		return nil, err
	}
	return getPageProps(body)
}

// GetPodcast fetches the podcast page for id and returns a Podcast.
func (c *Client) GetPodcast(ctx context.Context, id string) (*Podcast, error) {
	url := c.cfg.BaseURL + "/podcast/" + id
	props, err := c.fetchPage(ctx, url)
	if err != nil {
		return nil, err
	}
	p, _ := props["podcast"].(map[string]any)
	if p == nil {
		return nil, fmt.Errorf("%w: podcast %s", ErrNotFound, id)
	}
	return &Podcast{
		ID:                id,
		Title:             strVal(p["title"]),
		Author:            strVal(p["author"]),
		SubscriptionCount: intVal(p["subscriptionCount"]),
		EpisodeCount:      intVal(p["episodeCount"]),
		Description:       strVal(p["description"]),
		URL:               url,
	}, nil
}

// GetEpisode fetches the episode page for id and returns an Episode.
func (c *Client) GetEpisode(ctx context.Context, id string) (*Episode, error) {
	url := c.cfg.BaseURL + "/episode/" + id
	props, err := c.fetchPage(ctx, url)
	if err != nil {
		return nil, err
	}
	ep, _ := props["episode"].(map[string]any)
	if ep == nil {
		return nil, fmt.Errorf("%w: episode %s", ErrNotFound, id)
	}
	podTitle := ""
	if pod, ok := ep["podcast"].(map[string]any); ok {
		podTitle = strVal(pod["title"])
	}
	return &Episode{
		EID:          id,
		Title:        strVal(ep["title"]),
		PodcastTitle: podTitle,
		DurationSecs: floatVal(ep["duration"]),
		PlayCount:    intVal(ep["playCount"]),
		CommentCount: intVal(ep["commentCount"]),
		ClapCount:    intVal(ep["clapCount"]),
		PubDate:      strVal(ep["pubDate"]),
		URL:          url,
	}, nil
}

// ListEpisodes fetches the podcast page for podcastID and returns up to limit episodes.
// A limit of 0 returns all available episodes.
func (c *Client) ListEpisodes(ctx context.Context, podcastID string, limit int) ([]*Episode, error) {
	url := c.cfg.BaseURL + "/podcast/" + podcastID
	props, err := c.fetchPage(ctx, url)
	if err != nil {
		return nil, err
	}
	podcast, _ := props["podcast"].(map[string]any)
	if podcast == nil {
		return nil, fmt.Errorf("%w: podcast %s", ErrNotFound, podcastID)
	}
	allEps, _ := podcast["episodes"].([]any)
	out := make([]*Episode, 0, len(allEps))
	for _, raw := range allEps {
		ep, _ := raw.(map[string]any)
		if ep == nil {
			continue
		}
		eid := strVal(ep["eid"])
		out = append(out, &Episode{
			EID:          eid,
			Title:        strVal(ep["title"]),
			DurationSecs: floatVal(ep["duration"]),
			PlayCount:    intVal(ep["playCount"]),
			PubDate:      strVal(ep["pubDate"]),
			URL:          c.cfg.BaseURL + "/episode/" + eid,
		})
		if limit > 0 && len(out) >= limit {
			break
		}
	}
	return out, nil
}

// Podcast is an alias for backward compatibility.
// Deprecated: use GetPodcast.
func (c *Client) Podcast(ctx context.Context, id string) (*Podcast, error) {
	return c.GetPodcast(ctx, id)
}

// Episode is an alias for backward compatibility.
// Deprecated: use GetEpisode.
func (c *Client) Episode(ctx context.Context, id string) (*Episode, error) {
	return c.GetEpisode(ctx, id)
}

// Episodes is an alias for backward compatibility.
// Deprecated: use ListEpisodes.
func (c *Client) Episodes(ctx context.Context, podcastID string) ([]*Episode, error) {
	return c.ListEpisodes(ctx, podcastID, 0)
}

// get performs a GET with retry and pacing.
func (c *Client) get(ctx context.Context, url string) ([]byte, error) {
	var lastErr error
	for attempt := 0; attempt <= c.cfg.Retries; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(backoff(attempt)):
			}
		}
		body, retry, err := c.do(ctx, url)
		if err == nil {
			return body, nil
		}
		lastErr = err
		if !retry {
			return nil, err
		}
	}
	return nil, fmt.Errorf("get %s: %w", url, lastErr)
}

func (c *Client) do(ctx context.Context, url string) ([]byte, bool, error) {
	c.pace()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, false, err
	}
	req.Header.Set("User-Agent", c.cfg.UserAgent)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, true, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode >= 500 {
		return nil, true, fmt.Errorf("http %d", resp.StatusCode)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, false, fmt.Errorf("http %d", resp.StatusCode)
	}

	b, err := io.ReadAll(io.LimitReader(resp.Body, 8<<20))
	if err != nil {
		return nil, true, err
	}
	return b, false, nil
}

func (c *Client) pace() {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.cfg.Rate <= 0 {
		return
	}
	if wait := c.cfg.Rate - time.Since(c.last); wait > 0 {
		time.Sleep(wait)
	}
	c.last = time.Now()
}

func backoff(attempt int) time.Duration {
	d := time.Duration(attempt) * 500 * time.Millisecond
	if d > 5*time.Second {
		d = 5 * time.Second
	}
	return d
}

func intVal(v any) int {
	switch x := v.(type) {
	case float64:
		return int(x)
	case int:
		return x
	case int64:
		return int(x)
	}
	return 0
}

func floatVal(v any) float64 {
	switch x := v.(type) {
	case float64:
		return x
	case int:
		return float64(x)
	case int64:
		return float64(x)
	}
	return 0
}
