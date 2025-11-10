package ds

import "time"

// Requests соответствует таблице "requests" - заявки на расчет системы охлаждения
type Cooling struct {
	ID             uint       `gorm:"primaryKey;column:id"`
	Status         int        `gorm:"column:status;not null"` // 1-черновик, 2-удален, 3-сформирован, 4-завершен, 5-отклонен
	CreationDate   time.Time  `gorm:"column:creation_date;not null"`
	CreatorID      uint       `gorm:"column:creator_id;not null"`
	FormingDate    *time.Time `gorm:"column:forming_date"`
	CompletionDate *time.Time `gorm:"column:completion_date"`
	ModeratorID    *uint      `gorm:"column:moderator_id"`

	// Поля по предметной области
	RoomArea     *float64 `gorm:"column:room_area"`     // Площадь помещения в м²
	RoomHeight   *float64 `gorm:"column:room_height"`   // Высота помещения в м
	CoolingPower *float64 `gorm:"column:cooling_power"` // Требуемая мощность охлаждения в кВт (рассчитывается)

	// --- СВЯЗИ ---
	// Отношение "принадлежит к": каждая заявка принадлежит одному пользователю
	Creator Users `gorm:"foreignKey:CreatorID"`
	// Отношение "принадлежит к": модератор (может быть NULL)
	//Moderator *Users `gorm:"foreignKey:ModeratorID"`
	Moderator *Users `gorm:"foreignKey:ModeratorID"`

	// Отношение "один-ко-многим" к связующей таблице:
	ComponentLink []ComponentToCooling `gorm:"foreignKey:CoolingID"`
}

func (Cooling) TableName() string {
	return "cooling"
}
