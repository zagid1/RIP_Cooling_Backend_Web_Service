package handler

import (
	"RIP/internal/app/ds"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// POST /api/users - регистрация пользователя

// Register godoc
// @Summary      Регистрация нового пользователя (все)
// @Description  Создает нового пользователя в системе. По умолчанию роль "пользователь", не "модератор".
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        credentials body ds.UserRegisterRequest true "Данные для регистрации"
// @Success      201 {object} ds.UserDTO
// @Failure      400 {object} map[string]string "Ошибка валидации"
// @Failure      500 {object} map[string]string "Внутренняя ошибка сервера"
// @Router       /users [post]
func (h *Handler) Register(c *gin.Context) {
	var req ds.UserRegisterRequest
	if err := c.BindJSON(&req); err != nil {
		h.errorHandler(c, http.StatusBadRequest, err)
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		h.errorHandler(c, http.StatusInternalServerError, err)
		return
	}

	user := ds.Users{
		FullName:  req.FullName,
		Username:  req.Username,
		Password:  string(hashedPassword),
		Moderator: false,
	}

	if err := h.Repository.CreateUser(&user); err != nil {
		h.errorHandler(c, http.StatusInternalServerError, err)
		return
	}

	userDTO := ds.UserDTO{
		ID:        user.ID,
		FullName:  user.FullName,
		Username:  user.Username,
		Moderator: user.Moderator,
	}

	c.JSON(http.StatusCreated, userDTO)
}

// GET /api/users/:id - получение данных пользователя

// GetUserData godoc
// @Summary      Получение данных пользователя по ID (авторизованный пользователь)
// @Description  Возвращает публичные данные пользователя. Требует авторизации.
// @Tags         users
// @Produce      json
// @Security     ApiKeyAuth
// @Param        id path int true "ID пользователя"
// @Success      200 {object} ds.UserDTO
// @Failure      401 {object} map[string]string "Необходима авторизация"
// @Failure      404 {object} map[string]string "Пользователь не найден"
// @Router       /users/{id} [get]
func (h *Handler) GetUserData(c *gin.Context) {
	idStr := c.Param("id")

	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		h.errorHandler(c, http.StatusBadRequest, err)
		return
	}

	user, err := h.Repository.GetUserByID(uint(id))
	if err != nil {
		h.errorHandler(c, http.StatusNotFound, err)
		return
	}

	userDTO := ds.UserDTO{
		ID:        user.ID,
		FullName:  user.FullName,
		Username:  user.Username,
		Moderator: user.Moderator,
	}
	c.JSON(http.StatusOK, userDTO)
}

// PUT /api/users/:id - обновление данных пользователя

// UpdateUserData godoc
// @Summary      Обновление данных пользователя (авторизованный пользователь)
// @Description  Обновляет имя пользователя или пароль. Требует авторизации.
// @Tags         users
// @Accept       json
// @Security     ApiKeyAuth
// @Param        id path int true "ID пользователя"
// @Param        updateData body ds.UserUpdateRequest true "Данные для обновления"
// @Success      204 "No Content"
// @Failure      400 {object} map[string]string "Ошибка валидации"
// @Failure      401 {object} map[string]string "Необходима авторизация"
// @Failure      500 {object} map[string]string "Внутренняя ошибка сервера"
// @Router       /users/{id} [put]
func (h *Handler) UpdateUserData(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		h.errorHandler(c, http.StatusBadRequest, err)
		return
	}

	var req ds.UserUpdateRequest
	if err := c.BindJSON(&req); err != nil {
		h.errorHandler(c, http.StatusBadRequest, err)
		return
	}

	if err := h.Repository.UpdateUser(uint(id), req); err != nil {
		h.errorHandler(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusNoContent, gin.H{
		"message": "Данные пользователя обновлены",
	})
}

// POST /api/auth/login - аутентификация

// Login godoc
// @Summary      Аутентификация пользователя (все)
// @Description  Получение JWT токена по логину и паролю для доступа к защищенным эндпоинтам.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        credentials body ds.UserLoginRequest true "Учетные данные пользователя"
// @Success      200 {object} ds.LoginResponse
// @Failure      400 {object} map[string]string "Ошибка валидации"
// @Failure      401 {object} map[string]string "Неверные учетные данные"
// @Router       /auth/login [post]
func (h *Handler) Login(c *gin.Context) {
	var req ds.UserLoginRequest
	if err := c.BindJSON(&req); err != nil {
		h.errorHandler(c, http.StatusBadRequest, err)
		return
	}

	user, err := h.Repository.GetUserByUsername(req.Username)
	if err != nil {
		h.errorHandler(c, http.StatusUnauthorized, err)
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		h.errorHandler(c, http.StatusUnauthorized, err)
		return
	}

	claims := ds.JWTClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(h.JWTConfig.ExpiresIn)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
		UserID:      user.ID,
		IsModerator: user.Moderator,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(h.JWTConfig.Secret))
	if err != nil {
		h.errorHandler(c, http.StatusInternalServerError, err)
		return
	}

	response := ds.LoginResponse{
		Token: tokenString,
		User: ds.UserDTO{
			ID:        user.ID,
			FullName:  user.FullName,
			Username:  user.Username,
			Moderator: user.Moderator,
		},
	}

	c.JSON(http.StatusOK, response)
}

