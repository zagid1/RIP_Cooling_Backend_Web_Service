package repository

import (
	"RIP/internal/app/ds"
	"errors"
	"fmt"

	//"fmt"
	"strconv"
	"time"

	"gorm.io/gorm"
)

// GET /api/requests/cart - иконка корзины
func (r *Repository) GetDraftRequest(userID uint) (*ds.Cooling, error) {
	var cooling ds.Cooling
	err := r.db.Where("creator_id = ? AND status = ?", userID, ds.StatusDraft).First(&cooling).Error
	if err != nil {
		return nil, err
	}
	return &cooling, nil
}

// GET /api/requests/:id - одна заявка с компонентами
func (r *Repository) GetRequestWithComponents(requestID uint, UserID uint, isModerator bool) (*ds.Cooling, error) {
	var Cooling ds.Cooling
	var err error

	// Базовый запрос с предзагрузкой связей
	query := r.db.Preload("ComponentLink.Component").Preload("Creator").Preload("Moderator")

	if !isModerator {
		// Для обычных пользователей - только их запросы
		err = query.Where("id = ? AND creator_id = ?", requestID, UserID).First(&Cooling).Error
	} else {
		// Для модераторов - любой запрос
		err = query.First(&Cooling, requestID).Error
	}
	if err != nil {
		return nil, err
	}

	if Cooling.Status == ds.StatusDeleted {
		return nil, errors.New("cooling page not found or has been deleted")
	}
	return &Cooling, nil
}

// GET /api/requests - список заявок с фильтрацией
// func (r *Repository) RequestsListFiltered(userID uint, isModerator bool, status, from, to string) ([]ds.Cooling, error) {
// 	var results []ds.Cooling

// 	// БЕЗ PRELOAD!
// 	query := r.db

// 	query = query.Where("status != ? AND status != ?", 2, 1)

// 	if err := query.Find(&results).Error; err != nil {
// 		return nil, err
// 	}

// 	fmt.Printf("Records found: %d\n", len(results))
// 	for _, rec := range results {
// 		fmt.Printf("ID: %d, Status: %d, CreatorID: %d, ModeratorID: %v\n",
// 			rec.ID, rec.Status, rec.CreatorID, rec.ModeratorID)
// 	}

// 	return results, nil
// }

func (r *Repository) RequestsListFiltered(userID uint, isModerator bool, status, from, to string) ([]ds.CoolingDTO, error) {
	var requestLIst []ds.Cooling

	// СРОЧНАЯ ОТЛАДКА
	fmt.Printf("=== DEBUG ===\n")
	fmt.Printf("Input - userID: %d, isModerator: %t, status: '%s', from: '%s', to: '%s'\n",
		userID, isModerator, status, from, to)
	fmt.Printf("Constants - StatusDeleted: %d, StatusDraft: %d\n", ds.StatusDeleted, ds.StatusDraft)

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

	var result []ds.CoolingDTO
	for _, cooling := range requestLIst {
		dto := ds.CoolingDTO{
			ID:             cooling.ID,
			Status:         cooling.Status,
			CreationDate:   cooling.CreationDate,
			CreatorID:      cooling.Creator.ID,
			ModeratorID:    nil,
			FormingDate:    cooling.FormingDate,
			CompletionDate: cooling.CompletionDate,
			RoomArea:       cooling.RoomArea,
			RoomHeight:     cooling.RoomHeight,
			CoolingPower:   cooling.CoolingPower,
		}

		if cooling.ModeratorID != nil {
			dto.ModeratorID = &cooling.Moderator.ID
		}
		result = append(result, dto)
	}
	return result, nil
}

// PUT /api/requests/:id - изменение полей заявки
func (r *Repository) UpdateRequestUserFields(id uint, req ds.CoolingUpdateRequest) error {
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

	return r.db.Model(&ds.Cooling{}).Where("id = ?", id).Updates(updates).Error
}

