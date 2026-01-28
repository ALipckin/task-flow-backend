package initializers

import (
	"TaskStorageService/models"
	"log"
)

func SyncDatabaseForShards() {
	if models.ShardMgr == nil {
		log.Fatal("ShardManager not initialized")
	}

	allShards := models.ShardMgr.GetAllShards()
	rangeSize := models.ShardMgr.RangeSize()

	for i, shard := range allShards {
		err := shard.AutoMigrate(&models.Task{}, &models.Observer{}, &models.IdAllocator{})
		if err != nil {
			log.Printf("Error migrating shard %d: %v", i, err)
			continue
		}
		if err := models.SeedIdAllocator(shard, i, rangeSize); err != nil {
			log.Printf("Error seeding id_allocator on shard %d: %v", i, err)
		}
		log.Printf("Successful shard migration %d", i)
	}
	log.Println("all migration successful")
}
