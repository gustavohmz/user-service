package configs

import (
	"fmt"
	"os"

	"github.com/spf13/viper"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	JWT      JWTConfig
	PLD      PLDConfig
	RabbitMQ RabbitMQConfig
}

type ServerConfig struct {
	Port string
	Host string
}

type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
}

type JWTConfig struct {
	SecretKey string
	ExpiresIn int
}

type PLDConfig struct {
	BaseURL string
	Timeout int
}

type RabbitMQConfig struct {
	URL      string
	User     string
	Password string
	Host     string
	Port     string
}

func Load() (*Config, error) {
	viper.SetConfigName(".env")
	viper.SetConfigType("env")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./configs")

	viper.AutomaticEnv()

	viper.SetDefault("SERVER_PORT", "8080")
	viper.SetDefault("SERVER_HOST", "0.0.0.0")
	viper.SetDefault("DB_HOST", "localhost")
	viper.SetDefault("DB_PORT", "5432")
	viper.SetDefault("DB_USER", "postgres")
	viper.SetDefault("DB_PASSWORD", "postgres")
	viper.SetDefault("DB_NAME", "crabi_db")
	viper.SetDefault("DB_SSLMODE", "disable")
	viper.SetDefault("JWT_SECRET_KEY", "your-secret-key-change-in-production")
	viper.SetDefault("JWT_EXPIRES_IN", 24)
	viper.SetDefault("PLD_BASE_URL", "http://98.81.235.22")
	viper.SetDefault("PLD_TIMEOUT", 10)
	viper.SetDefault("RABBITMQ_HOST", "localhost")
	viper.SetDefault("RABBITMQ_PORT", "5672")
	viper.SetDefault("RABBITMQ_USER", "guest")
	viper.SetDefault("RABBITMQ_PASSWORD", "guest")

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("error leyendo archivo de configuración: %w", err)
		}
	}

	jwtSecret := os.Getenv("JWT_SECRET_KEY")
	if jwtSecret == "" {
		jwtSecret = viper.GetString("JWT_SECRET_KEY")
	}

	config := &Config{
		Server: ServerConfig{
			Port: viper.GetString("SERVER_PORT"),
			Host: viper.GetString("SERVER_HOST"),
		},
		Database: DatabaseConfig{
			Host:     viper.GetString("DB_HOST"),
			Port:     viper.GetString("DB_PORT"),
			User:     viper.GetString("DB_USER"),
			Password: viper.GetString("DB_PASSWORD"),
			DBName:   viper.GetString("DB_NAME"),
			SSLMode:  viper.GetString("DB_SSLMODE"),
		},
		JWT: JWTConfig{
			SecretKey: jwtSecret,
			ExpiresIn: viper.GetInt("JWT_EXPIRES_IN"),
		},
		PLD: PLDConfig{
			BaseURL: viper.GetString("PLD_BASE_URL"),
			Timeout: viper.GetInt("PLD_TIMEOUT"),
		},
		RabbitMQ: RabbitMQConfig{
			Host:     viper.GetString("RABBITMQ_HOST"),
			Port:     viper.GetString("RABBITMQ_PORT"),
			User:     viper.GetString("RABBITMQ_USER"),
			Password: viper.GetString("RABBITMQ_PASSWORD"),
		},
	}

	config.RabbitMQ.URL = fmt.Sprintf("amqp://%s:%s@%s:%s/",
		config.RabbitMQ.User,
		config.RabbitMQ.Password,
		config.RabbitMQ.Host,
		config.RabbitMQ.Port,
	)

	return config, nil
}

