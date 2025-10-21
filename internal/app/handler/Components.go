package handler

import (
	"RIP/internal/app/ds"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// GET /api/components - список компонентов с фильтрацией

// GetComponents godoc
// @Summary      Получить список компонентов (все)
// @Description  Возвращает постраничный список компонентов системы охлаждения.
// @Tags         components
// @Produce      json
// @Param        title query string false "Фильтр по названию компонента"
// @Success      200 {object} ds.PaginatedResponse
// @Failure      500 {object} map[string]string "Внутренняя ошибка сервера"
// @Router       /components [get]
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
			//Specifications: comp.Specifications,
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

// GetComponent godoc
// @Summary      Получить один компонент по ID (все)
// @Description  Возвращает детальную информацию о компоненте системы охлаждения.
// @Tags         components
// @Produce      json
// @Param        id path int true "ID компонента"
// @Success      200 {object} ds.ComponentDTO
// @Failure      404 {object} map[string]string "Компонент не найден"
// @Router       /components/{id} [get]
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
		//Specifications: component.Specifications,
		TDP:            component.TDP,
		ImageURL:       component.ImageURL,
		Status:         component.Status,
	}

	c.JSON(http.StatusOK, componentDTO)
}

// POST /api/components - создание компонента

// CreateComponent godoc
// @Summary      Создать новый компонент (только модератор)
// @Description  Создает новую запись о компоненте системы охлаждения.
// @Tags         components
// @Accept       json
// @Produce      json
// @Security     ApiKeyAuth
// @Param        componentData body ds.ComponentCreateRequest true "Данные нового компонента"
// @Success      201 {object} ds.ComponentDTO
// @Failure      400 {object} map[string]string "Ошибка валидации"
// @Failure      401 {object} map[string]string "Необходима авторизация"
// @Failure      403 {object} map[string]string "Доступ запрещен (не модератор)"
// @Router       /components [post]
func (h *Handler) CreateComponent(c *gin.Context) {
	var req ds.ComponentCreateRequest
	if err := c.BindJSON(&req); err != nil {
		h.errorHandler(c, http.StatusBadRequest, err)
		return
	}

	component := ds.Component{
		Title:          req.Title,
		Description:    req.Description,
		//Specifications: req.Specifications,
		TDP:            req.TDP,
		Status:         true, // по умолчанию активен
	}

	if err := h.Repository.CreateComponent(&component); err != nil {
		h.errorHandler(c, http.StatusInternalServerError, err)
		return
	}

	componentDTO := ds.ComponentDTO{
		ID:             component.ID,
		Title:          component.Title,
		Description:    component.Description,
		//Specifications: component.Specifications,
		TDP:            component.TDP,
		ImageURL:       component.ImageURL,
		Status:         component.Status,
	}

	c.JSON(http.StatusCreated, componentDTO)
}

// PUT /api/components/:id - обновление компонента

// UpdateComponent godoc
// @Summary      Обновить компонент (только модератор)
// @Description  Обновляет информацию о существующем компоненте системы охлаждения.
// @Tags         components
// @Accept       json
// @Produce      json
// @Security     ApiKeyAuth
// @Param        id path int true "ID компонента"
// @Param        updateData body ds.ComponentUpdateRequest true "Данные для обновления"
// @Success      200 {object} ds.ComponentDTO
// @Failure      400 {object} map[string]string "Ошибка валидации"
// @Failure      401 {object} map[string]string "Необходима авторизация"
// @Failure      403 {object} map[string]string "Доступ запрещен"
// @Router       /components/{id} [put]
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
		//Specifications: component.Specifications,
		TDP:            component.TDP,
		ImageURL:       component.ImageURL,
		Status:         component.Status,
	}

	c.JSON(http.StatusOK, componentDTO)
}

// DELETE /api/components/:id - удаление компонента

// DeleteComponent godoc
// @Summary      Удалить компонент (только модератор)
// @Description  Удаляет компонент системы охлаждения из системы.
// @Tags         components
// @Security     ApiKeyAuth
// @Param        id path int true "ID компонента для удаления"
// @Success      204 "No Content"
// @Failure      401 {object} map[string]string "Необходима авторизация"
// @Failure      403 {object} map[string]string "Доступ запрещен"
// @Router       /components/{id} [delete]
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

// POST /api/coolrequests/draft/components/:component_id - добавление компонента в черновик

// AddComponentToDraft godoc
// @Summary      Добавить компонент в черновик заявки (авторизованный пользователь)
// @Description  Находит или создает черновик заявки для текущего пользователя и добавляет в него компонент.
// @Tags         components
// @Security     ApiKeyAuth
// @Param        component_id path int true "ID компонента для добавления"
// @Success      201 {object} map[string]string "Сообщение об успехе"
// @Failure      401 {object} map[string]string "Необходима авторизация"
// @Failure      500 {object} map[string]string "Внутренняя ошибка сервера"
// @Router       /coolrequests/draft/components/{component_id} [post]
func (h *Handler) AddComponentToDraft(c *gin.Context) {
	componentID, err := strconv.Atoi(c.Param("component_id"))
	if err != nil {
		h.errorHandler(c, http.StatusBadRequest, err)
		return
	}

	userID, err := getUserIDFromContext(c)
	if err != nil {
		h.errorHandler(c, http.StatusUnauthorized, err)
		return
	}

	if err := h.Repository.AddComponentToDraft(userID, uint(componentID)); err != nil {
		h.errorHandler(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Компонент добавлен в заявку",
	})
}

// POST /api/components/:id/image - загрузка изображения компонента

// UploadComponentImage godoc
// @Summary      Загрузить изображение для компонента (только модератор)
// @Description  Загружает и привязывает изображение к компоненту системы охлаждения.
// @Tags         components
// @Accept       multipart/form-data
// @Produce      json
// @Security     ApiKeyAuth
// @Param        id path int true "ID компонента"
// @Param        file formData file true "Файл изображения"
// @Success      200 {object} map[string]string "URL загруженного изображения"
// @Failure      400 {object} map[string]string "Файл не предоставлен"
// @Failure      401 {object} map[string]string "Необходима авторизация"
// @Failure      403 {object} map[string]string "Доступ запрещен"
// @Router       /components/{id}/image [post]
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
