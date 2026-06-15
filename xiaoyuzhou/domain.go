package xiaoyuzhou

import (
	"context"

	"github.com/tamnd/any-cli/kit"
	"github.com/tamnd/any-cli/kit/errs"
)

func init() { kit.Register(Domain{}) }

// Domain is the Xiaoyuzhou kit driver.
type Domain struct{}

func (Domain) Info() kit.DomainInfo {
	return kit.DomainInfo{
		Scheme: "xiaoyuzhou",
		Hosts:  []string{Host},
		Identity: kit.Identity{
			Binary: "xiaoyuzhou",
			Short:  "A command line for Xiaoyuzhou FM.",
			Long: `A command line for Xiaoyuzhou FM (xiaoyuzhoufm.com).

Fetches podcast profiles and episode listings from Xiaoyuzhou (小宇宙),
China's leading independent podcast platform.

No API key required. Data is scraped from the public __NEXT_DATA__ JSON
embedded in each page.

xiaoyuzhou is an independent tool and is not affiliated with Xiaoyuzhou.`,
			Site: "https://" + Host,
			Repo: "https://github.com/tamnd/xiaoyuzhou-cli",
		},
	}
}

func (Domain) Register(app *kit.App) {
	app.SetClient(newKitClient)

	kit.Handle(app, kit.OpMeta{Name: "podcast", Group: "podcast", Single: true,
		URIType: "podcast", Summary: "Show a Xiaoyuzhou podcast profile",
		Args: []kit.Arg{{Name: "id", Help: "podcast ID"}}}, getPodcast)

	kit.Handle(app, kit.OpMeta{Name: "episode", Group: "podcast", Single: true,
		URIType: "episode", Summary: "Show a Xiaoyuzhou episode",
		Args: []kit.Arg{{Name: "id", Help: "episode ID"}}}, getEpisode)

	kit.Handle(app, kit.OpMeta{Name: "episodes", Group: "podcast", List: true,
		URIType: "episode", Summary: "List recent episodes of a Xiaoyuzhou podcast",
		Args: []kit.Arg{{Name: "podcast-id", Help: "podcast ID"}}}, listEpisodes)
}

func newKitClient(_ context.Context, cfg kit.Config) (any, error) {
	c := DefaultConfig()
	if cfg.UserAgent != "" {
		c.UserAgent = cfg.UserAgent
	}
	if cfg.Rate > 0 {
		c.Rate = cfg.Rate
	}
	if cfg.Retries > 0 {
		c.Retries = cfg.Retries
	}
	if cfg.Timeout > 0 {
		c.Timeout = cfg.Timeout
	}
	return NewClient(c), nil
}

// --- inputs ---

type podcastIn struct {
	ID     string  `kit:"arg" help:"podcast ID"`
	Client *Client `kit:"inject"`
}

type episodeIn struct {
	ID     string  `kit:"arg" help:"episode ID"`
	Client *Client `kit:"inject"`
}

type episodesIn struct {
	PodcastID string  `kit:"arg" help:"podcast ID"`
	Limit     int     `kit:"flag,inherit" help:"max episodes (0 = all)"`
	Client    *Client `kit:"inject"`
}

// --- handlers ---

func getPodcast(ctx context.Context, in podcastIn, emit func(*Podcast) error) error {
	if in.ID == "" {
		return errs.Usage("podcast id is required")
	}
	p, err := in.Client.GetPodcast(ctx, in.ID)
	if err != nil {
		return err
	}
	return emit(p)
}

func getEpisode(ctx context.Context, in episodeIn, emit func(*Episode) error) error {
	if in.ID == "" {
		return errs.Usage("episode id is required")
	}
	ep, err := in.Client.GetEpisode(ctx, in.ID)
	if err != nil {
		return err
	}
	return emit(ep)
}

func listEpisodes(ctx context.Context, in episodesIn, emit func(*Episode) error) error {
	if in.PodcastID == "" {
		return errs.Usage("podcast-id is required")
	}
	eps, err := in.Client.ListEpisodes(ctx, in.PodcastID, in.Limit)
	if err != nil {
		return err
	}
	for _, ep := range eps {
		if err := emit(ep); err != nil {
			return err
		}
	}
	return nil
}
