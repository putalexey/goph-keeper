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
	StoragePath string `env:"STORAGE_PATH" json:"storage_path"`
	//FileMode    os.FileMode `env:"FILE_MODE" json:"file_mode"`
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
	// if there is error with, files will be stored in current dir
	StoragePath := "store.json"
	LogfilePath := "gk-client.log"
	confDir, err := os.UserConfigDir()
	if err == nil {
		StoragePath = confDir + "/gk-client/store.json"
		LogfilePath = confDir + "/gk-client/gk-client.log"
	}
	cfg := &ClientConfig{
		ServerHost:  "goph-keeper.putalexey.ru:3030",
		StoragePath: StoragePath,
		LogfilePath: LogfilePath,
		//FileMode:    0600,
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
	storageFlag := flag.String("t", "", "Путь к файлу данных (~/.config/gk-client/store.json)")
	//fileModeFlag := flag.String("m", "", "Права создаваемых файлов (0600)")
	logfilePathFlag := flag.String("l", "", "Путь к файлу лога (~/.config/gk-client/gk-client.log)")
	flag.Parse()

	cfg := make(map[string]string)
	if *configFileFlag != "" {
		cfg["Config"] = *configFileFlag
	}
	if *serverHostFlag != "" {
		cfg["ServerHost"] = *serverHostFlag
	}
	if *storageFlag != "" {
		cfg["StoragePath"] = *storageFlag
	}
	//if *fileModeFlag != "" {
	//	cfg["FileMode"] = *fileModeFlag
	//}
	if *logfilePathFlag != "" {
		cfg["LogfilePath"] = *logfilePathFlag
	}
	return cfg
}

func applyArgsToConfig(config *ClientConfig, args map[string]string) {
	if value, ok := args["ServerHost"]; ok {
		config.ServerHost = value
	}
	if value, ok := args["StoragePath"]; ok {
		config.StoragePath = value
	}
	//if value, ok := args["FileMode"]; ok {
	//	i, err := strconv.ParseInt(value, 8, 8)
	//	if err != nil {
	//		log.Fatal(err)
	//	}
	//	config.FileMode = os.FileMode(i)
	//}
	if value, ok := args["LogfilePath"]; ok {
		config.LogfilePath = value
	}
}
