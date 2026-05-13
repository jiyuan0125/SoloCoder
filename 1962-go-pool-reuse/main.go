package main

import (
	"log"
	"os"
	"time"

	"github.com/gin-gonic/gin"

	"shardpool/internal/pool"
	"shardpool/internal/server"
)

func main() {
	cfg := pool.LoadConfigFromEnv()

	p := pool.NewShardPool(cfg)
	p.StartCleanup(30 * time.Second)

	srv := server.NewServer(p)

	r := gin.Default()

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	poolGroup := r.Group("/pool")
	{
		poolGroup.POST("/acquire", srv.Acquire)
		poolGroup.POST("/release", srv.Release)
		poolGroup.GET("/shards", srv.ListShards)
		poolGroup.GET("/stats", srv.Stats)
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8616"
	}

	log.Printf("Shard pool server starting on port %s", port)
	log.Printf("Configuration: shards=%d, virtual_nodes=%d, max_conns_per_shard=%d, idle_timeout=%s, max_lifetime=%s",
		cfg.ShardCount, cfg.VirtualNodeCount, cfg.MaxConnections, cfg.IdleTimeout, cfg.MaxLifetime)

	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
