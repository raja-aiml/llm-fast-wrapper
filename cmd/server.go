package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/your-org/llm-fast-wrapper/api/fiberapi"
	"github.com/your-org/llm-fast-wrapper/api/ginapi"
)

var useFiber bool
var useGin bool

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "start the API server",
	RunE: func(cmd *cobra.Command, args []string) error {
		if useFiber {
			return fiberapi.Start()
		}
		if useGin {
			return ginapi.Start()
		}
		return fmt.Errorf("no framework selected")
	},
}

func init() {
	serveCmd.Flags().BoolVar(&useFiber, "fiber", false, "use Fiber")
	serveCmd.Flags().BoolVar(&useGin, "gin", false, "use Gin")
}
