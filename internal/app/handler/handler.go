package handler

import (
	"RIP/internal/app/repository"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

const hardcodedUserID = 1

type Handler struct {
	Repository *repository.Repository
}

func NewHandler(r *repository.Repository) *Handler {
	return &Handler{
		Repository: r,
	}
}

func (h *Handler) RegisterAPI(r *gin.RouterGroup) {
	// Домен компонентов
	r.GET("/components", h.GetComponents)
	r.GET("/components/:id", h.GetComponent)
	r.POST("/components", h.CreateComponent)
	r.PUT("/components/:id", h.UpdateComponent)
	r.DELETE("/components/:id", h.DeleteComponent)
	r.POST("/coolrequest/draft/components/:component_id", h.AddComponentToDraft)
	r.POST("/components/:id/image", h.UploadComponentImage)

	// Домен заявок
	r.GET("/coolrequests/cart", h.GetCartBadge)
	r.GET("/coolrequests", h.ListRequests)
	r.GET("/coolrequests/:id", h.GetRequest)
	r.PUT("/coolrequests/:id", h.UpdateRequest)
	r.PUT("/coolrequests/:id/form", h.FormRequest)
	r.PUT("/coolrequests/:id/resolve", h.ResolveRequest)
	r.DELETE("/coolrequests/:id", h.DeleteRequest)

	// Домен м-м (компоненты в заявке)
	r.DELETE("/coolrequests/:id/components/:component_id", h.RemoveComponentFromRequest)
	r.PUT("/coolrequests/:id/components/:component_id", h.UpdateComponentInRequest)
	//r.POST("/requests/draft/components/:component_id", h.AddComponentToDraft)

	// Домен пользователь
	r.POST("/users", h.Register)
	r.GET("/users/:id", h.GetUserData)
	r.PUT("/users/:id", h.UpdateUserData)
	r.POST("/auth/login", h.Login)
	r.POST("/auth/logout", h.Logout)
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

// // RegisterStatic То же самое, что и с маршрутами, регистрируем статику
// func (h *Handler) RegisterStatic(router *gin.Engine) {
// 	router.LoadHTMLGlob("templates/*")
// 	router.Static("/styles", "./resources/styles")
// }

// // errorHandler для более удобного вывода ошибок
// func (h *Handler) errorHandler(ctx *gin.Context, errorStatusCode int, err error) {
// 	logrus.Error(err.Error())
// 	ctx.JSON(errorStatusCode, gin.H{
// 		"status":      "error",
// 		"description": err.Error(),
// 	})
// }

// func (h *Handler) GetComponents(ctx *gin.Context) {
// 	var components []repository.Component
// 	var err error

// 	searchQuery := ctx.Query("query")
// 	if searchQuery == "" {
// 		components, err = h.Repository.GetComponents()
// 		if err != nil {
// 			logrus.Error(err)
// 			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
// 			return
// 		}
// 	} else {
// 		components, err = h.Repository.GetComponentsByTitle(searchQuery)
// 		if err != nil {
// 			logrus.Error(err)
// 			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
// 			return
// 		}
// 	}

// 	ctx.HTML(http.StatusOK, "mainPage.html", gin.H{
// 		"Components": components,
// 		"query":      searchQuery,
// 	})
// }

// func (h *Handler) GetComponent(ctx *gin.Context) {
// 	idStr := ctx.Param("id")
// 	id, err := strconv.Atoi(idStr)
// 	if err != nil {
// 		logrus.Error("Invalid ID:", err)
// 		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
// 		return
// 	}

// 	component, err := h.Repository.GetComponent(id)
// 	if err != nil {
// 		logrus.Error("Component not found:", err)
// 		ctx.JSON(http.StatusNotFound, gin.H{"error": "Component not found"})
// 		return
// 	}

// 	ctx.HTML(http.StatusOK, "component.html", gin.H{
// 		"Component": component,
// 	})
// }

// // GetRequest - отображение статической заявки (аналог GetTask)
// func (h *Handler) GetCoolTask(ctx *gin.Context) {
// 	idStr := ctx.Param("id")

// 	id, err := strconv.Atoi(idStr)
// 	if err != nil {
// 		logrus.Error("Invalid request ID:", err)
// 		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request ID"})
// 		return
// 	}

// 	request, err := h.Repository.GetRequest(id)
// 	if err != nil {
// 		logrus.Error("Request not found:", err)
// 		ctx.JSON(http.StatusNotFound, gin.H{"error": "Request not found"})
// 		return
// 	}

// 	ctx.HTML(http.StatusOK, "task.html", gin.H{
// 		"request": request,
// 	})
// }

// RequestHandler - отображение страницы заявки
// func (h *Handler) RequestHandler(ctx *gin.Context) {

// 	request, err := h.Repository.GetRequest(1)
// 	if err != nil {
// 		logrus.Error("Request not found:", err)
// 		ctx.JSON(http.StatusNotFound, gin.H{"error": "Request not found"})
// 		return
// 	}

// 	ctx.HTML(http.StatusOK, "request.html", gin.H{
// 		"request": request,
// 	})
// }
