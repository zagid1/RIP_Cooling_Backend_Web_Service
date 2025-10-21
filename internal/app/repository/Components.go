package repository

import (
	"RIP/internal/app/ds"
	"context"
	"errors"
	"time"

	"fmt"
	"log"
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

	query := r.db.Model(&ds.Component{})
	if title != "" {
		query = query.Where("title ILIKE ?", "%"+title+"%")
	}
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	componentsQuery := query.Order("id asc")
	if err := componentsQuery.Find(&components).Error; err != nil {
		return nil, 0, err
	}

	return components, total, nil
}

// GET /api/components/:id - один компонент
func (r *Repository) GetComponentByID(id int) (*ds.Component, error) {
	var component ds.Component
	err := r.db.First(&component, id).Error
	if err != nil {
		return nil, err
	}
	return &component, nil
}

// POST /api/components - создание компонента
func (r *Repository) CreateComponent(component *ds.Component) error {
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
	// if req.Status != nil {
	// 	component.Status = *req.Status
	// }

	if err := r.db.Save(&component).Error; err != nil {
		return nil, err
	}

	return &component, nil
}

// DELETE /api/components/:id - удаление компонента
func (r *Repository) DeleteComponent(id uint) error {
	var component ds.Component
	var imageURLToDelete string

	err := r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.First(&component, id).Error; err != nil {
			return err
		}
		if component.ImageURL != nil {
			imageURLToDelete = *component.ImageURL
		}
		if err := tx.Delete(&ds.Component{}, id).Error; err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		return err
	}

	if imageURLToDelete != "" {
		parsedURL, err := url.Parse(imageURLToDelete)
		if err != nil {
			log.Printf("ERROR: could not parse image URL for deletion: %v", err)
			return nil
		}

		objectName := strings.TrimPrefix(parsedURL.Path, fmt.Sprintf("/%s/", r.bucketName))

		err = r.minioClient.RemoveObject(context.Background(), r.bucketName, objectName, minio.RemoveObjectOptions{})
		if err != nil {
			log.Printf("ERROR: failed to delete object '%s' from MinIO: %v", objectName, err)
		}
	}

	return nil
}

// POST /api/CoolRequest/draft/Components/:component_id - добавление компонента в черновик
func (r *Repository) AddComponentToDraft(userID, componentID uint) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		var CoolRequest ds.CoolRequest
		err := tx.Where("creator_id = ? AND status = ?", userID, ds.StatusDraft).First(&CoolRequest).Error
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				newCoolRequest := ds.CoolRequest{
					CreatorID:    userID,
					Status:       ds.StatusDraft,
					CreationDate: time.Now(),
				}
				if err := tx.Create(&newCoolRequest).Error; err != nil {
					return fmt.Errorf("failed to create draft coolrequest: %w", err)
				}
				CoolRequest = newCoolRequest
			} else {
				return err
			}
		}

		var count int64
		tx.Model(&ds.ComponentToRequest{}).Where("coolrequest_id = ? AND component_id = ?", CoolRequest.ID, componentID).Count(&count)
		if count > 0 {
			return errors.New("component already in coolrequest")
		}

		link := ds.ComponentToRequest{
			CoolRequestID: CoolRequest.ID,
			ComponentID:   componentID,
		}

		if err := tx.Create(&link).Error; err != nil {
			return fmt.Errorf("failed to add component to coolrequest: %w", err)
		}

		if err := tx.Model(&ds.Component{}).Where("id = ?", componentID).Update("status", true).Error; err != nil {
			return fmt.Errorf("failed to update component status: %w", err)
		}
		return nil
	})
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

		if component.ImageURL != nil && *component.ImageURL != "" {
			oldImageURL, err := url.Parse(*component.ImageURL)
			if err == nil {
				oldObjectName := strings.TrimPrefix(oldImageURL.Path, fmt.Sprintf("/%s/", r.bucketName))
				r.minioClient.RemoveObject(context.Background(), r.bucketName, oldObjectName, minio.RemoveObjectOptions{})
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

// package repository

// import (
// 	"fmt"

// 	"RIP/internal/app/ds"
// )

// func (r *Repository) GetComponents() ([]ds.Component, error) {
// 	var components []ds.Component
// 	err := r.db.Find(&components).Error
// 	// обязательно проверяем ошибки, и если они появились - передаем выше, то есть хендлеру
// 	if err != nil {
// 		return nil, err
// 	}
// 	if len(components) == 0 {
// 		return nil, fmt.Errorf("массив пустой")
// 	}

// 	return components, nil
// }

// func (r *Repository) GetComponentByID(id int) (*ds.Component, error) {
// 	component := ds.Component{}
// 	err := r.db.Where("id = ?", id).First(&component).Error
// 	if err != nil {
// 		return nil, err
// 	}
// 	return &component, nil
// }

// func (r *Repository) GetComponentsByTitle(title string) ([]ds.Component, error) {
// 	var components []ds.Component
// 	err := r.db.Where("title ILIKE ?", "%"+title+"%").Find(&components).Error
// 	if err != nil {
// 		return nil, err
// 	}
// 	return components, nil
// }
