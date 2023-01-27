package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/sirupsen/logrus"
)

func main() {
	logrus.SetLevel(logrus.DebugLevel)
	// create directory to hold uploads
	err := os.MkdirAll("./uploads", os.ModePerm)
	if err != nil {
		logrus.WithError(err).Fatal("unable to create the upload directory")
		return
	}

	router := http.NewServeMux()
	router.HandleFunc("/upload", upload)
	server := http.Server{
		Addr:    ":8080",
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

	logrus.Info("server started on :8080")
	if err := server.ListenAndServe(); err != http.ErrServerClosed {
		logrus.WithError(err).Fatal("server unable to start")
	}
	logrus.Info("server exiting")

	<-shutdownComplete

	logrus.Info("server gracefully shut down")
}
