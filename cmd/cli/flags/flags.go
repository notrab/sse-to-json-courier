package flags

import "github.com/spf13/cobra"

var (
	SourceURL string
	TargetURL string
	AuthToken string
	Port      string
)

func AddFlags(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&SourceURL, "source", "s", "", "Source URL for SSE events")
	cmd.Flags().StringVarP(&TargetURL, "target", "t", "", "Target URL to forward events")
	cmd.Flags().StringVarP(&AuthToken, "auth", "a", "", "Authentication token")
	cmd.Flags().StringVarP(&Port, "port", "p", "8080", "Port to run the server on")

	cmd.MarkFlagRequired("source")
	cmd.MarkFlagRequired("target")
}
