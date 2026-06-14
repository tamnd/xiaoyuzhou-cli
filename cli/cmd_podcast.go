package cli

import (
	"github.com/spf13/cobra"
	"github.com/tamnd/xiaoyuzhou-cli/xiaoyuzhou"
)

func (a *App) podcastCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "podcast <id>",
		Short: "Show a Xiaoyuzhou podcast profile",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id := args[0]
			a.progressf("fetching podcast %s...", id)
			p, err := a.client.Podcast(cmd.Context(), id)
			if err != nil {
				return mapFetchErr(err)
			}
			return a.render([]xiaoyuzhou.Podcast{*p})
		},
	}
}
