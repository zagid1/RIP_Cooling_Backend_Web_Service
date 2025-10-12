package repository

import (
	"fmt"

	"RIP/internal/app/ds"
)

func (r *Repository) GetComponents() ([]ds.Component, error) {
	var components []ds.Component
	err := r.db.Find(&components).Error
	// обязательно проверяем ошибки, и если они появились - передаем выше, то есть хендлеру
	if err != nil {
		return nil, err
	}
	if len(components) == 0 {
		return nil, fmt.Errorf("массив пустой")
	}

	return components, nil
}

func (r *Repository) GetComponentByID(id int) (*ds.Component, error) {
	component := ds.Component{}
	err := r.db.Where("id = ?", id).First(&component).Error
	if err != nil {
		return nil, err
	}
	return &component, nil
}

func (r *Repository) GetComponentsByTitle(title string) ([]ds.Component, error) {
	var components []ds.Component
	err := r.db.Where("title ILIKE ?", "%"+title+"%").Find(&components).Error
	if err != nil {
		return nil, err
	}
	return components, nil
}
