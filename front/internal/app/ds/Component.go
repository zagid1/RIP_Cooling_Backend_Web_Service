package ds

import (
	"github.com/lib/pq"
)

// Components соответствует таблице "components" - компоненты сервера
type Component struct {
	ID             uint           `gorm:"primaryKey;column:id"`
	Title          string         `gorm:"column:title;size:255;not null"`
	Description    string         `gorm:"column:description;type:text;not null"`
	Specifications pq.StringArray `gorm:"column:specifications;type:text[]"` // ← Используем pq.StringArray
	TDP            int            `gorm:"column:tdp;not null"`               // Thermal Design Power в ваттах
	ImageURL       *string        `gorm:"column:image_url;size:255"`
	Status         bool           `gorm:"column:status;not null;default:true"` // статус удален/действует

	// --- СВЯЗИ ---
	// Отношение "один-ко-многим" к связующей таблице:
	RequestLink []ComponentToRequest `gorm:"foreignKey:ComponentID"`
}
