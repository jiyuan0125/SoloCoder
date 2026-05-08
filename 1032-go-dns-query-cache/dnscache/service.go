package dnscache

import (
	"dns-cache-service/common"
	"time"
)

type Service struct {
	cache     *Cache
	resolver  *Resolver
	stats     *Stats
	preloader *Preloader
}

func NewService(config *Config) *Service {
	cache := NewCache(config)
	resolver := NewResolver()
	stats := NewStats()
	preloader := NewPreloader(cache, resolver, stats)

	return &Service{
		cache:     cache,
		resolver:  resolver,
		stats:     stats,
		preloader: preloader,
	}
}

func (s *Service) Query(domain string, recordTypes []common.RecordType) *common.QueryResponse {
	s.cache.CleanExpired()

	if len(recordTypes) == 0 {
		recordTypes = []common.RecordType{common.TypeA}
	}

	results := make([]common.QueryResult, 0, len(recordTypes))

	for _, recordType := range recordTypes {
		result := s.querySingle(domain, recordType)
		results = append(results, result)
	}

	return &common.QueryResponse{
		Domain:  domain,
		Results: results,
	}
}

func (s *Service) querySingle(domain string, recordType common.RecordType) common.QueryResult {
	if entry, hit := s.cache.Get(domain, recordType); hit {
		s.stats.RecordQuery(true)
		remainingTTL := int(time.Until(entry.ExpiresAt).Seconds())
		if remainingTTL < 0 {
			remainingTTL = 0
		}
		return common.QueryResult{
			Type:         recordType,
			Records:      entry.Records,
			HitCache:     true,
			RemainingTTL: remainingTTL,
			IsNXDOMAIN:   entry.IsNXDomain,
		}
	}

	s.stats.RecordQuery(false)

	records, isNXDomain, err := s.resolver.Resolve(domain, recordType)
	if err != nil {
		return common.QueryResult{
			Type:       recordType,
			HitCache:   false,
			IsNXDOMAIN: false,
		}
	}

	s.cache.Set(domain, recordType, records, isNXDomain)

	remainingTTL := 0
	if len(records) > 0 {
		remainingTTL = records[0].TTL
	}

	return common.QueryResult{
		Type:         recordType,
		Records:      records,
		HitCache:     false,
		RemainingTTL: remainingTTL,
		IsNXDOMAIN:   isNXDomain,
	}
}

func (s *Service) Refresh(domain string, recordTypes []common.RecordType) *common.RefreshResponse {
	if len(recordTypes) == 0 {
		recordTypes = []common.RecordType{common.TypeA, common.TypeAAAA, common.TypeCNAME, common.TypeMX, common.TypeTXT}
	}

	for _, recordType := range recordTypes {
		s.cache.Delete(domain, recordType)
	}

	queryResp := s.Query(domain, recordTypes)

	return &common.RefreshResponse{
		Domain:  domain,
		Results: queryResp.Results,
	}
}

func (s *Service) GetStats() *common.StatsResponse {
	return s.stats.GetStats(s.cache)
}

func (s *Service) PreloadDomains(domains []string, recordTypes []common.RecordType) *common.PreloadResponse {
	total, success, failed := s.preloader.PreloadDomains(domains, recordTypes)
	return &common.PreloadResponse{
		Total:      total,
		Successful: success,
		Failed:     failed,
	}
}

func (s *Service) StartPreloadAsync(domains []string, recordTypes []common.RecordType) {
	s.preloader.PreloadAsync(domains, recordTypes, nil)
}

func (s *Service) GetAllEntries() *common.AllEntriesResponse {
	entries := s.cache.GetAllEntries()
	responseEntries := make([]common.CacheEntryResponse, 0, len(entries))

	for _, entry := range entries {
		remainingTTL := int(time.Until(entry.ExpiresAt).Seconds())
		if remainingTTL < 0 {
			remainingTTL = 0
		}

		responseEntries = append(responseEntries, common.CacheEntryResponse{
			Domain:       entry.Domain,
			Type:         entry.Type,
			Records:      entry.Records,
			ExpiresAt:    entry.ExpiresAt.Format(time.RFC3339),
			RemainingTTL: remainingTTL,
			IsNXDOMAIN:   entry.IsNXDomain,
		})
	}

	return &common.AllEntriesResponse{
		Entries: responseEntries,
		Total:   len(responseEntries),
	}
}
