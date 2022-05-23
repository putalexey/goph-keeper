package config

import (
	"encoding/json"
	"flag"
	"github.com/caarlos0/env/v6"
	"os"
)

const DEFAULT_CONFIG_FILE = "gk-client.json"

type ClientConfig struct {
	// address of the goph-keeper server
	ServerHost string `env:"SERVER_HOST" json:"server_host"`
	// path to logfile, leave empty to log to console
	LogfilePath string `env:"LOGFILE_PATH" json:"logfile_path"`
}

type ConfigFile struct {
	// path to config file
	File string `env:"CONFIG"`
}

// Parse read config parameters. Read parameters from:
// 1. commandline arguments
// 2. environment parameters
// 3. json config file
func Parse() (*ClientConfig, error) {
	var err error
	cfg := &ClientConfig{
		ServerHost:  "goph-keeper.putalexey.ru",
		LogfilePath: "",
	}

	argFlags := parseFlags()
	configFile := os.Getenv("CONFIG")
	if file, ok := argFlags["Config"]; ok {
		configFile = file
	}
	if configFile != "" {
		err = parseConfigFile(cfg, configFile)
		if err != nil {
			return nil, err
		}
	} else {
		err = parseConfigFile(cfg, DEFAULT_CONFIG_FILE)
		if err != nil && !os.IsNotExist(err) {
			return nil, err
		}
	}

	err = env.Parse(cfg)
	if err != nil {
		return nil, err
	}

	applyArgsToConfig(cfg, argFlags)

	return cfg, nil
}

func parseConfigFile(cfg *ClientConfig, configFile string) error {
	f, err := os.Open(configFile)
	if err != nil {
		return err
	}

	d := json.NewDecoder(f)
	err = d.Decode(&cfg)
	if err != nil {
		return err
	}

	return nil
}

func parseFlags() map[string]string {
	configFileFlag := flag.String("c", "", "Файл конфига в формате JSON")
	serverHostFlag := flag.String("s", "", "Адрес сервера")
	logfilePathFlag := flag.String("l", "", "Путь к файлу лога (по-умолчанию вывод в консоль)")
	flag.Parse()

	cfg := make(map[string]string)
	if *configFileFlag != "" {
		cfg["Config"] = *configFileFlag
	}
	if *serverHostFlag != "" {
		cfg["ServerHost"] = *serverHostFlag
	}
	if *logfilePathFlag != "" {
		cfg["LogfilePath"] = *logfilePathFlag
	}
	return cfg
}

func applyArgsToConfig(config *ClientConfig, args map[string]string) {
	if value, ok := args["ServerHost"]; ok {
		config.ServerHost = value
	}
	if value, ok := args["LogfilePath"]; ok {
		config.LogfilePath = value
	}
}
