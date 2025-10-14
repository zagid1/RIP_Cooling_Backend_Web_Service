package handler

import (
	"RIP/internal/app/config"
	"RIP/internal/app/redis"
	"RIP/internal/app/repository"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type Handler struct {
	Repository *repository.Repository
	Redis      *redis.Client
	JWTConfig  *config.JWTConfig
}

func NewHandler(r *repository.Repository, redis *redis.Client, jwtConfig *config.JWTConfig) *Handler {
	return &Handler{
		Repository: r,
		Redis:      redis,
		JWTConfig:  jwtConfig,
	}
}

func (h *Handler) RegisterAPI(r *gin.RouterGroup) {

	// Доступны всем
	r.POST("/users", h.Register)
	r.POST("/auth/login", h.Login)
	r.GET("/components", h.GetComponents)
	r.GET("/components/:id", h.GetComponent)

	// Эндпоинты, доступные только авторизованным пользователям
	auth := r.Group("/")
	auth.Use(h.AuthMiddleware)
	{
		// Пользователи
		auth.POST("/auth/logout", h.Logout)
		auth.GET("/users/:id", h.GetUserData)
		auth.PUT("/users/:id", h.UpdateUserData)

		// Заявки
		auth.POST("/coolrequest/draft/components/:component_id", h.AddComponentToDraft)
		auth.GET("/coolrequests/coolcart", h.GetCartBadge)
		auth.GET("/coolrequests", h.ListRequests)
		auth.GET("/coolrequests/:id", h.GetRequest)
		auth.PUT("/coolrequests/:id", h.UpdateRequest)
		auth.PUT("/coolrequests/:id/form", h.FormRequest)
		auth.DELETE("/coolrequests/:id", h.DeleteRequest)
		auth.DELETE("/coolrequests/:id/components/:component_id", h.RemoveComponentFromRequest)
		auth.PUT("/coolrequests/:id/components/:component_id", h.UpdateComponentInRequest)
	}

	// Эндпоинты, доступные только модераторам
	moderator := r.Group("/")
	moderator.Use(h.AuthMiddleware, h.ModeratorMiddleware)
	{
		// Управление компонентами (создание, изменение, удаление)
		moderator.POST("/components", h.CreateComponent)
		moderator.PUT("/components/:id", h.UpdateComponent)
		moderator.DELETE("/components/:id", h.DeleteComponent)
		moderator.POST("/components/:id/image", h.UploadComponentImage)

		// Управление заявками (завершение/отклонение)
		moderator.PUT("/coolrequests/:id/resolve", h.ResolveRequest)
	}
	// Домен пользователь
	// r.POST("/users", h.Register)
	// r.GET("/users/:id", h.GetUserData)
	// r.PUT("/users/:id", h.UpdateUserData)
	// r.POST("/auth/login", h.Login)
	// r.POST("/auth/logout", h.Logout)
}

func (h *Handler) errorHandler(ctx *gin.Context, errorStatusCode int, err error) {
	logrus.Error(err.Error())
	ctx.JSON(errorStatusCode, gin.H{
		"status":      "error",
		"description": err.Error(),
	})
}

// package handler

// import (
// 	"RIP/internal/app/repository"

// 	"github.com/gin-gonic/gin"
// 	"github.com/sirupsen/logrus"
// )

// type Handler struct {
// 	Repository *repository.Repository
// }

// func NewHandler(r *repository.Repository) *Handler {
// 	return &Handler{
// 		Repository: r,
// 	}
// }

// // RegisterHandler Функция, в которой мы отдельно регистрируем маршруты, чтобы не писать все в одном месте
// func (h *Handler) RegisterHandler(router *gin.Engine) {
// 	router.GET("/CoolSystems", h.GetComponents)
// 	router.GET("/Component/:id", h.GetComponentByID)
// 	router.GET("/CoolRequest/:CoolRequest_id", h.GetCoolRequest)
// 	router.POST("/CoolRequest/add/Component/:component_id", h.AddComponentToCoolRequest)
// 	router.POST("/CoolRequest/:CoolRequest_id/delete", h.DeleteCoolRequest)
// }
