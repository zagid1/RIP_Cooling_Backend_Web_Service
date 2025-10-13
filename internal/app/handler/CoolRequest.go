package handler

import (
	"RIP/internal/app/ds"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// GET /api/requests/cart - иконка корзины
func (h *Handler) GetCartBadge(c *gin.Context) {
	draft, err := h.Repository.GetDraftRequest(hardcodedUserID)
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

// GET /api/requests - список заявок с фильтрацией
func (h *Handler) ListRequests(c *gin.Context) {
	status := c.Query("status")
	from := c.Query("from")
	to := c.Query("to")

	requests, err := h.Repository.RequestsListFiltered(status, from, to)
	if err != nil {
		h.errorHandler(c, http.StatusInternalServerError, err)
		return
	}
	// if requests == nil{
	// 	h.errorHandler(c, http.StatusInternalServerError, err)
	// }
	

	c.JSON(http.StatusOK, requests)
}

// GET /api/requests/:id - одна заявка с компонентами
func (h *Handler) GetRequest(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		h.errorHandler(c, http.StatusBadRequest, err)
		return
	}

	request, err := h.Repository.GetRequestWithComponents(uint(id))
	if err != nil {
		h.errorHandler(c, http.StatusNotFound, err)
		return
	}

	var components []ds.ComponentInRequest
	for _, link := range request.ComponentLink {
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
		ID:             request.ID,
		Status:         request.Status,
		CreationDate:   request.CreationDate,
		CreatorID:      request.Creator.ID,
		ModeratorID:    nil,
		FormingDate:    request.FormingDate,
		CompletionDate: request.CompletionDate,
		RoomArea:       request.RoomArea,
		RoomHeight:     request.RoomHeight,
		CoolingPower:   request.CoolingPower,
		Components:     components,
	}

	if request.ModeratorID != nil {
		requestDTO.ModeratorID = &request.Moderator.ID
	}

	c.JSON(http.StatusOK, requestDTO)
}

// PUT /api/coolrequests/:id - изменение полей заявки
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

// PUT /api/requests/:id/form - сформировать заявку
func (h *Handler) FormRequest(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		h.errorHandler(c, http.StatusBadRequest, err)
		return
	}

	if err := h.Repository.FormRequest(uint(id), hardcodedUserID); err != nil {
		h.errorHandler(c, http.StatusBadRequest, err)
		return
	}

	c.JSON(http.StatusNoContent, gin.H{
		"message": "Заявка сформирована",
	})
}

// PUT /api/requests/:id/resolve - завершить/отклонить заявку
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

	moderatorID := uint(hardcodedUserID) // В реальном приложении брать из авторизации
	if err := h.Repository.ResolveRequest(uint(id), moderatorID, req.Action); err != nil {
		h.errorHandler(c, http.StatusBadRequest, err)
		return
	}

	c.JSON(http.StatusNoContent, gin.H{
		"message": "Заявка обработана модератором",
	})
}

// DELETE /api/requests/:id - удаление заявки
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

// DELETE /api/requests/:id/components/:component_id - удаление компонента из заявки
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

// PUT /api/requests/:id/components/:component_id - изменение м-м связи
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

// 	Request, err := h.Repository.GetDraftCoolRequest(hardcodedUserID)
// 	if errors.Is(err, gorm.ErrRecordNotFound) {
// 		newRequest := ds.CoolRequest{
// 			CreatorID: hardcodedUserID,
// 			Status:    ds.StatusDraft,
// 		}
// 		if createErr := h.Repository.CreateCoolRequest(&newRequest); createErr != nil {
// 			h.errorHandler(c, http.StatusInternalServerError, createErr)
// 			return
// 		}
// 		Request = &newRequest
// 	} else if err != nil {
// 		h.errorHandler(c, http.StatusInternalServerError, err)
// 		return
// 	}

// 	if err = h.Repository.AddComponentToCoolRequest(Request.ID, uint(componentID)); err != nil {
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

// 		//h.errorHandler(c, http.StatusForbidden, errors.New("cannot access an empty cool-request page, add component first"))
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
