package handler

import (
	"RIP/internal/app/ds"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// GET /api/coolrequests/coolcart - иконка корзины

// GetCartBadge godoc
// @Summary      Получить информацию для иконки корзины (авторизованный пользователь)
// @Description  Возвращает ID черновика текущего пользователя и количество компонентов в нем.
// @Tags         coolrequests
// @Produce      json
// @Security     ApiKeyAuth
// @Success      200 {object} ds.CartBadgeDTO
// @Failure      401 {object} map[string]string "Необходима авторизация"
// @Router       /coolrequests/coolcart [get]
func (h *Handler) GetCartBadge(c *gin.Context) {
	userID, err := getUserIDFromContext(c)
	if err != nil {
		h.errorHandler(c, http.StatusUnauthorized, err)
		return
	}

	draft, err := h.Repository.GetDraftRequest(userID)
	if err != nil {
		c.JSON(http.StatusOK, ds.CartBadgeDTO{
			RequestID: nil,
			Count:     0,
		})
		return
	}

	fullRequest, err := h.Repository.GetRequestWithComponents(draft.ID)
	if err != nil {
		c.JSON(http.StatusOK, ds.CartBadgeDTO{
			RequestID: nil,
			Count:     0,
		})
		return
	}

	c.JSON(http.StatusOK, ds.CartBadgeDTO{
		RequestID: &fullRequest.ID,
		Count:     len(fullRequest.ComponentLink),
	})
}

// GET /api/coolrequests - список заявок с фильтрацией

// ListRequests godoc
// @Summary      Получить список заявок (авторизованный пользователь)
// @Description  Возвращает отфильтрованный список всех сформированных заявок (кроме черновиков и удаленных).
// @Tags         coolrequests
// @Produce      json
// @Security     ApiKeyAuth
// @Param        status query int false "Фильтр по статусу заявки"
// @Param        from query string false "Фильтр по дате 'от' (формат YYYY-MM-DD)"
// @Param        to query string false "Фильтр по дате 'до' (формат YYYY-MM-DD)"
// @Success      200 {array} ds.CoolRequestDTO
// @Failure      401 {object} map[string]string "Необходима авторизация"
// @Router       /coolrequests [get]
func (h *Handler) ListRequests(c *gin.Context) {
	status := c.Query("status")
	from := c.Query("from")
	to := c.Query("to")

	requests, err := h.Repository.RequestsListFiltered(status, from, to)
	if err != nil {
		h.errorHandler(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, requests)
}

// GET /api/coolrequests/:id - одна заявка с компонентами

// GetRequest godoc
// @Summary      Получить одну заявку по ID (авторизованный пользователь)
// @Description  Возвращает полную информацию о заявке, включая привязанные компоненты.
// @Tags         coolrequests
// @Produce      json
// @Security     ApiKeyAuth
// @Param        id path int true "ID заявки"
// @Success      200 {object} ds.CoolRequestDTO
// @Failure      401 {object} map[string]string "Необходима авторизация"
// @Failure      404 {object} map[string]string "Заявка не найдена"
// @Router       /coolrequests/{id} [get]
func (h *Handler) GetRequest(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		h.errorHandler(c, http.StatusBadRequest, err)
		return
	}

	coolrequest, err := h.Repository.GetRequestWithComponents(uint(id))
	if err != nil {
		h.errorHandler(c, http.StatusNotFound, err)
		return
	}

	var components []ds.ComponentInRequest
	for _, link := range coolrequest.ComponentLink {
		components = append(components, ds.ComponentInRequest{
			ComponentID:    link.Component.ID,
			Title:          link.Component.Title,
			Description:    link.Component.Description,
			Specifications: link.Component.Specifications,
			TDP:            link.Component.TDP,
			ImageURL:       link.Component.ImageURL,
			Count:          link.Count,
		})
	}

	requestDTO := ds.CoolRequestDTO{
		ID:             coolrequest.ID,
		Status:         coolrequest.Status,
		CreationDate:   coolrequest.CreationDate,
		CreatorID:      coolrequest.Creator.ID,
		ModeratorID:    nil,
		FormingDate:    coolrequest.FormingDate,
		CompletionDate: coolrequest.CompletionDate,
		RoomArea:       coolrequest.RoomArea,
		RoomHeight:     coolrequest.RoomHeight,
		CoolingPower:   coolrequest.CoolingPower,
		Components:     components,
	}

	if coolrequest.ModeratorID != nil {
		requestDTO.ModeratorID = &coolrequest.Moderator.ID
	}

	c.JSON(http.StatusOK, requestDTO)
}

// PUT /api/coolrequests/:id - изменение полей заявки

// UpdateRequest godoc
// @Summary      Обновить данные заявки (авторизованный пользователь)
// @Description  Позволяет пользователю обновить поля своей заявки (площадь помещения, высота помещения).
// @Tags         coolrequests
// @Accept       json
// @Security     ApiKeyAuth
// @Param        id path int true "ID заявки"
// @Param        updateData body ds.CoolRequestUpdateRequest true "Данные для обновления"
// @Success      204 "No Content"
// @Failure      401 {object} map[string]string "Необходима авторизация"
// @Router       /coolrequests/{id} [put]
func (h *Handler) UpdateRequest(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		h.errorHandler(c, http.StatusBadRequest, err)
		return
	}

	var req ds.CoolRequestUpdateRequest
	if err := c.BindJSON(&req); err != nil {
		h.errorHandler(c, http.StatusBadRequest, err)
		return
	}

	if err := h.Repository.UpdateRequestUserFields(uint(id), req); err != nil {
		h.errorHandler(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusNoContent, gin.H{
		"message": "Данные заявки обновлены",
	})
}

// PUT /api/coolrequests/:id/form - сформировать заявку

// FormRequest godoc
// @Summary      Сформировать заявку (авторизованный пользователь)
// @Description  Переводит заявку из статуса "черновик" в "сформирована" и рассчитывает мощность охлаждения.
// @Tags         coolrequests
// @Security     ApiKeyAuth
// @Param        id path int true "ID заявки (черновика)"
// @Success      204 "No Content"
// @Failure      400 {object} map[string]string "Не все поля заполнены"
// @Failure      401 {object} map[string]string "Необходима авторизация"
// @Router       /coolrequests/{id}/form [put]
func (h *Handler) FormRequest(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		h.errorHandler(c, http.StatusBadRequest, err)
		return
	}

	userID, err := getUserIDFromContext(c)
	if err != nil {
		h.errorHandler(c, http.StatusUnauthorized, err)
		return
	}

	if err := h.Repository.FormRequest(uint(id), userID); err != nil {
		h.errorHandler(c, http.StatusBadRequest, err)
		return
	}

	c.JSON(http.StatusNoContent, gin.H{
		"message": "Заявка сформирована",
	})
}

