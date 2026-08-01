package main

import (
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/Harichandra-Prasath/Kaekudha/internal/logger"
	"github.com/Harichandra-Prasath/Kaekudha/internal/pipeline"
)

func main() {
	logger.InitialiseLogger()
	slog.Info("Logger Initialised")

	server, err := pipeline.NewServer("0.0.0.0:9000")
	if err != nil {
		panic(fmt.Sprintf("error getting new server: %v", err))
	}
	slog.Info("Listening for data....")
	server.Start()
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	sig := <-sigChan
	slog.Info("End Signal Recieved", "Signal", sig.String())
	server.Stop()
}
