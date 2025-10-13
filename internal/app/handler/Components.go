package handler

import (
	"RIP/internal/app/ds"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// GET /api/components - список компонентов с фильтрацией
func (h *Handler) GetComponents(c *gin.Context) {
	title := c.Query("title")

	components, total, err := h.Repository.ComponentsList(title)
	if err != nil {
		h.errorHandler(c, http.StatusInternalServerError, err)
		return
	}

	var componentDTOs []ds.ComponentDTO
	for _, comp := range components {
		componentDTOs = append(componentDTOs, ds.ComponentDTO{
			ID:             comp.ID,
			Title:          comp.Title,
			Description:    comp.Description,
			Specifications: comp.Specifications,
			TDP:            comp.TDP,
			ImageURL:       comp.ImageURL,
			Status:         comp.Status,
		})
	}

	c.JSON(http.StatusOK, ds.PaginatedResponse{
		Items: componentDTOs,
		Total: total,
	})
}

// GET /api/components/:id - один компонент
func (h *Handler) GetComponent(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		h.errorHandler(c, http.StatusBadRequest, err)
		return
	}

	component, err := h.Repository.GetComponentByID(id)
	if err != nil {
		h.errorHandler(c, http.StatusNotFound, err)
		return
	}

	componentDTO := ds.ComponentDTO{
		ID:             component.ID,
		Title:          component.Title,
		Description:    component.Description,
		Specifications: component.Specifications,
		TDP:            component.TDP,
		ImageURL:       component.ImageURL,
		Status:         component.Status,
	}

	c.JSON(http.StatusOK, componentDTO)
}

// POST /api/components - создание компонента
func (h *Handler) CreateComponent(c *gin.Context) {
	var req ds.ComponentCreateRequest
	if err := c.BindJSON(&req); err != nil {
		h.errorHandler(c, http.StatusBadRequest, err)
		return
	}

	statusValue := false

	component := ds.Component{
		Title:          req.Title,
		Description:    req.Description,
		Specifications: req.Specifications,
		TDP:            req.TDP,
		Status:         statusValue,
		//Status:         true, // по умолчанию активен
	}

	if err := h.Repository.CreateComponent(&component); err != nil {
		h.errorHandler(c, http.StatusInternalServerError, err)
		return
	}

	componentDTO := ds.ComponentDTO{
		ID:             component.ID,
		Title:          component.Title,
		Description:    component.Description,
		Specifications: component.Specifications,
		TDP:            component.TDP,
		ImageURL:       component.ImageURL,
		Status:         component.Status,
	}

	c.JSON(http.StatusCreated, componentDTO)
}

// PUT /api/components/:id - обновление компонента
func (h *Handler) UpdateComponent(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		h.errorHandler(c, http.StatusBadRequest, err)
		return
	}

	var req ds.ComponentUpdateRequest
	if err := c.BindJSON(&req); err != nil {
		h.errorHandler(c, http.StatusBadRequest, err)
		return
	}

	component, err := h.Repository.UpdateComponent(uint(id), req)
	if err != nil {
		h.errorHandler(c, http.StatusInternalServerError, err)
		return
	}

	componentDTO := ds.ComponentDTO{
		ID:             component.ID,
		Title:          component.Title,
		Description:    component.Description,
		Specifications: component.Specifications,
		TDP:            component.TDP,
		ImageURL:       component.ImageURL,
		Status:         component.Status,
	}

	c.JSON(http.StatusOK, componentDTO)
}

// DELETE /api/components/:id - удаление компонента
func (h *Handler) DeleteComponent(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		h.errorHandler(c, http.StatusBadRequest, err)
		return
	}

	if err := h.Repository.DeleteComponent(uint(id)); err != nil {
		h.errorHandler(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusNoContent, gin.H{
		"message": "Компонент удален",
	})
}

// POST /api/coolrequest/draft/components/:component_id - добавление компонента в черновик
func (h *Handler) AddComponentToDraft(c *gin.Context) {
	componentID, err := strconv.Atoi(c.Param("component_id"))
	if err != nil {
		h.errorHandler(c, http.StatusBadRequest, err)
		return
	}

	if err := h.Repository.AddComponentToDraft(hardcodedUserID, uint(componentID)); err != nil {
		h.errorHandler(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Черновик создан. Компонент добавлен в черновик.",
	})
}

// POST /api/components/:id/image - загрузка изображения компонента
func (h *Handler) UploadComponentImage(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		h.errorHandler(c, http.StatusBadRequest, err)
		return
	}

	file, err := c.FormFile("file")
	if err != nil {
		h.errorHandler(c, http.StatusBadRequest, err)
		return
	}

	imageURL, err := h.Repository.UploadComponentImage(uint(id), file)
	if err != nil {
		h.errorHandler(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"image_url": imageURL})
}

// package handler

// import (
// 	"net/http"
// 	"strconv"

// 	"RIP/internal/app/ds"

// 	"github.com/gin-gonic/gin"
// 	"github.com/sirupsen/logrus"
// )

// func (h *Handler) GetComponents(ctx *gin.Context) {
// 	var components []ds.Component
// 	var err error

// 	searchingComponents := ctx.Query("searchingComponents") // получаем значение из нашего поля
// 	if searchingComponents == "" {                          // если поле поиска пусто, то просто получаем из репозитория все записи
// 		components, err = h.Repository.GetComponents()
// 	} else {
// 		components, err = h.Repository.GetComponentsByTitle(searchingComponents)
// 	}

// 	if err != nil {
// 		ctx.JSON(http.StatusInternalServerError, gin.H{
// 			"error": err.Error(),
// 		})
// 		logrus.Error(err)
// 		return
// 	}

// 	// Получаем черновик заявки для отображения корзины
// 	draftRequest, err := h.Repository.GetDraftCoolRequest(hardcodedUserID)
// 	var requestID uint = 0
// 	var componentsCount int = 0

// 	if err == nil && draftRequest != nil {
// 		fullRequest, err := h.Repository.GetCoolRequestWithComponents(draftRequest.ID)
// 		if err == nil {
// 			requestID = fullRequest.ID
// 			componentsCount = len(fullRequest.ComponentLink)
// 		}
// 	}

// 	ctx.HTML(http.StatusOK, "components.html", gin.H{
// 		"components": components,
// 		"query":      searchingComponents, // передаем введенный запрос обратно на страницу
// 		"requestID":  requestID,
// 		"cartCount":  componentsCount,
// 	})
// }

// func (h *Handler) GetComponentByID(ctx *gin.Context) {
// 	idStr := ctx.Param("id") // получаем id заказа из урла (то есть из /Component/:id)
// 	// через двоеточие мы указываем параметры, которые потом сможем считать через функцию выше
// 	id, err := strconv.Atoi(idStr) // так как функция выше возвращает нам строку, нужно ее преобразовать в int
// 	if err != nil {
// 		ctx.JSON(http.StatusInternalServerError, gin.H{
// 			"error": err.Error(),
// 		})
// 		logrus.Error(err)
// 		return
// 	}

// 	component, err := h.Repository.GetComponentByID(id)
// 	if err != nil {
// 		logrus.Error(err)
// 	}

// 	ctx.HTML(http.StatusOK, "oneComponent.html", gin.H{
// 		"Component": component,
// 	})
// }
