package main

import (
	"RIP/internal/app/config"
	"RIP/internal/app/dsn"
	"RIP/internal/app/handler"
	"RIP/internal/app/redis"
	"RIP/internal/app/repository"
	"RIP/internal/pkg"
	"context"
	"fmt"
	//"time"

	//"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// @title           API для системы CoolingSystems
// @version         1.0
// @description     API-сервер для управления заявками и компонентами серверов в системе CoolingSystems.
// @contact.name    API Support
// @contact.email   support@example.com
// @host            localhost:8080
// @BasePath        /api
// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization

func main() {
	router := gin.Default()

	// router.Use(cors.New(cors.Config{
	// 	AllowOrigins:     []string{"*"},
	// 	AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
	// 	AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
	// 	AllowCredentials: true,
	// 	MaxAge:           12 * time.Hour,
	// }))

	conf, err := config.NewConfig()
	if err != nil {
		logrus.Fatalf("error loading config: %v", err)
	}

	postgresString := dsn.FromEnv()
	fmt.Println(postgresString)

	rep, errRep := repository.New(postgresString)
	if errRep != nil {
		logrus.Fatalf("error initializing repository: %v", errRep)
	}

	redisClient, errRedis := redis.New(context.Background(), conf.Redis)
	if errRedis != nil {
		logrus.Fatalf("error initializing redis: %v", errRedis)
	}

	hand := handler.NewHandler(rep, redisClient, &conf.JWT)

	application := pkg.NewApp(conf, router, hand)
	application.RunApp()
}
