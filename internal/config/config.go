package config

import (
	"flag"
	"log"
	"os"

	"github.com/ilyakaznacheev/cleanenv"
)

type HTTPServer struct {
	Addr string `yaml:"address" env-required:"true"`
}

// env-default:"production"
type Config struct {
	Env         string `yaml:"env" env:"ENV" env-required:"true" `
	DatabaseURL string `yaml:"database_url" env:"DATABASE_URL" env-required:"true"`
	Redis       Redis  `yaml:"redis"`
	HTTPServer  `yaml:"http_server"`
}

type Redis struct {
	Addr     string `yaml:"address" env:"REDIS_ADDR" env-default:"localhost:6379"`
	Password string `yaml:"password" env:"REDIS_PASSWORD"`
	DB       int    `yaml:"db" env:"REDIS_DB" env-default:"0"`
}

func MustLoad() *Config {
	var configPath string

	configPath = os.Getenv("CONFIG_PATH")

	if configPath == "" {
		flags := flag.String("config", "", "path to the configuration file")

		flag.Parse()
		configPath = *flags

		if configPath == "" {
			log.Fatal("Config path is not set")
		}

	}

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		log.Fatalf("config file does not exist: %s", configPath)
	}

	var cfg Config

	err := cleanenv.ReadConfig(configPath, &cfg)
	if err != nil {
		log.Fatalf("can not read config file: %s", err.Error())
	}
	return &cfg
}
