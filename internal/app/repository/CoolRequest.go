package repository

import (
	"RIP/internal/app/ds"
	"errors"

	//"fmt"
	"strconv"
	"time"

	"gorm.io/gorm"
)

// GET /api/requests/cart - иконка корзины
func (r *Repository) GetDraftRequest(userID uint) (*ds.CoolRequest, error) {
	var coolrequest ds.CoolRequest
	err := r.db.Where("creator_id = ? AND status = ?", userID, ds.StatusDraft).First(&coolrequest).Error
	if err != nil {
		return nil, err
	}
	return &coolrequest, nil
}

// GET /api/requests/:id - одна заявка с компонентами
func (r *Repository) GetRequestWithComponents(requestID uint, UserID uint, isModerator bool) (*ds.CoolRequest, error) {
	var CoolRequest ds.CoolRequest
	var err error

	// Базовый запрос с предзагрузкой связей
	query := r.db.Preload("ComponentLink.Component").Preload("Creator").Preload("Moderator")

	if !isModerator {
		// Для обычных пользователей - только их запросы
		err = query.Where("id = ? AND creator_id = ?", requestID, UserID).First(&CoolRequest).Error
	} else {
		// Для модераторов - любой запрос
		err = query.First(&CoolRequest, requestID).Error
	}
	if err != nil {
		return nil, err
	}

	if CoolRequest.Status == ds.StatusDeleted {
		return nil, errors.New("coolrequest page not found or has been deleted")
	}
	return &CoolRequest, nil
}

// GET /api/requests - список заявок с фильтрацией
func (r *Repository) RequestsListFiltered(userID uint, isModerator bool, status, from, to string) ([]ds.CoolRequestDTO, error) {
	var requestLIst []ds.CoolRequest
	query := r.db.Preload("Creator").Preload("Moderator")

	query = query.Where("status != ? AND status != ?", ds.StatusDeleted, ds.StatusDraft)

	if !isModerator {
		query = query.Where("creator_id = ?", userID)
	}

	if status != "" {
		if statusInt, err := strconv.Atoi(status); err == nil {
			query = query.Where("status = ?", statusInt)
		}
	}

	if from != "" {
		if fromTime, err := time.Parse("2006-01-02", from); err == nil {
			query = query.Where("forming_date >= ?", fromTime)
		}
	}

	if to != "" {
		if toTime, err := time.Parse("2006-01-02", to); err == nil {
			query = query.Where("forming_date <= ?", toTime)
		}
	}

	if err := query.Find(&requestLIst).Error; err != nil {
		return nil, err
	}

	var result []ds.CoolRequestDTO
	for _, coolrequest := range requestLIst {
		dto := ds.CoolRequestDTO{
			ID:             coolrequest.ID,
			Status:         coolrequest.Status,
			CreationDate:   coolrequest.CreationDate,
			CreatorID:      coolrequest.Creator.ID,
			ModeratorID:    nil,
			FormingDate:    coolrequest.FormingDate,
			CompletionDate: coolrequest.CompletionDate,
			RoomArea:       coolrequest.RoomArea,
			RoomHeight:     coolrequest.RoomHeight,
			CoolingPower:   coolrequest.CoolingPower,
		}

		if coolrequest.ModeratorID != nil {
			dto.ModeratorID = &coolrequest.Moderator.ID
		}
		result = append(result, dto)
	}
	return result, nil
}

// PUT /api/requests/:id - изменение полей заявки
func (r *Repository) UpdateRequestUserFields(id uint, req ds.CoolRequestUpdateRequest) error {
	updates := make(map[string]interface{})

	if req.RoomArea != nil {
		updates["room_area"] = *req.RoomArea
	}
	if req.RoomHeight != nil {
		updates["room_height"] = *req.RoomHeight
	}

	if len(updates) == 0 {
		return nil
	}

	return r.db.Model(&ds.CoolRequest{}).Where("id = ?", id).Updates(updates).Error
}

