package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	// "os"
	// "strings"
    "html/template"
    "bytes"
    config "github.com/weldpua2008/supraworker/config"
    "github.com/sirupsen/logrus"

)

var (
verbose bool
log = logrus.WithFields(logrus.Fields{"package": "cmd"})

)

func init() {

  // Output to stdout instead of the default stderr
  // Can be any io.Writer, see below for File example
  // logrus.SetFormatter(&logrus.JSONFormatter{})

  // logrus.SetOutput(os.Stdout)
  logrus.SetFormatter(&logrus.TextFormatter{ForceColors: true, FullTimestamp: true, TimestampFormat: "2020-03-12 15:00:05"})
  rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose")
  // Only log the warning severity or above.
  // logrus.SetLevel(logrus.InfoLevel)
  logrus.SetLevel(logrus.InfoLevel)
}




var rootCmd = &cobra.Command{
	Use:   "supraworker",
	Short: "Supraworker is abstraction layer around jobs",
	Long: `A Fast and Flexible Abstraction around jobs built with
                love by weldpua2008 and friends in Go.
                Complete documentation is available at github.com/weldpua2008/supraworker/cmd`,
	Version: FormattedVersion(),
	Run: func(cmd *cobra.Command, args []string) {
        if verbose {
            logrus.SetLevel(logrus.DebugLevel)
        }
        log.Trace("Config file:", viper.ConfigFileUsed())

		log.Debug(viper.GetString("jobs.get.url"))
        // log.Debug(viper.GetStringMapString("jobs.get.headers"))
        log.Debug(viper.GetStringMapString("headers"))

        t:=viper.GetStringMapString("jobs.get.params")

        for k, v := range t {
            var tpl_bytes bytes.Buffer
            tpl := template.Must(template.New("params").Parse(v))
            err := tpl.Execute(&tpl_bytes, config.C)
            fmt.Println("")
    		if err != nil {
    			log.Warn("executing template:", err)
    		}
            log.Info(fmt.Sprintf("%s -> %s\n", k, tpl_bytes.String()))

        }
        // fmt.Println(t)
        // fmt.Println("---")
		// fmt.Println(config.C)
		// fmt.Println("Done")

        log.Info("Finished")

	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
