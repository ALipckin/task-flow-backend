package shard

import (
	"log"
	"tasks/internal/infrastructure/persistence"
)

// SyncDatabaseForShards performs AutoMigrate for each shard DB. Call after InitShardManager.
func SyncDatabaseForShards() {
	if ShardMgr == nil {
		log.Fatal("ShardManager not initialized")
	}

	allShards := ShardMgr.GetAllShards()

	for i, db := range allShards {
		err := db.AutoMigrate(&persistence.Task{}, &persistence.Observer{})
		if err != nil {
			log.Printf("Error migrating shard %d: %v", i, err)
			continue
		}
		log.Printf("Successful shard migration %d", i)
	}
	log.Println("all migration successful")
}
