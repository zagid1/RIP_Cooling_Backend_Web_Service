package ds

import (
	"time"

	"github.com/lib/pq"
)

// Component DTOs
type ComponentDTO struct {
	ID          uint   `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	//Specifications pq.StringArray `json:"specifications"`
	TDP      int     `json:"tdp"`
	ImageURL *string `json:"image_url"`
	Status   bool    `json:"status"`
}

type ComponentCreateRequest struct {
	Title          string         `json:"title" binding:"required"`
	Description    string         `json:"description" binding:"required"`
	Specifications pq.StringArray `json:"specifications"`
	TDP            int            `json:"tdp" binding:"required"`
}

type ComponentUpdateRequest struct {
	Title       *string `json:"title"`
	Description *string `json:"description"`
	//Specifications pq.StringArray `json:"specifications"`
	TDP *int `json:"tdp"`
	// ImageURL       *string        `json:"image_url"`
	// Status         *bool          `json:"status"`
}

// Cooling DTOs
type CoolingDTO struct {
	ID             uint                 `json:"id"`
	Status         int                  `json:"status"`
	CreationDate   time.Time            `json:"creation_date"`
	CreatorID      uint                 `json:"creator_id"`
	ModeratorID    *uint                `json:"moderator_id"`
	FormingDate    *time.Time           `json:"forming_date"`
	CompletionDate *time.Time           `json:"completion_date"`
	RoomArea       *float64             `json:"room_area"`
	RoomHeight     *float64             `json:"room_height"`
	CoolingPower   *float64             `json:"cooling_power"`
	Components     []ComponentInRequest `json:"components,omitempty"`
}

type ComponentInRequest struct {
	ComponentID uint   `json:"component_id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	//Specifications pq.StringArray `json:"specifications"`
	TDP      int     `json:"tdp"`
	ImageURL *string `json:"image_url"`
	Count    uint    `json:"count"`
}

type CoolingUpdateRequest struct {
	RoomArea   *float64 `json:"room_area"`
	RoomHeight *float64 `json:"room_height"`
}

type CoolingResolveRequest struct {
	Action string `json:"action" binding:"required"` // "complete" | "reject"
}

type ComponentToCoolingUpdateRequest struct {
	Count uint `json:"count"`
}

// Cart DTO
type CartBadgeDTO struct {
	CoolingID *uint `json:"cooling_id"`
	Count     int   `json:"count"`
}

// User DTOs
type UserRegisterRequest struct {
	FullName string `json:"full_name" binding:"required"`
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type UserLoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type UserDTO struct {
	ID        uint   `json:"id"`
	FullName  string `json:"full_name"`
	Username  string `json:"username"`
	Moderator bool   `json:"moderator"`
}

type UserUpdateRequest struct {
	FullName *string `json:"full_name"`
	Username *string `json:"username"`
	Password *string `json:"password"`
}

type LoginResponse struct {
	Token string  `json:"token"`
	User  UserDTO `json:"user"`
}

// Pagination
type PaginatedResponse struct {
	Items interface{} `json:"items"`
	Total int64       `json:"total"`
}

type AsyncCoolingRequest struct {
	ID         uint    `json:"id"`
	RoomArea   float64 `json:"room_area"`
	RoomHeight float64 `json:"room_height"`
	TotalTDP   int     `json:"total_tdp"` // <-- Новое поле вместо Components
}

type ComponentData struct {
	ID    uint `json:"id"`
	TDP   int  `json:"tdp"`
	Count int  `json:"count"`
}

// То, что Python присылает нам обратно (Webhook)
type AsyncCoolingResponse struct {
	ID           uint    `json:"id"`
	CoolingPower float64 `json:"cooling_power"`
}

type PostgresExplainResponse []struct {
	ExecutionTime float64 `json:"Execution Time"`
}

// Обновленная структура ответа с пагинацией и временем
type PaginatedCoolingResponse struct {
	Items         []CoolingDTO `json:"items"`
	Total         int64        `json:"total"`
	QueryDuration float64      `json:"query_duration_ms"` // Время выполнения в мс
	UserStats     []UserStatDTO   `json:"user_stats,omitempty"` 
}

type UserStatDTO struct {
	UserID uint  `json:"user_id"` // Исправил int на uint, так как ID обычно uint
	Count  int64 `json:"count"`
}
