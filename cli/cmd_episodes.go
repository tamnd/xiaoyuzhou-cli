package cli

import (
	"github.com/spf13/cobra"
)

func (a *App) episodesCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "episodes <podcast-id>",
		Short: "List recent episodes of a Xiaoyuzhou podcast",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id := args[0]
			lim := a.effectiveLimit(15)
			a.progressf("fetching episodes for podcast %s...", id)
			eps, err := a.client.Episodes(cmd.Context(), id)
			if err != nil {
				return mapFetchErr(err)
			}
			if lim > 0 && lim < len(eps) {
				eps = eps[:lim]
			}
			return a.renderOrEmpty(eps, len(eps))
		},
	}
}
