package shard

import "gorm.io/gorm"

// IdAllocator stores the next ID for the shard (range allocation).
// In each shard, there is one string (id=1): next_id = start of range + number of IDs issued.
type IdAllocator struct {
	ID     int   `gorm:"primaryKey;column:id"`
	NextID int64 `gorm:"column:next_id;not null"`
}

func (IdAllocator) TableName() string {
	return "id_allocator"
}

// SeedIdAllocator sets the initial value of next_id for the shard with index shardIndex.
// Range of shard i: [i*rangeSize+1, (i+1)*rangeSize].
func SeedIdAllocator(db *gorm.DB, shardIndex int, rangeSize uint64) error {
	start := int64(shardIndex)*int64(rangeSize) + 1
	var count int64
	if err := db.Table("id_allocator").Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return nil
	}
	return db.Exec("INSERT INTO id_allocator (id, next_id) VALUES (1, ?)", start).Error
}

// AllocNextID atomically increments next_id on the shard and returns the issued ID.
func AllocNextID(db *gorm.DB) (uint, error) {
	var nextID int64
	// PostgreSQL: RETURNING returns a value after UPDATE
	err := db.Raw("UPDATE id_allocator SET next_id = next_id + 1 WHERE id = 1 RETURNING next_id").Scan(&nextID).Error
	if err != nil {
		return 0, err
	}
	return uint(nextID), nil
}
