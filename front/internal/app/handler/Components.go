package handler

import (
	"net/http"
	"strconv"

	"RIP/internal/app/ds"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func (h *Handler) GetComponents(ctx *gin.Context) {
	var components []ds.Component
	var err error

	searchingComponents := ctx.Query("searchingComponents") // получаем значение из нашего поля
	if searchingComponents == "" {                          // если поле поиска пусто, то просто получаем из репозитория все записи
		components, err = h.Repository.GetComponents()
	} else {
		components, err = h.Repository.GetComponentsByTitle(searchingComponents)
	}

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		logrus.Error(err)
		return
	}

	// Получаем черновик заявки для отображения корзины
	draftRequest, err := h.Repository.GetDraftCoolRequest(hardcodedUserID)
	var requestID uint = 0
	var componentsCount int = 0

	if err == nil && draftRequest != nil {
		fullRequest, err := h.Repository.GetCoolRequestWithComponents(draftRequest.ID)
		if err == nil {
			requestID = fullRequest.ID
			componentsCount = len(fullRequest.ComponentLink)
		}
	}

	ctx.HTML(http.StatusOK, "components.html", gin.H{
		"components": components,
		"query":      searchingComponents, // передаем введенный запрос обратно на страницу
		"requestID":  requestID,
		"cartCount":  componentsCount,
	})
}

func (h *Handler) GetComponentByID(ctx *gin.Context) {
	idStr := ctx.Param("id") // получаем id заказа из урла (то есть из /Component/:id)
	// через двоеточие мы указываем параметры, которые потом сможем считать через функцию выше
	id, err := strconv.Atoi(idStr) // так как функция выше возвращает нам строку, нужно ее преобразовать в int
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		logrus.Error(err)
		return
	}

	component, err := h.Repository.GetComponentByID(id)
	if err != nil {
		logrus.Error(err)
	}

	ctx.HTML(http.StatusOK, "oneComponent.html", gin.H{
		"Component": component,
	})
}
