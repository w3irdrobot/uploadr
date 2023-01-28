package main

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"path/filepath"
	"time"

	"github.com/sirupsen/logrus"
	flag "github.com/spf13/pflag"
)

func main() {
	f := flag.NewFlagSet("config", flag.ContinueOnError)
	f.Usage = func() {
		fmt.Println(f.FlagUsages())
		os.Exit(0)
	}
	f.StringSlice("config", []string{}, "path to one or more .toml config files")
	f.String("host", "127.0.0.1", "host to listen on")
	f.Int("port", 8080, "port to listen on")
	f.String("dir", "./uploads", "directory to upload and serve files from")
	f.String("domain", "", "domain files are served from")
	f.String("log-level", "info", "level of logs to output")
	f.Parse(os.Args[1:])

	config, err := getConfiguration(f)
	if err != nil {
		logrus.WithError(err).Fatal("error getting configuration")
	}

	// validate domain structure
	domain := config.String("domain")
	if domain == "" {
		logrus.Fatal("domain must be set")
	}
	if _, err := url.ParseRequestURI(domain); err != nil {
		logrus.WithError(err).Fatal("invalid domain")
	}

	level, err := logrus.ParseLevel(config.String("log-level"))
	if err != nil {
		logrus.WithError(err).Fatal("invalid log level")
	}
	logrus.SetLevel(level)

	// create directory to hold uploads
	dir, err := filepath.Abs(config.String("dir"))
	if err != nil {
		logrus.WithError(err).Fatal("error creating uploads directory path")
	}
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		logrus.WithError(err).Fatal("unable to create the upload directory")
	}

	router := http.NewServeMux()
	router.HandleFunc("/upload", upload(config.String("dir"), config.String("domain")))
	host := fmt.Sprintf("%s:%d", config.String("host"), config.Int("port"))
	server := http.Server{
		Addr:    host,
		Handler: router,
	}

	// create channel for waiting until interrupt received
	shutdownStart := make(chan os.Signal, 1)
	signal.Notify(shutdownStart, os.Interrupt)
	// create channel for waiting until shutdown is complete
	shutdownComplete := make(chan interface{}, 1)

	go func() {
		<-shutdownStart
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := server.Shutdown(ctx); err != nil {
			logrus.WithError(err).Fatal("graceful shutdown failed")
		}
		close(shutdownComplete)
	}()

	logrus.Infof("server started on %s", host)
	if err := server.ListenAndServe(); err != http.ErrServerClosed {
		logrus.WithError(err).Fatal("server unable to start")
	}
	logrus.Info("server exiting")

	<-shutdownComplete

	logrus.Info("server gracefully shut down")
}
