package ds

// Users соответствует таблице "Users".
type Users struct {
	ID        uint   `gorm:"primaryKey;column:id"`
	FullName  string `gorm:"column:full_name;size:255;not null"`
	Username  string `gorm:"column:username;size:255;not null"`
	Password  string `gorm:"column:password;size:255;not null"`
	Moderator bool   `gorm:"column:moderator;not null"`

	// --- СВЯЗИ ---
	// Отношение "один-ко-многим": один пользователь может иметь много поисковых сессий.
	CoolRequest []CoolRequest `gorm:"foreignKey:CreatorID"`
}
