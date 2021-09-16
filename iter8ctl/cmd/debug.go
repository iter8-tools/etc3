package cmd

import (
	"errors"
	"fmt"
	"os"

	"github.com/iter8-tools/etc3/iter8ctl/debug"
	expr "github.com/iter8-tools/etc3/iter8ctl/experiment"
	"github.com/spf13/cobra"
)

// debugCmd represents the debug command
var debugCmd = &cobra.Command{
	Use:   "debug [experiment-name]",
	Short: "Debug an Iter8 experiment",
	Long:  `Print logs for an Iter8 experiment sorted in chronological order`,
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) > 1 {
			return errors.New("more than one positional argument supplied")
		}

		// at this stage, either latest must be true or expName must be non-empty
		latest = (len(args) == 0)
		if !latest {
			expName = args[0]
		}
		if !latest && expName == "" {
			panic("either latest must be true or expName must be non-empty")
		}

		// get experiment from cluster
		var err error
		if exp, err = expr.GetExperiment(latest, expName, expNamespace); err != nil {
			return err
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		deb, err := debug.Debug(exp)
		if err == nil {
			fmt.Print(deb)
		} else {
			fmt.Println(err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(debugCmd)
}