// POST /api/auth/logout - деавторизация

// Logout godoc
// @Summary      Выход из системы (авторизованный пользователь)
// @Description  Добавляет текущий JWT токен в черный список, делая его недействительным. Требует авторизации.
// @Tags         auth
// @Security     ApiKeyAuth
// @Success      200 {object} map[string]string "Сообщение об успехе"
// @Failure      401 {object} map[string]string "Необходима авторизация"
// @Router       /auth/logout [post]
func (h *Handler) Logout(c *gin.Context) {
	authHeader := c.GetHeader("Authorization")
	if !strings.HasPrefix(authHeader, "Bearer ") {
		h.errorHandler(c, http.StatusBadRequest, errors.New("invalid header"))
		return
	}
	tokenStr := authHeader[len("Bearer "):]

	err := h.Redis.WriteJWTToBlacklist(c.Request.Context(), tokenStr, h.JWTConfig.ExpiresIn)
	if err != nil {
		h.errorHandler(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Деавторизация прошла успешно",
	})
}

// package handler

// import (
// 	"RIP/internal/app/ds"
// 	"net/http"
// 	"strconv"

// 	"github.com/gin-gonic/gin"
// 	"golang.org/x/crypto/bcrypt"
// )

// // POST /api/users - регистрация пользователя
// func (h *Handler) Register(c *gin.Context) {
// 	var req ds.UserRegisterRequest
// 	if err := c.BindJSON(&req); err != nil {
// 		h.errorHandler(c, http.StatusBadRequest, err)
// 		return
// 	}

// 	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
// 	if err != nil {
// 		h.errorHandler(c, http.StatusInternalServerError, err)
// 		return
// 	}

// 	user := ds.Users{
// 		Username:  req.Username,
// 		Password:  string(hashedPassword),
// 		Moderator: false,
// 	}

// 	if err := h.Repository.CreateUser(&user); err != nil {
// 		h.errorHandler(c, http.StatusInternalServerError, err)
// 		return
// 	}

// 	userDTO := ds.UserDTO{
// 		ID:        user.ID,
// 		Username:  user.Username,
// 		Moderator: user.Moderator,
// 	}

// 	c.JSON(http.StatusCreated, userDTO)
// }

// // GET /api/users/:id - получение данных пользователя
// func (h *Handler) GetUserData(c *gin.Context) {
// 	idStr := c.Param("id")

// 	id, err := strconv.ParseUint(idStr, 10, 64)
// 	if err != nil {
// 		h.errorHandler(c, http.StatusBadRequest, err)
// 		return
// 	}

// 	user, err := h.Repository.GetUserByID(uint(id))
// 	if err != nil {
// 		h.errorHandler(c, http.StatusNotFound, err)
// 		return
// 	}

// 	userDTO := ds.UserDTO{
// 		ID:        user.ID,
// 		Username:  user.Username,
// 		Moderator: user.Moderator,
// 	}
// 	c.JSON(http.StatusOK, userDTO)
// }

// // PUT /api/users/:id - обновление данных пользователя
// func (h *Handler) UpdateUserData(c *gin.Context) {
// 	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
// 	if err != nil {
// 		h.errorHandler(c, http.StatusBadRequest, err)
// 		return
// 	}

// 	var req ds.UserUpdateRequest
// 	if err := c.BindJSON(&req); err != nil {
// 		h.errorHandler(c, http.StatusBadRequest, err)
// 		return
// 	}

// 	if err := h.Repository.UpdateUser(uint(id), req); err != nil {
// 		h.errorHandler(c, http.StatusInternalServerError, err)
// 		return
// 	}
// 	c.JSON(http.StatusNoContent, gin.H{
// 		"message": "Данные пользователя обновлены",
// 	})
// }

// // POST /api/auth/login - аутентификация
// func (h *Handler) Login(c *gin.Context) {
// 	var req ds.UserLoginRequest
// 	if err := c.BindJSON(&req); err != nil {
// 		h.errorHandler(c, http.StatusBadRequest, err)
// 		return
// 	}

// 	user, err := h.Repository.GetUserByUsername(req.Username)
// 	if err != nil {
// 		h.errorHandler(c, http.StatusUnauthorized, err)
// 		return
// 	}

// 	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
// 		h.errorHandler(c, http.StatusUnauthorized, err)
// 		return
// 	}

// 	response := ds.LoginResponse{
// 		Token: "wanderful",
// 		User: ds.UserDTO{
// 			ID:        user.ID,
// 			Username:  user.Username,
// 			Moderator: user.Moderator,
// 		},
// 	}

// 	c.JSON(http.StatusOK, response)
// }

// // POST /api/auth/logout - деавторизация
// func (h *Handler) Logout(c *gin.Context) {
// 	c.JSON(http.StatusOK, gin.H{
// 		"message": "Деавторизация прошла успешно",
// 	})
// }
