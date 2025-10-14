package config

import (
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type JWTConfig struct {
	Secret    string
	ExpiresIn time.Duration
}

type RedisConfig struct {
	Host     string
	Port     int
	Password string
	User     string
}

type Config struct {
	ServiceHost string
	ServicePort int
	JWT         JWTConfig
	Redis       RedisConfig
}

func NewConfig() (*Config, error) {
	_ = godotenv.Load()
	host := os.Getenv("SERVICE_HOST")
	if host == "" {
		host = "localhost"
	}
	portStr := os.Getenv("SERVICE_PORT")
	port, _ := strconv.Atoi(portStr)
	if port == 0 {
		port = 8080
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	jwtExpiresIn, err := time.ParseDuration(os.Getenv("JWT_EXPIRES_IN"))
	if err != nil {
		jwtExpiresIn = time.Hour * 1
	}

	redisHost := os.Getenv("REDIS_HOST")
	redisPort, _ := strconv.Atoi(os.Getenv("REDIS_PORT"))
	redisPassword := os.Getenv("REDIS_PASSWORD")
	redisUser := os.Getenv("REDIS_USER")

	return &Config{
		ServiceHost: host,
		ServicePort: port,
		JWT: JWTConfig{
			Secret:    jwtSecret,
			ExpiresIn: jwtExpiresIn,
		},
		Redis: RedisConfig{
			Host:     redisHost,
			Port:     redisPort,
			Password: redisPassword,
			User:     redisUser,
		},
	}, nil
}

// package config

// import (
// 	"os"

// 	"github.com/joho/godotenv"
// 	log "github.com/sirupsen/logrus"
// 	"github.com/spf13/viper"
// )

// type Config struct {
// 	ServiceHost string
// 	ServicePort int
// }

// func NewConfig() (*Config, error) {
// 	var err error

//    configName := "config"
//    _ = godotenv.Load()
//    if os.Getenv("CONFIG_NAME") != "" {
//       configName = os.Getenv("CONFIG_NAME")
//    }

//    viper.SetConfigName(configName)
//    viper.SetConfigType("toml")
//    viper.AddConfigPath("config")
//    viper.AddConfigPath(".")
//    viper.WatchConfig()

// 	err = viper.ReadInConfig()
// 	if err != nil {
// 		return nil, err
// 	}

// 	cfg := &Config{} // создаем объект конфига
// 	err = viper.Unmarshal(cfg) // читаем информацию из файла,
// 	// конвертируем и затем кладем в нашу переменную cfg
// 	if err != nil {
// 		return nil, err
// 	}

// 	log.Info("config parsed")

// 	return cfg, nil
// }