// PUT /api/coolrequests/:id/resolve - завершить/отклонить заявку

// ResolveRequest godoc
// @Summary      Завершить или отклонить заявку (только модератор)
// @Description  Модератор завершает или отклоняет заявку системы охлаждения.
// @Tags         coolrequests
// @Accept       json
// @Security     ApiKeyAuth
// @Param        id path int true "ID заявки"
// @Param        action body ds.CoolRequestResolveRequest true "Действие: 'complete' или 'reject'"
// @Success      204 "No Content"
// @Failure      401 {object} map[string]string "Необходима авторизация"
// @Failure      403 {object} map[string]string "Доступ запрещен"
// @Router       /coolrequests/{id}/resolve [put]
func (h *Handler) ResolveRequest(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		h.errorHandler(c, http.StatusBadRequest, err)
		return
	}

	var req ds.CoolRequestResolveRequest
	if err := c.BindJSON(&req); err != nil {
		h.errorHandler(c, http.StatusBadRequest, err)
		return
	}

	userID, err := getUserIDFromContext(c)
	if err != nil {
		h.errorHandler(c, http.StatusUnauthorized, err)
		return
	}

	moderatorID := uint(userID)
	if err := h.Repository.ResolveRequest(uint(id), moderatorID, req.Action); err != nil {
		h.errorHandler(c, http.StatusBadRequest, err)
		return
	}

	c.JSON(http.StatusNoContent, gin.H{
		"message": "Заявка обработана модератором",
	})
}

