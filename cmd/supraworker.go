package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"time"
	// "strings"
	"bytes"
	"context"
	"github.com/sirupsen/logrus"
	config "github.com/weldpua2008/supraworker/config"
	job "github.com/weldpua2008/supraworker/job"
	model "github.com/weldpua2008/supraworker/model"
	worker "github.com/weldpua2008/supraworker/worker"
	"html/template"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

var (
	verbose    bool
	log            = logrus.WithFields(logrus.Fields{"package": "cmd"})
	numWorkers int = 5
)

func init() {

	// Output to stdout instead of the default stderr
	// Can be any io.Writer, see below for File example
	// logrus.SetFormatter(&logrus.JSONFormatter{})

	// logrus.SetOutput(os.Stdout)
	logrus.SetFormatter(&logrus.TextFormatter{ForceColors: true, FullTimestamp: true, TimestampFormat: "2020-03-12 15:00:05"})
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose")
	rootCmd.PersistentFlags().IntVarP(&numWorkers, "workers", "w", 5, "Number of workers")
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
		sigs := make(chan os.Signal, 1)
		shutchan := make(chan bool, 1)
		signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
		// signal.Notify(sigs, os.Interrupt)
		//    signal.Notify(sigs, os.Kill)
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel() // cancel when we are getting the kill signal or exit
		var wg sync.WaitGroup
		jobs := make(chan *model.Job, 1)

		go func() {
			sig := <-sigs
			log.Info(fmt.Sprintf("Shutting down - got %v signal", sig))
			cancel()
			shutchan <- true
		}()

		if verbose {
			logrus.SetLevel(logrus.DebugLevel)
		}
		log.Trace("Config file:", viper.ConfigFileUsed())

		log.Debug(viper.GetString("jobs.get.url"))
		// log.Debug(viper.GetStringMapString("jobs.get.headers"))
		log.Debug(viper.GetStringMapString("headers"))

		t := viper.GetStringMapString("jobs.get.params")
		delay := int64(viper.GetInt("api_delay_sec"))
		if delay < 1 {
			delay = 1
		}

		api_delay_sec := time.Duration(delay) * time.Second

		for k, v := range t {
			var tpl_bytes bytes.Buffer
			tpl := template.Must(template.New("params").Parse(v))
			err := tpl.Execute(&tpl_bytes, config.C)
			if err != nil {
				log.Warn("executing template:", err)
			}
			log.Info(fmt.Sprintf("%s -> %s\n", k, tpl_bytes.String()))
		}

		go job.StartGenerateJobs(jobs, ctx, api_delay_sec)
		for w := 1; w <= numWorkers; w++ {
			wg.Add(1)
			go worker.StartWorker(w, jobs, &wg)
		}

		// fmt.Println(t)
		// fmt.Println("---")
		// fmt.Println(config.C)
		// fmt.Println("Done")
		// Block until the WaitGroup counter goes back to 0;
		// all the workers notified they're done.
		wg.Wait()

		time.Sleep(150 * time.Millisecond)

	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
