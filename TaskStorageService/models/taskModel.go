package models

import (
	"gorm.io/gorm"
	"log"
	"time"
)

type Task struct {
	ID          uint       `gorm:"primaryKey"`
	Title       string     `gorm:"type:varchar(255);not null"`
	Description string     `gorm:"type:text"`
	PerformerId uint       `gorm:"not null;index"`
	CreatorId   uint       `gorm:"not null;index"`
	Observers   []Observer `gorm:"foreignKey:TaskId;references:ID"`
	Status      string     `gorm:"type:varchar(50);not null;default:'pending'"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   gorm.DeletedAt `gorm:"index"`
}

func (t *Task) ObserverIDs() []uint64 {
	// Загрузка связанных наблюдателей
	if err := DB.Preload("Observers").First(t, t.ID).Error; err != nil {
		log.Println("Error loading task with observers:", err)
		return nil
	}

	if len(t.Observers) == 0 {
		log.Println("No observers found for the task")
		return nil
	}

	observerIds := make([]uint64, len(t.Observers))
	for i, observer := range t.Observers {
		log.Println("observer id: ", observer.ID)
		observerIds[i] = uint64(observer.UserId)
	}

	log.Println("observerIds:", observerIds)
	return observerIds
}

func ObserversFromIDs(observerIds []uint64) []Observer {
	observers := make([]Observer, len(observerIds))
	for i, observerId := range observerIds {
		observers[i] = Observer{UserId: uint(observerId)}
	}
	return observers
}
