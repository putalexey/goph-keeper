package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/pkg/errors"
	"github.com/putalexey/goph-keeper/internal/client"
	"github.com/putalexey/goph-keeper/internal/client/config"
	"go.uber.org/zap"
	"log"
	"os"
	"path"
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

	fmt.Printf("goph-keeper client v%s\n", buildVersion)

	c, err := client.NewClient(context.Background(), logger, conf)
	if err != nil {
		log.Fatalln("Failed to create logger", err)
	}
	defer c.Close()

	c.ProcessCommand(context.Background(), flag.Args())
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
	outputPaths := []string{"stderr"}
	errorOutputPaths := []string{"stderr"}

	if conf.LogfilePath != "" {
		dir := path.Dir(conf.LogfilePath)
		err = os.MkdirAll(dir, 0700)
		if err != nil {
			return nil, nil, errors.Wrap(err, "Failed to create folder for logfile")
		}

		outputPaths = []string{conf.LogfilePath}
		errorOutputPaths = []string{conf.LogfilePath}
	}

	//cfg := zap.NewProductionConfig()
	cfg := zap.NewDevelopmentConfig()
	cfg.OutputPaths = outputPaths
	cfg.ErrorOutputPaths = errorOutputPaths

	//baseLogger, err := zap.NewProduction()
	baseLogger, err := cfg.Build()
	return baseLogger.Sugar(), func() {
		//goland:noinspection GoUnhandledErrorResult
		baseLogger.Sync() // flushes buffer, if any
	}, err
}
