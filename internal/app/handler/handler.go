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

		// Заявки, доступно только авторизованным пользователям
		auth.POST("/cooling/draft/components/:component_id", h.AddComponentToDraft)
		auth.GET("/cooling/coolcart", h.GetCartBadge)
		auth.GET("/cooling", h.ListRequests)
		auth.GET("/cooling/:id", h.GetRequest)
		auth.PUT("/cooling/:id", h.UpdateRequest)
		auth.PUT("/cooling/:id/form", h.FormRequest)
		auth.DELETE("/cooling/:id", h.DeleteRequest)
		auth.DELETE("/cooling/:id/components/:component_id", h.RemoveComponentFromRequest)
		auth.PUT("/cooling/:id/components/:component_id", h.UpdateComponentInRequest)
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
		moderator.PUT("/cooling/:id/resolve", h.ResolveRequest)
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
