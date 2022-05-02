package main

import (
	"fmt"
	"github.com/pkg/errors"
	"github.com/putalexey/goph-keeper/internal/client/config"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"log"
	"os"
)

var buildVersion = "N/A"
var buildDate = "N/A"

func main() {
	exit := checkArgs()
	if exit {
		return
	}
	conf, err := config.Parse()
	if err != nil {
		log.Fatalln("Failed to read config", err)
	}

	logger, logclose, err := makeLogger(conf)
	if err != nil {
		log.Fatalln("Failed to create logger", err)
	}
	defer logclose()

	logger.Info("goph-keeper client")
}

func checkArgs() bool {
	for _, arg := range os.Args[1:] {
		if arg == "-v" {
			fmt.Println("Goph-keeper client")
			fmt.Println("Build version:", buildVersion)
			fmt.Println("Build date:", buildDate)
			return true
		}
	}
	return false
}

// makeLogger creates configured logger and returns zap.SugaredLogger and func that will sync and close logger
func makeLogger(conf *config.ClientConfig) (*zap.SugaredLogger, func(), error) {
	var err error
	var logwriter zapcore.WriteSyncer
	logclose := func() {}

	if conf.LogfilePath != "" {
		logwriter, logclose, err = zap.Open(conf.LogfilePath)
		if err != nil {
			return nil, nil, errors.Wrap(err, "Failed to open logfile for writing")
		}
	}

	opts := make([]zap.Option, 0)
	if logwriter != nil {
		opts = append(opts, zap.ErrorOutput(logwriter))
	}
	//baseLogger, err := zap.NewProduction()
	baseLogger, err := zap.NewDevelopment(opts...)
	return baseLogger.Sugar(), func() {
		//goland:noinspection GoUnhandledErrorResult
		baseLogger.Sync() // flushes buffer, if any
		logclose()
	}, err
}
