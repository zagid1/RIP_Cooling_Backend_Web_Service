package repository

import (
	ds "RIP/internal/app/ds"
	"context"
	"fmt"
	"math/rand"
	"os"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Repository struct {
	db            *gorm.DB
	minioClient   *minio.Client
	bucketName    string
	minioEndpoint string
}

// SeedCoolingData генерирует 100,000 фейковых заявок
func (r *Repository) SeedCoolingData() error {
	// Настройка рандома (для старых версий Go, в новых не обязательно)
	rand.Seed(time.Now().UnixNano())

	const totalRecords = 100000
	const batchSize = 2000 // Вставляем по 2000 штук за раз

	// Список ID пользователей
	userIDs := []uint{1, 3, 8, 11, 13}

	var batch []ds.Cooling

	for i := 0; i < totalRecords; i++ {
		// 1. Выбираем случайного пользователя
		randomUser := userIDs[rand.Intn(len(userIDs))]

		// 2. Площадь от 40 до 100
		area := 40.0 + rand.Float64()*(100.0-40.0)

		// 3. Высота от 2.5 до 4.0 метров (стандартные диапазоны)
		height := 2.5 + rand.Float64()*(4.0-2.5)

		// 4. Дата формирования: от "сейчас" до "полгода назад"
		// 6 месяцев * 30 дней * 24 часа = ~4320 часов
		hoursBack := rand.Intn(4320)
		randomTime := time.Now().Add(-time.Duration(hoursBack) * time.Hour)

		// Создаем объект
		req := ds.Cooling{
			CreatorID:    randomUser,
			Status:       3, // Статус Черновик (как просил)
			CreationDate: randomTime,
			FormingDate:  &randomTime, // Заполняем, чтобы работали фильтры по дате
			RoomArea:     &area,
			RoomHeight:   &height,
			CoolingPower: nil, // Мощности нет (пусто)
		}

		batch = append(batch, req)

		// Когда накопилась пачка, записываем в БД
		if len(batch) >= batchSize {
			if err := r.db.CreateInBatches(batch, batchSize).Error; err != nil {
				return err
			}
			batch = nil // Очищаем слайс для новой пачки
		}
	}

	// Записываем остаток, если количество не кратно batchSize
	if len(batch) > 0 {
		if err := r.db.CreateInBatches(batch, len(batch)).Error; err != nil {
			return err
		}
	}

	return nil
}

func New(dsn string) (*Repository, error) {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	endpoint := os.Getenv("MINIO_ENDPOINT")
	accessKey := os.Getenv("MINIO_ACCESS_KEY")
	secretKey := os.Getenv("MINIO_SECRET_KEY")
	bucketName := os.Getenv("MINIO_BUCKET")
	useSSL := os.Getenv("MINIO_USE_SSL") == "true"

	minioClient, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create MinIO client: %v", err)
	}

	ctx := context.Background()
	exists, err := minioClient.BucketExists(ctx, bucketName)
	if err != nil {
		return nil, fmt.Errorf("failed to check bucket existence: %v", err)
	}
	if !exists {
		err = minioClient.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{})
		if err != nil {
			return nil, fmt.Errorf("failed to create bucket: %v", err)
		}
	}

	return &Repository{
		db:            db,
		minioClient:   minioClient,
		bucketName:    bucketName,
		minioEndpoint: endpoint,
	}, nil
}
