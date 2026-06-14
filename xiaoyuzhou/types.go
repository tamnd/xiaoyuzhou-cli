package xiaoyuzhou

// Podcast holds the data extracted for the podcast command.
type Podcast struct {
	Title             string  `json:"title"`
	Author            string  `json:"author"`
	SubscriptionCount float64 `json:"subscription_count"`
	EpisodeCount      float64 `json:"episode_count"`
	Description       string  `json:"description"`
	URL               string  `json:"url"`
}

// Episode holds the data extracted for the episode command.
type Episode struct {
	Title        string  `json:"title"`
	PodcastTitle string  `json:"podcast"`
	Duration     string  `json:"duration"`
	PlayCount    float64 `json:"play_count"`
	CommentCount float64 `json:"comment_count"`
	ClapCount    float64 `json:"clap_count"`
	URL          string  `json:"url"`
}

// EpisodeSummary holds one row for the episodes list command.
type EpisodeSummary struct {
	EID       string  `json:"eid"`
	Title     string  `json:"title"`
	Duration  string  `json:"duration"`
	PlayCount float64 `json:"play_count"`
	PubDate   string  `json:"pub_date"`
	URL       string  `json:"url"`
}
