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
	logger := logrus.New()
	router := http.NewServeMux()
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
			logger.WithError(err).Fatal("graceful shutdown failed")
		}
		close(shutdownComplete)
	}()

	logger.Info("server started on :8080")
	if err := server.ListenAndServe(); err != http.ErrServerClosed {
		logger.WithError(err).Fatal("server unable to start")
	}
	logger.Info("server exiting")

	<-shutdownComplete

	logger.Info("server gracefully shut down")
}
