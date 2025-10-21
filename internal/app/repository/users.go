package repository

import (
	"RIP/internal/app/ds"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

// POST /api/users - регистрация пользователя
func (r *Repository) CreateUser(user *ds.Users) error {
	// Проверяем существование пользователя с таким username
	var count int64
	if err := r.db.Model(&ds.Users{}).Where("username = ?", user.Username).Count(&count).Error; err != nil {
		return fmt.Errorf("%w: %v", ds.ErrDatabase, err)
	}

	if count > 0 {
		return ds.ErrUsernameExists
	}

	// Создаем пользователя
	if err := r.db.Create(user).Error; err != nil {
		// // Дополнительная проверка на случай race condition
		// if errors.Is(err, gorm.ErrDuplicatedKey) ||
		//    errors.Is(err, gorm.ErrUniqueViolated) {
		// 	return ds.ErrUsernameExists
		// }
		return fmt.Errorf("%w: %v", ds.ErrDatabase, err)
	}

	return nil
}

// GET /api/users/:id - получение данных пользователя
func (r *Repository) GetUserByID(id uint) (*ds.Users, error) {
	var user ds.Users
	if err := r.db.First(&user, id).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// PUT /api/users/:id - обновление данных пользователя
func (r *Repository) UpdateUser(id uint, req ds.UserUpdateRequest) error {
	updates := make(map[string]interface{})

	if req.Username != nil {
		updates["username"] = *req.Username
	}
	if req.Password != nil {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(*req.Password), bcrypt.DefaultCost)
		if err != nil {
			return err
		}
		updates["password"] = string(hashedPassword)
	}

	if len(updates) == 0 {
		return nil
	}

	return r.db.Model(&ds.Users{}).Where("id = ?", id).Updates(updates).Error
}

// POST /api/auth/login - аутентификация
func (r *Repository) GetUserByUsername(username string) (*ds.Users, error) {
	var user ds.Users
	if err := r.db.Where("username = ?", username).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}
