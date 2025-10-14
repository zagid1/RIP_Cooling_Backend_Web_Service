package pkg

import (
	"fmt"

	"RIP/internal/app/config"
	"RIP/internal/app/handler"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	_ "RIP/docs"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

type Application struct {
	Config  *config.Config
	Router  *gin.Engine
	Handler *handler.Handler
}

func NewApp(c *config.Config, r *gin.Engine, h *handler.Handler) *Application {
	return &Application{
		Config:  c,
		Router:  r,
		Handler: h,
	}
}

func (a *Application) RunApp() {
	logrus.Info("Server start up")

	a.Router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Создаем группу маршрутов /api
	api := a.Router.Group("/api")
	a.Handler.RegisterAPI(api)

	serverAddress := fmt.Sprintf("%s:%d", a.Config.ServiceHost, a.Config.ServicePort)
	if err := a.Router.Run(serverAddress); err != nil {
		logrus.Fatal(err)
	}
	logrus.Info("Server down")
}

// package pkg

// import (
//    "fmt"

//    "github.com/gin-gonic/gin"
//    "github.com/sirupsen/logrus"
//    "RIP/internal/app/config"
//    "RIP/internal/app/handler"
// )

// type Application struct {
//    Config  *config.Config
//    Router  *gin.Engine
//    Handler *handler.Handler
// }

// func NewApp(c *config.Config, r *gin.Engine, h *handler.Handler) *Application {
//    return &Application{
//       Config:  c,
//       Router:  r,
//       Handler: h,
//    }
// }

// func (a *Application) RunApp() {
//    logrus.Info("Server start up")

//    a.Handler.RegisterHandler(a.Router)
//    a.Handler.RegisterStatic(a.Router)

//    serverAddress := fmt.Sprintf("%s:%d", a.Config.ServiceHost, a.Config.ServicePort)
//    if err := a.Router.Run(serverAddress); err != nil {
//       logrus.Fatal(err)
//    }
//    logrus.Info("Server down")
// }
