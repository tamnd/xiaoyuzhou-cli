package cli

import (
	"github.com/spf13/cobra"
	"github.com/tamnd/xiaoyuzhou-cli/xiaoyuzhou"
)

func (a *App) episodeCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "episode <id>",
		Short: "Show a Xiaoyuzhou episode",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id := args[0]
			a.progressf("fetching episode %s...", id)
			ep, err := a.client.Episode(cmd.Context(), id)
			if err != nil {
				return mapFetchErr(err)
			}
			return a.render([]xiaoyuzhou.Episode{*ep})
		},
	}
}