// PUT /api/requests/:id/form - сформировать заявку
func (r *Repository) FormRequest(id uint, creatorID uint) error {
	var coolrequest ds.CoolRequest
	// if err := r.db.Preload("ComponentLink.Component").First(&coolrequest, id).Error; err != nil {
	// 	return err
	// }
	if err := r.db.First(&coolrequest, id).Error; err != nil {
		return err
	}

	if coolrequest.CreatorID != creatorID {
		return errors.New("only creator can form coolrequest")
	}

	if coolrequest.Status != ds.StatusDraft {
		return errors.New("only draft coolrequest can be formed")
	}

	if coolrequest.RoomArea == nil || coolrequest.RoomHeight == nil {
		return errors.New("room area and height are required")
	}

	// Расчет мощности охлаждения на основе компонентов
	// totalTDP := 0
	// for _, link := range coolrequest.ComponentLink {
	// 	totalTDP += link.Component.TDP * int(link.Count)
	// }

	//coolingPower := float64(totalTDP)*1.2 + (*coolrequest.RoomArea * *coolrequest.RoomHeight * 0.1)

	now := time.Now()
	return r.db.Model(&coolrequest).Updates(map[string]interface{}{
		"status":       ds.StatusFormed,
		"forming_date": now,
		//	"cooling_power": coolingPower,
	}).Error
}

