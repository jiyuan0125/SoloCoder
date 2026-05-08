package dnscache

import (
	"dns-cache-service/common"
	"log"
	"sync"
)

type Preloader struct {
	cache    *Cache
	resolver *Resolver
	stats    *Stats
}

func NewPreloader(cache *Cache, resolver *Resolver, stats *Stats) *Preloader {
	return &Preloader{
		cache:    cache,
		resolver: resolver,
		stats:    stats,
	}
}

func (p *Preloader) PreloadDomain(domain string, recordTypes []common.RecordType) (int, int) {
	if len(recordTypes) == 0 {
		recordTypes = []common.RecordType{common.TypeA, common.TypeAAAA}
	}

	successCount := 0
	failedCount := 0

	for _, recordType := range recordTypes {
		records, isNXDomain, err := p.resolver.Resolve(domain, recordType)
		if err != nil {
			failedCount++
			continue
		}

		if len(records) > 0 || isNXDomain {
			p.cache.Set(domain, recordType, records, isNXDomain)
			if !isNXDomain {
				successCount++
			} else {
				failedCount++
			}
		}
	}

	return successCount, failedCount
}

func (p *Preloader) PreloadDomains(domains []string, recordTypes []common.RecordType) (int, int, int) {
	total := len(domains)
	successCount := 0
	failedCount := 0

	for _, domain := range domains {
		succ, fail := p.PreloadDomain(domain, recordTypes)
		successCount += succ
		failedCount += fail
	}

	return total, successCount, failedCount
}

func (p *Preloader) PreloadAsync(domains []string, recordTypes []common.RecordType, wg *sync.WaitGroup) {
	if wg != nil {
		wg.Add(1)
	}

	go func() {
		if wg != nil {
			defer wg.Done()
		}
		log.Printf("Starting preloading %d domains...", len(domains))
		total, success, failed := p.PreloadDomains(domains, recordTypes)
		log.Printf("Preload completed: total=%d, success=%d, failed=%d", total, success, failed)
	}()
}