// DELETE /api/coolrequests/:id - удаление заявки

// DeleteRequest godoc
// @Summary      Удалить заявку (авторизованный пользователь)
// @Description  Логически удаляет заявку, переводя ее в статус "удалена".
// @Tags         coolrequests
// @Security     ApiKeyAuth
// @Param        id path int true "ID заявки"
// @Success      204 "No Content"
// @Failure      401 {object} map[string]string "Необходима авторизация"
// @Router       /coolrequests/{id} [delete]
func (h *Handler) DeleteRequest(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		h.errorHandler(c, http.StatusBadRequest, err)
		return
	}

	if err := h.Repository.LogicallyDeleteRequest(uint(id)); err != nil {
		h.errorHandler(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusNoContent, gin.H{
		"message": "Заявка удалена",
	})
}

// DELETE /api/coolrequests/:id/components/:component_id - удаление компонента из заявки

// RemoveComponentFromRequest godoc
// @Summary      Удалить компонент из заявки (авторизованный пользователь)
// @Description  Удаляет связь между заявкой и компонентом.
// @Tags         m-m
// @Security     ApiKeyAuth
// @Param        id path int true "ID заявки"
// @Param        component_id path int true "ID компонента"
// @Success      204 "No Content"
// @Failure      401 {object} map[string]string "Необходима авторизация"
// @Router       /coolrequests/{id}/components/{component_id} [delete]
func (h *Handler) RemoveComponentFromRequest(c *gin.Context) {
	requestID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		h.errorHandler(c, http.StatusBadRequest, err)
		return
	}

	componentID, err := strconv.Atoi(c.Param("component_id"))
	if err != nil {
		h.errorHandler(c, http.StatusBadRequest, err)
		return
	}

	if err := h.Repository.RemoveComponentFromRequest(uint(requestID), uint(componentID)); err != nil {
		h.errorHandler(c, http.StatusBadRequest, err)
		return
	}

	c.JSON(http.StatusNoContent, gin.H{
		"message": "Компонент удален из заявки",
	})
}

// PUT /api/coolrequests/:id/components/:component_id - изменение м-м связи

// UpdateComponentInRequest godoc
// @Summary      Обновить количество компонента в заявке (авторизованный пользователь)
// @Description  Изменяет количество конкретного компонента в рамках одной заявки.
// @Tags         m-m
// @Accept       json
// @Security     ApiKeyAuth
// @Param        id path int true "ID заявки"
// @Param        component_id path int true "ID компонента"
// @Param        updateData body ds.ComponentToRequestUpdateRequest true "Новое количество"
// @Success      204 "No Content"
// @Failure      401 {object} map[string]string "Необходима авторизация"
// @Router       /coolrequests/{id}/components/{component_id} [put]
func (h *Handler) UpdateComponentInRequest(c *gin.Context) {
	requestID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		h.errorHandler(c, http.StatusBadRequest, err)
		return
	}

	componentID, err := strconv.Atoi(c.Param("component_id"))
	if err != nil {
		h.errorHandler(c, http.StatusBadRequest, err)
		return
	}

	var req ds.ComponentToRequestUpdateRequest
	if err := c.BindJSON(&req); err != nil {
		h.errorHandler(c, http.StatusBadRequest, err)
		return
	}

	updateData := ds.ComponentToRequest{
		Count: req.Count,
	}

	if err := h.Repository.UpdateMM(uint(requestID), uint(componentID), updateData); err != nil {
		h.errorHandler(c, http.StatusBadRequest, err)
		return
	}

	c.JSON(http.StatusNoContent, gin.H{
		"message": "Количество компонента обновлено",
	})
}

