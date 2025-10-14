package handler

import (
	"RIP/internal/app/ds"
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/golang-jwt/jwt/v5"
)

const (
	jwtPrefix    = "Bearer "
	userCtx      = "userID"
	moderatorCtx = "isModerator"
)

func (h *Handler) AuthMiddleware(c *gin.Context) {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		h.errorHandler(c, http.StatusUnauthorized, errors.New("empty auth header"))
		c.Abort()
		return
	}

	if !strings.HasPrefix(authHeader, jwtPrefix) {
		h.errorHandler(c, http.StatusUnauthorized, errors.New("invalid auth header format"))
		c.Abort()
		return
	}

	tokenStr := authHeader[len(jwtPrefix):]

	// Проверка в черном списке Redis
	err := h.Redis.CheckJWTInBlacklist(c.Request.Context(), tokenStr)
	if err == nil { // Ошибки нет -> токен найден в списке -> доступ запрещен
		h.errorHandler(c, http.StatusUnauthorized, errors.New("token is blacklisted"))
		c.Abort()
		return
	}
	if !errors.Is(err, redis.Nil) { // Если ошибка не "не найдено", а что-то другое
		h.errorHandler(c, http.StatusInternalServerError, err)
		c.Abort()
		return
	}

	// Парсинг токена
	claims := &ds.JWTClaims{}
	token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(h.JWTConfig.Secret), nil
	})

	if err != nil || !token.Valid {
		h.errorHandler(c, http.StatusUnauthorized, errors.New("invalid token"))
		c.Abort()
		return
	}

	// Сохраняем данные пользователя в контекст для дальнейшего использования
	c.Set(userCtx, claims.UserID)
	c.Set(moderatorCtx, claims.IsModerator)
	c.Next()
}

func (h *Handler) ModeratorMiddleware(c *gin.Context) {
	isModerator, exists := c.Get(moderatorCtx)
	if !exists || !isModerator.(bool) {
		h.errorHandler(c, http.StatusForbidden, errors.New("access denied: moderator rights required"))
		c.Abort()
		return
	}
	c.Next()
}
