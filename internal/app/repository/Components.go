package repository

import (
	"RIP/internal/app/ds"
	"context"
	"errors"
	"time"

	"fmt"
	"mime/multipart"
	"net/url"
	"path/filepath"
	"strings"

	"github.com/minio/minio-go/v7"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// GET /api/components - список компонентов с фильтрацией
func (r *Repository) ComponentsList(title string) ([]ds.Component, int64, error) {
	var components []ds.Component
	var total int64

	// Начинаем запрос
	query := r.db.Model(&ds.Component{})

	// --- МЯГКОЕ УДАЛЕНИЕ: Фильтр ---
	// Считаем и ищем только активные компоненты
	query = query.Where("status = ?", true)

	if title != "" {
		query = query.Where("title ILIKE ?", "%"+title+"%")
	}

	// Считаем количество (с учетом фильтра по статусу и названию)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Получаем данные
	componentsQuery := query.Order("id asc")
	if err := componentsQuery.Find(&components).Error; err != nil {
		return nil, 0, err
	}

	return components, total, nil
}

// GET /api/components/:id - один компонент
func (r *Repository) GetComponentByID(id int) (*ds.Component, error) {
	var component ds.Component
	// Здесь мы просто получаем компонент как есть.
	// Хендлер сам проверит поле Status и вернет 404, если Status == false.
	// Это полезно, если в будущем админ захочет посмотреть удаленный товар по прямой ссылке.
	err := r.db.First(&component, id).Error
	if err != nil {
		return nil, err
	}
	return &component, nil
}

// POST /api/components - создание компонента
func (r *Repository) CreateComponent(component *ds.Component) error {
	// При создании компонент активен (это задается в хендлере или здесь)
	// component.Status = true // Можно раскомментировать для надежности
	return r.db.Create(component).Error
}

// PUT /api/components/:id - обновление компонента
func (r *Repository) UpdateComponent(id uint, req ds.ComponentUpdateRequest) (*ds.Component, error) {
	var component ds.Component
	if err := r.db.First(&component, id).Error; err != nil {
		return nil, err
	}

	if req.Title != nil {
		component.Title = *req.Title
	}
	if req.Description != nil {
		component.Description = *req.Description
	}
	// if req.Specifications != nil {
	// 	component.Specifications = req.Specifications
	// }
	if req.TDP != nil {
		component.TDP = *req.TDP
	}
	// if req.ImageURL != nil {
	// 	component.ImageURL = req.ImageURL
	// }
	
	// Если нужно обновлять статус через PUT (например, для восстановления), раскомментируйте:
	// if req.Status != nil {
	// 	component.Status = *req.Status
	// }

	if err := r.db.Save(&component).Error; err != nil {
		return nil, err
	}

	return &component, nil
}

// DELETE /api/components/:id - МЯГКОЕ удаление компонента
func (r *Repository) DeleteComponent(id uint) error {
	// --- МЯГКОЕ УДАЛЕНИЕ ---
	// Мы НЕ удаляем запись из БД и НЕ удаляем картинку из MinIO.
	// Мы просто меняем флаг status на false.
	
	result := r.db.Model(&ds.Component{}).Where("id = ?", id).Update("status", false)
	
	if result.Error != nil {
		return result.Error
	}
	
	// (Опционально) Проверка, что запись действительно была
	if result.RowsAffected == 0 {
		return fmt.Errorf("компонент с id %d не найден", id)
	}

	return nil
}

// POST /api/Cooling/draft/Components/:component_id - добавление компонента в черновик
func (r *Repository) AddComponentToDraft(userID, componentID uint) error {
	// 1. Ищем или создаем Корзину
	var cooling ds.Cooling
	err := r.db.Where("creator_id = ? AND status = ?", userID, ds.StatusDraft).First(&cooling).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			cooling = ds.Cooling{
				CreatorID:    userID,
				Status:       ds.StatusDraft,
				CreationDate: time.Now(),
			}
			if err := r.db.Create(&cooling).Error; err != nil {
				return fmt.Errorf("ошибка создания корзины: %w", err)
			}
		} else {
			return err
		}
	}

	// 2. Проверка на дубликаты
	var count int64
	r.db.Model(&ds.ComponentToCooling{}).
		Where("cooling_id = ? AND component_id = ?", cooling.ID, componentID).
		Count(&count)

	if count > 0 {
		return nil 
	}

	// 3. Добавляем
	link := ds.ComponentToCooling{
		CoolingID:   cooling.ID,
		ComponentID: componentID,
		Count:       1,
	}

	if err := r.db.Create(&link).Error; err != nil {
		return fmt.Errorf("ошибка записи в БД: %w", err)
	}

	return nil
}

// POST /api/components/:id/image - загрузка изображения компонента
func (r *Repository) UploadComponentImage(componentID uint, fileHeader *multipart.FileHeader) (string, error) {
	var finalImageURL string
	err := r.db.Transaction(func(tx *gorm.DB) error {
		var component ds.Component
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&component, componentID).Error; err != nil {
			return fmt.Errorf("component with id %d not found: %w", componentID, err)
		}

		const imagePathPrefix = "Components/"

		// Удаляем старую картинку, только если загружаем новую
		if component.ImageURL != nil && *component.ImageURL != "" {
			oldImageURL, err := url.Parse(*component.ImageURL)
			if err == nil {
				oldObjectName := strings.TrimPrefix(oldImageURL.Path, fmt.Sprintf("/%s/", r.bucketName))
				// Ошибки удаления старой картинки игнорируем, чтобы не ломать флоу
				_ = r.minioClient.RemoveObject(context.Background(), r.bucketName, oldObjectName, minio.RemoveObjectOptions{})
			}
		}

		fileName := filepath.Base(fileHeader.Filename)
		objectName := imagePathPrefix + fileName

		file, err := fileHeader.Open()
		if err != nil {
			return err
		}
		defer file.Close()

		_, err = r.minioClient.PutObject(context.Background(), r.bucketName, objectName, file, fileHeader.Size, minio.PutObjectOptions{
			ContentType: fileHeader.Header.Get("Content-Type"),
		})

		if err != nil {
			return fmt.Errorf("failed to upload to minio: %w", err)
		}

		imageURL := fmt.Sprintf("http://%s/%s/%s", r.minioEndpoint, r.bucketName, objectName)

		if err := tx.Model(&component).Update("image_url", imageURL).Error; err != nil {
			return fmt.Errorf("failed to update component image url in db: %w", err)
		}

		finalImageURL = imageURL
		return nil
	})
	if err != nil {
		return "", err
	}
	return finalImageURL, nil
}