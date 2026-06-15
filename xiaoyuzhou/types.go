package xiaoyuzhou

// Podcast holds the data extracted for the podcast command.
type Podcast struct {
	ID                string `json:"id"                 kit:"id" table:"id"`
	Title             string `json:"title"                       table:"title"`
	Author            string `json:"author"                      table:"author"`
	SubscriptionCount int    `json:"subscription_count"          table:"subscribers"`
	EpisodeCount      int    `json:"episode_count"               table:"episodes"`
	Description       string `json:"description,omitempty"       table:"-"`
	URL               string `json:"url"                         table:"url,url"`
}

// Episode holds the data extracted for the episode command.
type Episode struct {
	EID          string  `json:"eid"           kit:"id" table:"eid"`
	Title        string  `json:"title"                  table:"title"`
	DurationSecs float64 `json:"duration_secs"          table:"duration"`
	PlayCount    int     `json:"play_count"             table:"plays"`
	CommentCount int     `json:"comment_count"          table:"comments"`
	ClapCount    int     `json:"clap_count"             table:"claps"`
	PubDate      string  `json:"pub_date"               table:"date"`
	PodcastTitle string  `json:"podcast_title"          table:"podcast"`
	URL          string  `json:"url"                    table:"url,url"`
}
