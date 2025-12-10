	package repository

	import (
		"RIP/internal/app/ds"
		"errors"

		//"fmt"
		"strconv"
		"time"

		"gorm.io/gorm"
	)

	// GET /api/cooling/coolcart - иконка корзины
	func (r *Repository) GetDraftRequest(userID uint) (*ds.Cooling, error) {
		var cooling ds.Cooling
		// Используем Limit(1).Find вместо First
		// Find НЕ вызывает ошибку, если запись не найдена, а просто возвращает RowsAffected = 0
		result := r.db.Where("creator_id = ? AND status = ?", userID, ds.StatusDraft).Limit(1).Find(&cooling)

		if result.Error != nil {
			return nil, result.Error // Это реальная ошибка БД (соединение и т.д.)
		}

		// Если записей 0, значит черновика нет -> возвращаем nil (без ошибки)
		if result.RowsAffected == 0 {
			return nil, nil
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
		// fmt.Printf("=== DEBUG ===\n")
		// fmt.Printf("Input - userID: %d, isModerator: %t, status: '%s', from: '%s', to: '%s'\n",
		// 	userID, isModerator, status, from, to)
		// fmt.Printf("Constants - StatusDeleted: %d, StatusDraft: %d\n", ds.StatusDeleted, ds.StatusDraft)

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

		now := time.Now()
		return r.db.Model(&cooling).Updates(map[string]interface{}{
			"status":       ds.StatusFormed,
			"forming_date": now,
			//	"cooling_power": coolingPower,
		}).Error
	}

	// 1. Метод для обновления результата (вызывается хендлером, когда Python присылает ответ)
	func (r *Repository) UpdateCoolingResult(id uint, coolingPower float64) error {
		// Обновляем только поле cooling_power
		return r.db.Model(&ds.Cooling{}).Where("id = ?", id).Updates(map[string]interface{}{
			"cooling_power": coolingPower,
		}).Error
	}

	// 2. Метод решения модератора (Принять/Отклонить)
	func (r *Repository) ResolveRequest(id uint, moderatorID uint, action string) error {
		return r.db.Transaction(func(tx *gorm.DB) error {

			var cooling ds.Cooling
			// Preload нужен, так как ниже мы используем ComponentLink для получения ID компонентов
			if err := tx.Preload("ComponentLink").First(&cooling, id).Error; err != nil {
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
					// Ставим статус "Завершена"
					updates["status"] = ds.StatusCompleted
				}
			case "reject":
				{
					// Ставим статус "Отклонена"
					updates["status"] = ds.StatusRejected
				}
			default:
				{
					return errors.New("invalid action, must be 'complete' or 'reject'")
				}
			}

			// --- ЛОГИКА КОМПОНЕНТОВ (Ваша бизнес-логика) ---
			// Если заявка обрабатывается, меняем статус самих компонентов (например, списываем)
			var CoolingIDs []uint
			for _, link := range cooling.ComponentLink {
				CoolingIDs = append(CoolingIDs, link.ComponentID)
			}

			if len(CoolingIDs) > 0 {
				// Меняем статус компонентов на false (заняты/использованы)
				if err := tx.Model(&ds.Component{}).Where("id IN ?", CoolingIDs).Update("status", false).Error; err != nil {
					return err
				}
			}

			// Применяем обновления к самой заявке
			if err := tx.Model(&cooling).Updates(updates).Error; err != nil {
				return err
			}

			return nil
		})
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