// POST /api/requests/draft/components/:component_id - добавление компонента в черновик
// func (h *Handler) AddComponentToDraft(c *gin.Context) {
// 	componentID, err := strconv.Atoi(c.Param("component_id"))
// 	if err != nil {
// 		h.errorHandler(c, http.StatusBadRequest, err)
// 		return
// 	}

// 	if err := h.Repository.AddComponentToDraft(hardcodedUserID, uint(componentID)); err != nil {
// 		h.errorHandler(c, http.StatusInternalServerError, err)
// 		return
// 	}

// 	c.JSON(http.StatusCreated, gin.H{
// 		"message": "Компонент добавлен в заявку",
// 	})
// }

// package handler

// import (
// 	"RIP/internal/app/ds"
// 	"errors"
// 	"net/http"
// 	"strconv"

// 	"github.com/gin-gonic/gin"
// 	"gorm.io/gorm"
// )

// func (h *Handler) AddComponentToCoolRequest(c *gin.Context) {
// 	componentID, err := strconv.Atoi(c.Param("component_id"))
// 	if err != nil {
// 		h.errorHandler(c, http.StatusBadRequest, err)
// 		return
// 	}

// 	coolrequest, err := h.Repository.GetDraftCoolRequest(hardcodedUserID)
// 	if errors.Is(err, gorm.ErrRecordNotFound) {
// 		newRequest := ds.CoolRequest{
// 			CreatorID: hardcodedUserID,
// 			Status:    ds.StatusDraft,
// 		}
// 		if createErr := h.Repository.CreateCoolRequest(&newRequest); createErr != nil {
// 			h.errorHandler(c, http.StatusInternalServerError, createErr)
// 			return
// 		}
// 		coolrequest = &newRequest
// 	} else if err != nil {
// 		h.errorHandler(c, http.StatusInternalServerError, err)
// 		return
// 	}

// 	if err = h.Repository.AddComponentToCoolRequest(coolrequest.ID, uint(componentID)); err != nil {
// 	}

// 	c.Redirect(http.StatusFound, "/CoolSystems")
// }

// func (h *Handler) GetCoolRequest(c *gin.Context) {
// 	CoolRequestID, err := strconv.Atoi(c.Param("CoolRequest_id"))
// 	if err != nil {
// 		h.errorHandler(c, http.StatusBadRequest, err)
// 		return
// 	}

// 	CoolRequest, err := h.Repository.GetCoolRequestWithComponents(uint(CoolRequestID))
// 	if err != nil {
// 		var deletedRequest ds.CoolRequest
// 		c.HTML(http.StatusOK, "coolRequest.html", gin.H{
// 			"CoolRequest": deletedRequest,
// 			"Error":       err,
// 		})
// 		//h.errorHandler(c, http.StatusNotFound, err)
// 		return
// 	}

// 	if len(CoolRequest.ComponentLink) == 0 {
// 		//CoolRequestCount := len(CoolRequest.ComponentLink)
// 		c.HTML(http.StatusOK, "coolRequest.html", CoolRequest)

// 		//h.errorHandler(c, http.StatusForbidden, errors.New("cannot access an empty cool-coolrequest page, add component first"))
// 		return
// 	}

// 	// c.HTML(http.StatusOK, "coolRequest.html", H.gin{
// 	// 	CoolRequest
// 	// }
// 	// )
// 	c.HTML(http.StatusOK, "coolRequest.html", CoolRequest)
// }

// func (h *Handler) DeleteCoolRequest(c *gin.Context) {
// 	CoolRequestID, _ := strconv.Atoi(c.Param("CoolRequest_id"))

// 	if err := h.Repository.LogicallyDeleteCoolRequest(uint(CoolRequestID)); err != nil {
// 		h.errorHandler(c, http.StatusInternalServerError, err)
// 		return
// 	}

// 	c.Redirect(http.StatusFound, "/CoolSystems")
// }
