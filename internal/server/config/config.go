package config

import (
	"encoding/json"
	"flag"
	"github.com/caarlos0/env/v6"
	"os"
)

const DEFAULT_CONFIG_FILE = "gk-server.json"

type ServerConfig struct {
	// local address and port server will listen connections on
	Address string `env:"SERVER_ADDRESS" json:"server_address"`
	// path to logfile, leave empty to log to console
	LogfilePath string `env:"LOGFILE_PATH" json:"logfile_path"`
	// database DSN connection string
	DatabaseDSN string `env:"DATABASE_URI" json:"database"`
	// path to migrations dir
	MigrationsDir string `env:"DATABASE_MIGRATIONS" envDefault:"migrations" json:"migrations"`
	// key to encrypt data in DB with
	EncryptionKey string `env:"ENCRYPTION_KEY" json:"encryption_key"`
}

type ConfigFile struct {
	// path to config file
	File string `env:"CONFIG"`
}

// Parse read config parameters. Read parameters from:
// 1. commandline arguments
// 2. environment parameters
// 3. json config file
func Parse() (*ServerConfig, error) {
	var err error
	cfg := &ServerConfig{
		Address:     ":3030",
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

func parseConfigFile(cfg *ServerConfig, configFile string) error {
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
	databaseDSNFlag := flag.String("d", "", "Адрес подключения к базе данных")
	migrationsDirFlag := flag.String("m", "", "Путь до папки с миграциями базы")
	addressFlag := flag.String("a", "", "Адрес запуска HTTP-сервера")
	logfilePathFlag := flag.String("l", "", "Путь к файлу лога (по-умолчанию вывод в консоль)")
	encryptKeyFlag := flag.String("k", "", "Пароль шифрования данных записей")
	flag.Parse()

	cfg := make(map[string]string)
	if *configFileFlag != "" {
		cfg["Config"] = *configFileFlag
	}
	if *addressFlag != "" {
		cfg["Address"] = *addressFlag
	}
	if *databaseDSNFlag != "" {
		cfg["DatabaseDSN"] = *databaseDSNFlag
	}
	if *migrationsDirFlag != "" {
		cfg["MigrationsDir"] = *migrationsDirFlag
	}
	if *logfilePathFlag != "" {
		cfg["LogfilePath"] = *logfilePathFlag
	}
	if *encryptKeyFlag != "" {
		cfg["EncryptionKey"] = *encryptKeyFlag
	}
	return cfg
}

func applyArgsToConfig(config *ServerConfig, args map[string]string) {
	if value, ok := args["Address"]; ok {
		config.Address = value
	}
	if value, ok := args["DatabaseDSN"]; ok {
		config.DatabaseDSN = value
	}
	if value, ok := args["MigrationsDir"]; ok {
		config.MigrationsDir = value
	}
	if value, ok := args["LogfilePath"]; ok {
		config.LogfilePath = value
	}
	if value, ok := args["EncryptionKey"]; ok {
		config.EncryptionKey = value
	}
}
