package config

import (
	"fmt"
	"github.com/joho/godotenv"
	"os"
)

type Config struct {
	ServerAddress string
	PostgresConn  string
}

func LoadConfig() (*Config, error) {
	err := godotenv.Load()
	if err != nil {
		fmt.Println("Ошибка загрузки файла .env")
	}

	config := &Config{
		ServerAddress: os.Getenv("SERVER_ADDRESS"),
		PostgresConn:  os.Getenv("POSTGRES_CONN"),
	}

	return config, nil
}
