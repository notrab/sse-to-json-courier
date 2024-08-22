package flags

import "github.com/spf13/cobra"

var (
	SourceURL string
	TargetURL string
	AuthToken string
)

func AddFlags(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&SourceURL, "source", "s", "", "Source URL for SSE events")
	cmd.Flags().StringVarP(&TargetURL, "target", "t", "", "Target URL to forward events")
	cmd.Flags().StringVarP(&AuthToken, "auth", "a", "", "Authentication token")

	cmd.MarkFlagRequired("source")
	cmd.MarkFlagRequired("target")
}
