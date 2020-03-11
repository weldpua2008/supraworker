package cmd

import (
	"bytes"
	"fmt"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(versionCmd)
	rootCmd.SetVersionTemplate(`{{.Version}}{{printf "\n" }}`)
}

// The git commit that was compiled. This will be filled in by the compiler.
var GitCommit string

// The main version number
const Version = "0.1.1"

// A pre-release marker for the version
// such as "dev" (in development), "beta", "rc1", etc.
const VersionPrerelease = "dev"

func FormattedVersion() string {
	var versionString bytes.Buffer
	fmt.Fprintf(&versionString, "Supraworker v")
	fmt.Fprintf(&versionString, "%s", Version)
	if VersionPrerelease != "" {
		fmt.Fprintf(&versionString, "-%s", VersionPrerelease)

		if GitCommit != "" {
			fmt.Fprintf(&versionString, " (%s)", GitCommit)
		}
	}

	return versionString.String()
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of Supraworker",
	Long:  `All software has versions. This is Supraworker's`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(FormattedVersion())
	},
}
