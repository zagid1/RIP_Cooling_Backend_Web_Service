package repository

import (
	"encoding/json"
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"errors"
	ds "RIP/internal/app/ds"
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


func (r *Repository) GetCoolingListFiltered(userID uint, isModerator bool, filterCreatorID uint, statusStr, from, to string, page, limit int, useIndex bool) (*ds.PaginatedCoolingResponse, error) {
	var requestList []ds.Cooling
	var total int64
	var userStats []ds.UserStatDTO

	// --- 1. ФУНКЦИЯ ФИЛЬТРАЦИИ ---
	applyFilters := func(db *gorm.DB) *gorm.DB {
		q := db.Model(&ds.Cooling{})

		// Исключаем удаленные
		q = q.Where("status != ?", ds.StatusDeleted)

		// ЛОГИКА ФИЛЬТРАЦИИ ПО ПОЛЬЗОВАТЕЛЮ
		if !isModerator {
			// Обычный юзер видит только свои
			q = q.Where("creator_id = ?", userID)
		} else {
			// Модератор: если выбран конкретный юзер для фильтрации
			if filterCreatorID != 0 {
				q = q.Where("creator_id = ?", filterCreatorID)
			}
		}

		// Остальные фильтры (Статус)
		if statusStr != "" && statusStr != "all" {
			if strings.Contains(statusStr, ",") {
				parts := strings.Split(statusStr, ",")
				q = q.Where("status IN ?", parts)
			} else {
				if v, err := strconv.Atoi(statusStr); err == nil {
					q = q.Where("status = ?", v)
				}
			}
		}

		// Фильтры по дате
		if from != "" {
			if fromTime, err := time.Parse("2006-01-02", from); err == nil {
				q = q.Where("forming_date >= ?", fromTime)
			}
		}
		if to != "" {
			if toTime, err := time.Parse("2006-01-02", to); err == nil {
				toTime = toTime.Add(24 * time.Hour)
				q = q.Where("forming_date < ?", toTime)
			}
		}
		return q
	}

	// --- 2. ПОДСЧЕТ TOTAL (С учетом фильтров!) ---
	if err := applyFilters(r.db).Count(&total).Error; err != nil {
		return nil, err
	}

	// --- 3. ЗАПРОС СПИСКА ---
	buildFullQuery := func(db *gorm.DB) *gorm.DB {
		// Применяем фильтры и подгружаем связи
		q := applyFilters(db).Preload("Creator").Preload("Moderator")

		// Логика Индекса (для задания)
		if useIndex {
			q = q.Order("id DESC")
		} else {
			// Трюк для отключения индекса (Full Scan)
			q = q.Clauses(clause.OrderBy{Expression: clause.Expr{SQL: "(id + 0) DESC", WithoutParentheses: true}})
		}

		offset := (page - 1) * limit
		return q.Limit(limit).Offset(offset)
	}

	// --- 4. EXPLAIN ANALYZE (Замер времени) ---
	var executionTimeMs float64 = 0
	dummy := []ds.Cooling{}
	// Генерируем SQL для Explain
	sql := r.db.ToSQL(func(tx *gorm.DB) *gorm.DB {
		return buildFullQuery(tx).Find(&dummy)
	})
	
	var explainJSON string
	// Пытаемся получить план запроса
	r.db.Raw("EXPLAIN (ANALYZE, FORMAT JSON) " + sql).Row().Scan(&explainJSON)
	
	// Парсим результат
	var resp ds.PostgresExplainResponse
	if json.Unmarshal([]byte(explainJSON), &resp) == nil && len(resp) > 0 {
		executionTimeMs = resp[0].ExecutionTime
	}

	// --- 5. ВЫПОЛНЕНИЕ РЕАЛЬНОГО ЗАПРОСА ---
	if err := buildFullQuery(r.db).Find(&requestList).Error; err != nil {
		return nil, err
	}

	// --- 6. ПОДСЧЕТ СТАТИСТИКИ (Только для модератора) ---
	if isModerator {
		// Считаем без фильтров даты/статуса, чтобы видеть общую картину
		r.db.Model(&ds.Cooling{}).
			Select("creator_id as user_id, count(*) as count").
			Where("status != ?", ds.StatusDeleted).
			Group("creator_id").
			Scan(&userStats)
	}

	// --- 7. МАППИНГ (ИСПРАВЛЕНО: Добавлены поля!) ---
	var resultItems []ds.CoolingDTO
	for _, cooling := range requestList {
		dto := ds.CoolingDTO{
			ID:             cooling.ID,
			Status:         cooling.Status,
			CreationDate:   cooling.CreationDate, 
			CreatorID:      cooling.CreatorID,
			
			// Указатели перекладываем как есть
			FormingDate:    cooling.FormingDate,
			CompletionDate: cooling.CompletionDate,
			RoomArea:       cooling.RoomArea,
			RoomHeight:     cooling.RoomHeight,
			CoolingPower:   cooling.CoolingPower,
		}
		
		if cooling.ModeratorID != nil {
			dto.ModeratorID = cooling.ModeratorID
		}
		
		resultItems = append(resultItems, dto)
	}

	return &ds.PaginatedCoolingResponse{
		Items:         resultItems,
		Total:         total,
		QueryDuration: executionTimeMs,
		UserStats:     userStats,
	}, nil
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
