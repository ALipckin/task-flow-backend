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
	for i, shard := range allShards {
		err := shard.AutoMigrate(&models.Task{}, &models.Observer{})
		if err != nil {
			log.Printf("Error migrating shard %d: %v", i, err)
			continue
		}
		log.Printf("Successful shard migration %d", i)
	}
	log.Println("all migration successful")
}
