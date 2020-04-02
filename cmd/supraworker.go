// Copyright 2020 Valeriy Soloviov. All rights reserved.
// Licensed under the Apache License, Version 2.0 (the "License");
// license that can be found in the LICENSE file.

// Package cmd provides CLI interfaces for the `supraworker` application.
package cmd

import (
	"fmt"
	"github.com/fsnotify/fsnotify"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"time"
	// "strings"
	// "bytes"
	"context"
	"github.com/sirupsen/logrus"
	// config "github.com/weldpua2008/supraworker/config"
	job "github.com/weldpua2008/supraworker/job"
	model "github.com/weldpua2008/supraworker/model"
	worker "github.com/weldpua2008/supraworker/worker"
	// "html/template"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

var (
	verbose    bool
	traceFlag  bool
	log            = logrus.WithFields(logrus.Fields{"package": "cmd"})
	numWorkers int = 5
)

func init() {

	// Define Persistent Flags and configuration settings, which, if defined here,
	// will be global for application.
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose")
	rootCmd.PersistentFlags().BoolVarP(&traceFlag, "trace", "t", false, "trace")

	rootCmd.PersistentFlags().IntVarP(&numWorkers, "workers", "w", 5, "Number of workers")
	// local flags, which will only run
	// when this action is called directly.

	logrus.SetFormatter(&logrus.TextFormatter{ForceColors: true, FullTimestamp: true})
	// Only log the warning severity or above.
	logrus.SetLevel(logrus.InfoLevel)
}

// This represents the base command when called without any subcommands
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
        log.Info(fmt.Sprintf("Starting Supraworker"))
		go func() {
			sig := <-sigs
			log.Info(fmt.Sprintf("Shutting down - got %v signal", sig))
			cancel()
			shutchan <- true
		}()

		if traceFlag {
			logrus.SetLevel(logrus.TraceLevel)
		} else if verbose {
			logrus.SetLevel(logrus.DebugLevel)
		}

		log.Trace("Config file:", viper.ConfigFileUsed())
		delay := int64(viper.GetInt("api_delay_sec"))
		if delay < 1 {
			delay = 1
		}

		api_delay_sec := time.Duration(delay) * time.Second

		// load config
		model.ReinitializeConfig()
		viper.WatchConfig()
		viper.OnConfigChange(func(e fsnotify.Event) {
			log.Trace("Config file changed:", e.Name)
			model.ReinitializeConfig()
		})

		go job.StartGenerateJobs(jobs, ctx, api_delay_sec)
		for w := 1; w <= numWorkers; w++ {
			wg.Add(1)
			go worker.StartWorker(w, jobs, &wg)
		}

		wg.Wait()
		time.Sleep(150 * time.Millisecond)

	},
}

// Execute adds all child commands to the root command sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
// return error
func Execute() error {
	if err := rootCmd.Execute(); err != nil {
		return err
	}
	return nil
}
