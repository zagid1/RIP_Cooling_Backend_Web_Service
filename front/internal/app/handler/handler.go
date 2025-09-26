package handler

import (
	"initial-design/internal/app/repository"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type Handler struct {
	Repository *repository.Repository
}

func NewHandler(r *repository.Repository) *Handler {
	return &Handler{
		Repository: r,
	}
}

func (h *Handler) GetComponents(ctx *gin.Context) {
	var components []repository.Component
	var err error

	searchQuery := ctx.Query("query")
	if searchQuery == "" {
		components, err = h.Repository.GetComponents()
		if err != nil {
			logrus.Error(err)
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			return
		}
	} else {
		components, err = h.Repository.GetComponentsByTitle(searchQuery)
		if err != nil {
			logrus.Error(err)
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			return
		}
	}

	ctx.HTML(http.StatusOK, "mainPage.html", gin.H{
		"Components": components,
		"query":      searchQuery,
	})
}

func (h *Handler) GetComponent(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		logrus.Error("Invalid ID:", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	component, err := h.Repository.GetComponent(id)
	if err != nil {
		logrus.Error("Component not found:", err)
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Component not found"})
		return
	}

	ctx.HTML(http.StatusOK, "component.html", gin.H{
		"Component": component,
	})
}


// GetRequest - отображение статической заявки (аналог GetTask)
func (h *Handler) GetCoolTask(ctx *gin.Context) {
	idStr := ctx.Param("id")

	id, err := strconv.Atoi(idStr)
	if err != nil {
		logrus.Error("Invalid request ID:", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request ID"})
		return
	}

	request, err := h.Repository.GetRequest(id)
	if err != nil {
		logrus.Error("Request not found:", err)
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Request not found"})
		return
	}

	ctx.HTML(http.StatusOK, "task.html", gin.H{
		"request": request,
	})
}

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