// PUT /api/requests/:id/form - сформировать заявку
func (r *Repository) FormRequest(id uint, creatorID uint) error {
	var cooling ds.Cooling
	// if err := r.db.Preload("ComponentLink.Component").First(&cooling, id).Error; err != nil {
	// 	return err
	// }
	if err := r.db.First(&cooling, id).Error; err != nil {
		return err
	}

	if cooling.CreatorID != creatorID {
		return errors.New("only creator can form cooling")
	}

	if cooling.Status != ds.StatusDraft {
		return errors.New("only draft cooling can be formed")
	}

	if cooling.RoomArea == nil || cooling.RoomHeight == nil {
		return errors.New("room area and height are required")
	}

	// Расчет мощности охлаждения на основе компонентов
	// totalTDP := 0
	// for _, link := range cooling.ComponentLink {
	// 	totalTDP += link.Component.TDP * int(link.Count)
	// }

	//coolingPower := float64(totalTDP)*1.2 + (*cooling.RoomArea * *cooling.RoomHeight * 0.1)

	now := time.Now()
	return r.db.Model(&cooling).Updates(map[string]interface{}{
		"status":       ds.StatusFormed,
		"forming_date": now,
		//	"cooling_power": coolingPower,
	}).Error
}

// PUT /api/requests/:id/resolve - завершить/отклонить заявку
func (r *Repository) ResolveRequest(id uint, moderatorID uint, action string) error {
	return r.db.Transaction(func(tx *gorm.DB) error {

		var cooling ds.Cooling
		if err := tx.Preload("ComponentLink.Component").First(&cooling, id).Error; err != nil {
			return err
		}

		if cooling.Status != ds.StatusFormed {
			return errors.New("only formed cooling can be resolved")
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
				coolingPower := r.calculateCoolingPower(cooling)
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

		var CoolingIDs []uint
		for _, link := range cooling.ComponentLink {
			CoolingIDs = append(CoolingIDs, link.ComponentID)
		}

		if len(CoolingIDs) > 0 {
			if err := tx.Model(&ds.Component{}).Where("id IN ?", CoolingIDs).Update("status", false).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// Функция расчета мощности охлаждения по формуле: Q = P × (1.3 + 15/V) [кВт]
func (r *Repository) calculateCoolingPower(cooling ds.Cooling) float64 {
	// P = Σ(TDP компонентов) / 1000 - тепловыделение оборудования [кВт]
	totalTDP := 0
	for _, link := range cooling.ComponentLink {
		totalTDP += link.Component.TDP * int(link.Count)
	}
	P := float64(totalTDP) / 1000.0 // переводим в кВт

	// V = S × h - объем помещения [м³]
	V := *cooling.RoomArea * *cooling.RoomHeight

	// Q = P × (1.3 + 15/V) [кВт]
	Q := P * (1.3 + 15.0/V)

	return Q
}

// DELETE /api/requests/:id - удаление заявки
func (r *Repository) LogicallyDeleteRequest(requestID uint) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		var cooling ds.Cooling

		if err := tx.Preload("ComponentLink").First(&cooling, requestID).Error; err != nil {
			return err
		}

		updates := map[string]interface{}{
			"status":       ds.StatusDeleted,
			"forming_date": time.Now(),
		}

		if err := tx.Model(&ds.Cooling{}).Where("id = ?", requestID).Updates(updates).Error; err != nil {
			return err
		}

		// Обновляем статус связанных компонентов (делаем неактивными)
		var componentIDs []uint
		for _, link := range cooling.ComponentLink {
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
		result := tx.Where("cooling_id = ? AND component_id = ?", requestID, componentID).Delete(&ds.ComponentToCooling{})
		if result.Error != nil {
			return result.Error
		}

		if result.RowsAffected == 0 {
			return errors.New("component not found in this cooling")
		}

		// Обновляем статус компонента
		if err := tx.Model(&ds.Component{}).Where("id = ?", componentID).Update("status", false).Error; err != nil {
			return err
		}

		// Проверяем, остались ли еще компоненты в заявке
		var remainingCount int64
		if err := tx.Model(&ds.ComponentToCooling{}).Where("cooling_id = ?", requestID).Count(&remainingCount).Error; err != nil {
			return err
		}

		// Если компонентов не осталось, удаляем заявку
		if remainingCount == 0 {
			updates := map[string]interface{}{
				"status":       ds.StatusDeleted,
				"forming_date": time.Now(),
			}
			if err := tx.Model(&ds.Cooling{}).Where("id = ?", requestID).Updates(updates).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// PUT /api/requests/:id/components/:component_id - изменение м-м связи
func (r *Repository) UpdateMM(requestID, componentID uint, updateData ds.ComponentToCooling) error {
	var link ds.ComponentToCooling
	if err := r.db.Where("cooling_id = ? AND component_id = ?", requestID, componentID).First(&link).Error; err != nil {
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
// 		var cooling ds.Cooling
// 		err := tx.Where("creator_id = ? AND status = ?", userID, ds.StatusDraft).First(&cooling).Error
// 		if err != nil {
// 			if errors.Is(err, gorm.ErrRecordNotFound) {
// 				newRequest := ds.Cooling{
// 					CreatorID:    userID,
// 					Status:       ds.StatusDraft,
// 					CreationDate: time.Now(),
// 				}
// 				if err := tx.Create(&newRequest).Error; err != nil {
// 					return fmt.Errorf("failed to create draft cooling: %w", err)
// 				}
// 				cooling = newRequest
// 			} else {
// 				return err
// 			}
// 		}

// 		var count int64
// 		tx.Model(&ds.ComponentToCooling{}).Where("cooling_id = ? AND component_id = ?", cooling.ID, componentID).Count(&count)
// 		if count > 0 {
// 			// Если компонент уже есть, увеличиваем количество
// 			return tx.Model(&ds.ComponentToCooling{}).
// 				Where("cooling_id = ? AND component_id = ?", cooling.ID, componentID).
// 				Update("count", gorm.Expr("count + 1")).Error
// 		}

// 		link := ds.ComponentToCooling{
// 			CoolingID: cooling.ID,
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

// func (r *Repository) GetDraftCooling(userID uint) (*ds.Cooling, error) {
// 	var cooling ds.Cooling

// 	err := r.db.Where("creator_id = ? AND status = ?", userID, ds.StatusDraft).First(&cooling).Error
// 	if err != nil {
// 		return nil, err
// 	}
// 	return &cooling, nil
// }

// func (r *Repository) CreateCooling(cooling *ds.Cooling) error {
// 	return r.db.Create(cooling).Error
// }

// func (r *Repository) AddComponentToCooling(coolRequestID, componentID uint) error {
// 	var count int64

// 	r.db.Model(&ds.ComponentToCooling{}).Where("cooling_id = ? AND component_id = ?", coolRequestID, componentID).Count(&count)
// 	if count > 0 {
// 		return errors.New("components already in cooling")
// 	}

// 	link := ds.ComponentToCooling{
// 		CoolingID: coolRequestID,
// 		ComponentID:   componentID,
// 	}
// 	return r.db.Create(&link).Error
// }

// func (r *Repository) GetCoolingWithComponents(CoolingID uint) (*ds.Cooling, error) {
// 	var Cooling ds.Cooling

// 	err := r.db.Preload("ComponentLink.Component").First(&Cooling, CoolingID).Error //???????
// 	if err != nil {
// 		return nil, err
// 	}

// 	if Cooling.Status == ds.StatusDeleted {
// 		return nil, errors.New("Sorry. Page not found or has been deleted")
// 	}

// 	return &Cooling, nil
// }

// func (r *Repository) LogicallyDeleteCooling(CoolingID uint) error {
// 	result := r.db.Exec("UPDATE cool_requests SET status = ? WHERE id = ?", ds.StatusDeleted, CoolingID)
// 	return result.Error
// }
