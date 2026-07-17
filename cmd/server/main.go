package main

import (
	"fmt"
	"log/slog"

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
}
