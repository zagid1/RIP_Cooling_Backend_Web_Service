package api

import (
	"initial-design/internal/app/handler"
	"initial-design/internal/app/repository"
	"log"

	"github.com/sirupsen/logrus"

	"github.com/gin-gonic/gin"
)

func StartServer() {
	log.Println("Starting server")

	repo, err := repository.NewRepository()
	if err != nil {
		logrus.Error("ошибка инициализации репозитория")
	}

	handler := handler.NewHandler(repo)

	r := gin.Default()

	r.LoadHTMLGlob("templates/*")
	r.Static("/static", "./resources")

	r.GET("/CoolSystems", handler.GetComponents)
	r.GET("/Component/:id", handler.GetComponent)
	r.GET("/CoolServerTask/:id", handler.GetCoolTask)

	r.Run() // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
	log.Println("Server down")
}
