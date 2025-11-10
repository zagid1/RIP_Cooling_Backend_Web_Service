package ds

type ComponentToCooling struct {
	ID          uint `gorm:"primaryKey;column:id"`
	CoolingID   uint `gorm:"column:cooling_id;not null"`
	ComponentID uint `gorm:"column:component_id;not null"`
	Count       uint `gorm:"column:count;not null"` // Количество компонентов

	// --- СВЯЗИ ---
	// Отношение "принадлежит к" для каждой из связанных таблиц.
	Request   Cooling   `gorm:"foreignKey:CoolingID"`
	Component Component `gorm:"foreignKey:ComponentID"`
}

func (ComponentToCooling) TableName() string {
	return "component_to_cooling"
}
