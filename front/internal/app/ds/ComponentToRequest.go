package ds

// FactorToFrax соответствует таблице "FactorToFrax"
type ComponentToRequest struct {
	ID            uint `gorm:"primaryKey;column:id"`
	CoolRequestID uint `gorm:"column:coolrequest_id;not null"` // Внешний ключ к FraxSearching
	ComponentID   uint `gorm:"column:component_id;not null"`   // Внешний ключ к Factors
	Count         uint `gorm:"column:count;not null"`          // Количество компонентов

	// --- СВЯЗИ ---
	// Отношение "принадлежит к" для каждой из связанных таблиц.
	Request   CoolRequest `gorm:"foreignKey:CoolRequestID"`
	Component Component   `gorm:"foreignKey:ComponentID"`
}