// PUT /api/requests/:id/resolve - завершить/отклонить заявку
func (r *Repository) ResolveRequest(id uint, moderatorID uint, action string) error {
	return r.db.Transaction(func(tx *gorm.DB) error {

		var coolrequest ds.CoolRequest
		if err := tx.Preload("ComponentLink.Component").First(&coolrequest, id).Error; err != nil {
			return err
		}

		if coolrequest.Status != ds.StatusFormed {
			return errors.New("only formed coolrequest can be resolved")
		}

		now := time.Now()
		updates := map[string]interface{}{
			"moderator_id":    moderatorID,
			"completion_date": now,
		}

		switch action {
		case "complete":
			{
				updates["status"] = ds.StatusCompleted
				coolingPower := r.calculateCoolingPower(coolrequest)
				updates["cooling_power"] = coolingPower
			}
		case "reject":
			{
				updates["status"] = ds.StatusRejected
			}
		default:
			{
				return errors.New("invalid action, must be 'complete' or 'reject'")
			}
		}

		var CoolRequestIDs []uint
		for _, link := range coolrequest.ComponentLink {
			CoolRequestIDs = append(CoolRequestIDs, link.ComponentID)
		}

		if len(CoolRequestIDs) > 0 {
			if err := tx.Model(&ds.Component{}).Where("id IN ?", CoolRequestIDs).Update("status", false).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// Функция расчета мощности охлаждения по формуле: Q = P × (1.3 + 15/V) [кВт]
func (r *Repository) calculateCoolingPower(coolrequest ds.CoolRequest) float64 {
	// P = Σ(TDP компонентов) / 1000 - тепловыделение оборудования [кВт]
	totalTDP := 0
	for _, link := range coolrequest.ComponentLink {
		totalTDP += link.Component.TDP * int(link.Count)
	}
	P := float64(totalTDP) / 1000.0 // переводим в кВт

	// V = S × h - объем помещения [м³]
	V := *coolrequest.RoomArea * *coolrequest.RoomHeight

	// Q = P × (1.3 + 15/V) [кВт]
	Q := P * (1.3 + 15.0/V)

	return Q
}

// DELETE /api/requests/:id - удаление заявки
func (r *Repository) LogicallyDeleteRequest(requestID uint) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		var coolrequest ds.CoolRequest

		if err := tx.Preload("ComponentLink").First(&coolrequest, requestID).Error; err != nil {
			return err
		}

		updates := map[string]interface{}{
			"status":       ds.StatusDeleted,
			"forming_date": time.Now(),
		}

		if err := tx.Model(&ds.CoolRequest{}).Where("id = ?", requestID).Updates(updates).Error; err != nil {
			return err
		}

		// Обновляем статус связанных компонентов (делаем неактивными)
		var componentIDs []uint
		for _, link := range coolrequest.ComponentLink {
			componentIDs = append(componentIDs, link.ComponentID)
		}

		if len(componentIDs) > 0 {
			if err := tx.Model(&ds.Component{}).Where("id IN ?", componentIDs).Update("status", false).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// DELETE /api/requests/:id/components/:component_id - удаление компонента из заявки
func (r *Repository) RemoveComponentFromRequest(requestID, componentID uint) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		result := tx.Where("coolrequest_id = ? AND component_id = ?", requestID, componentID).Delete(&ds.ComponentToRequest{})
		if result.Error != nil {
			return result.Error
		}

		if result.RowsAffected == 0 {
			return errors.New("component not found in this coolrequest")
		}

		// Обновляем статус компонента
		if err := tx.Model(&ds.Component{}).Where("id = ?", componentID).Update("status", false).Error; err != nil {
			return err
		}

		// Проверяем, остались ли еще компоненты в заявке
		var remainingCount int64
		if err := tx.Model(&ds.ComponentToRequest{}).Where("coolrequest_id = ?", requestID).Count(&remainingCount).Error; err != nil {
			return err
		}

		// Если компонентов не осталось, удаляем заявку
		if remainingCount == 0 {
			updates := map[string]interface{}{
				"status":       ds.StatusDeleted,
				"forming_date": time.Now(),
			}
			if err := tx.Model(&ds.CoolRequest{}).Where("id = ?", requestID).Updates(updates).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// PUT /api/requests/:id/components/:component_id - изменение м-м связи
func (r *Repository) UpdateMM(requestID, componentID uint, updateData ds.ComponentToRequest) error {
	var link ds.ComponentToRequest
	if err := r.db.Where("coolrequest_id = ? AND component_id = ?", requestID, componentID).First(&link).Error; err != nil {
		return err
	}

	updates := make(map[string]interface{})

	if updateData.Count != 0 {
		updates["count"] = updateData.Count
	}

	if len(updates) == 0 {
		return nil
	}

	return r.db.Model(&link).Updates(updates).Error
}

// POST /api/requests/draft/components/:component_id - добавление компонента в черновик
// func (r *Repository) AddComponentToDraft(userID, componentID uint) error {
// 	return r.db.Transaction(func(tx *gorm.DB) error {
// 		var coolrequest ds.CoolRequest
// 		err := tx.Where("creator_id = ? AND status = ?", userID, ds.StatusDraft).First(&coolrequest).Error
// 		if err != nil {
// 			if errors.Is(err, gorm.ErrRecordNotFound) {
// 				newRequest := ds.CoolRequest{
// 					CreatorID:    userID,
// 					Status:       ds.StatusDraft,
// 					CreationDate: time.Now(),
// 				}
// 				if err := tx.Create(&newRequest).Error; err != nil {
// 					return fmt.Errorf("failed to create draft coolrequest: %w", err)
// 				}
// 				coolrequest = newRequest
// 			} else {
// 				return err
// 			}
// 		}

// 		var count int64
// 		tx.Model(&ds.ComponentToRequest{}).Where("coolrequest_id = ? AND component_id = ?", coolrequest.ID, componentID).Count(&count)
// 		if count > 0 {
// 			// Если компонент уже есть, увеличиваем количество
// 			return tx.Model(&ds.ComponentToRequest{}).
// 				Where("coolrequest_id = ? AND component_id = ?", coolrequest.ID, componentID).
// 				Update("count", gorm.Expr("count + 1")).Error
// 		}

// 		link := ds.ComponentToRequest{
// 			CoolRequestID: coolrequest.ID,
// 			ComponentID:   componentID,
// 			Count:         1,
// 		}

// 		return tx.Create(&link).Error
// 	})
// }

// package repository

// import (
// 	"RIP/internal/app/ds"
// 	"errors"
// )

// func (r *Repository) GetDraftCoolRequest(userID uint) (*ds.CoolRequest, error) {
// 	var coolrequest ds.CoolRequest

// 	err := r.db.Where("creator_id = ? AND status = ?", userID, ds.StatusDraft).First(&coolrequest).Error
// 	if err != nil {
// 		return nil, err
// 	}
// 	return &coolrequest, nil
// }

// func (r *Repository) CreateCoolRequest(coolrequest *ds.CoolRequest) error {
// 	return r.db.Create(coolrequest).Error
// }

// func (r *Repository) AddComponentToCoolRequest(coolRequestID, componentID uint) error {
// 	var count int64

// 	r.db.Model(&ds.ComponentToRequest{}).Where("coolrequest_id = ? AND component_id = ?", coolRequestID, componentID).Count(&count)
// 	if count > 0 {
// 		return errors.New("components already in coolrequest")
// 	}

// 	link := ds.ComponentToRequest{
// 		CoolRequestID: coolRequestID,
// 		ComponentID:   componentID,
// 	}
// 	return r.db.Create(&link).Error
// }

// func (r *Repository) GetCoolRequestWithComponents(CoolRequestID uint) (*ds.CoolRequest, error) {
// 	var CoolRequest ds.CoolRequest

// 	err := r.db.Preload("ComponentLink.Component").First(&CoolRequest, CoolRequestID).Error //???????
// 	if err != nil {
// 		return nil, err
// 	}

// 	if CoolRequest.Status == ds.StatusDeleted {
// 		return nil, errors.New("Sorry. Page not found or has been deleted")
// 	}

// 	return &CoolRequest, nil
// }

// func (r *Repository) LogicallyDeleteCoolRequest(CoolRequestID uint) error {
// 	result := r.db.Exec("UPDATE cool_requests SET status = ? WHERE id = ?", ds.StatusDeleted, CoolRequestID)
// 	return result.Error
// }
