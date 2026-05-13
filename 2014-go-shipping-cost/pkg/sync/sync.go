package sync

import (
	"log"

	"shipping-cost/pkg/config"
)

func OnComplete() {
	log.Println("[状态同步] 开始执行关联的状态同步...")

	cfg := config.Get()
	if cfg == nil {
		log.Println("[状态同步] 警告: 配置为空，跳过确认")
		return
	}

	addressCount := len(cfg.AddressDB)
	remoteAreaCount := len(cfg.RemoteAreas)

	log.Printf("[状态同步] 确认数据: 地址库 %d 条, 偏远地区 %d 个", addressCount, remoteAreaCount)
	log.Println("[状态同步] 状态同步完成")
}
