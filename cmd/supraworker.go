package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"os"
	// "strings"
    "html/template"
    "bytes"
    config "github.com/weldpua2008/supraworker/config"

)

var rootCmd = &cobra.Command{
	Use:   "supraworker",
	Short: "Supraworker is abstraction layer around jobs",
	Long: `A Fast and Flexible Abstraction around jobs built with
                love by weldpua2008 and friends in Go.
                Complete documentation is available at github.com/weldpua2008/supraworker/cmd`,
	Version: FormattedVersion(),
	Run: func(cmd *cobra.Command, args []string) {

		fmt.Println(viper.GetString("jobs.get.url"))
        // fmt.Println(viper.GetStringMapString("jobs.get.headers"))
        fmt.Println(viper.GetStringMapString("headers"))
        // fmt.Printf("%T\n", viper)
        t:=viper.GetStringMapString("jobs.get.params")

        for k, v := range t {
            var tpl_bytes bytes.Buffer
            tpl := template.Must(template.New("params").Parse(v))
            err := tpl.Execute(&tpl_bytes, config.C)
            fmt.Println("")
    		if err != nil {
    			fmt.Println("executing template:", err)
    		}
            fmt.Printf("%s -> %s\n", k, tpl_bytes.String())

        }
fmt.Println(t)
fmt.Println("---")
		fmt.Println(config.C)
		fmt.Println("Done")

	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
