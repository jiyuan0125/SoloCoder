#ifndef DNS_CACHE_H
#define DNS_CACHE_H

#include "dns_protocol.h"
#include "dns_query.h"
#include <time.h>

#define DNS_MAX_CACHE_ENTRIES 1024
#define DNS_MAX_CNAME_DEPTH 10

typedef struct {
    char name[DNS_MAX_NAME_LENGTH];
    uint16_t qtype;
    dns_rr_t answers[128];
    int answer_count;
    time_t expires_at;
    uint32_t ttl;
    uint8_t raw_packet[DNS_MAX_PACKET_SIZE];
    size_t raw_packet_len;
} dns_cache_entry_t;

typedef struct {
    dns_cache_entry_t entries[DNS_MAX_CACHE_ENTRIES];
    int count;
} dns_cache_t;

void dns_cache_init(dns_cache_t *cache);
void dns_cache_cleanup(dns_cache_t *cache);

int dns_cache_lookup(dns_cache_t *cache, const char *name, uint16_t qtype,
                      dns_message_t *msg, uint8_t *raw_packet, size_t *raw_packet_len);

int dns_cache_store(dns_cache_t *cache, const char *name, uint16_t qtype,
                    const dns_message_t *msg, const uint8_t *raw_packet, size_t raw_packet_len);

void dns_cache_expire(dns_cache_t *cache);

int dns_resolve_with_cname(dns_resolver_t *resolver, dns_cache_t *cache,
                            const char *name, uint16_t qtype,
                            dns_message_t *msg, uint8_t *raw_packet, size_t *raw_packet_len);

int dns_resolve_a(dns_resolver_t *resolver, dns_cache_t *cache,
                  const char *name,
                  char (*ip_buf)[16], int max_ips, int *ip_count);

int dns_resolve_aaaa(dns_resolver_t *resolver, dns_cache_t *cache,
                      const char *name,
                      char (*ip_buf)[46], int max_ips, int *ip_count);

int dns_resolve_mx(dns_resolver_t *resolver, dns_cache_t *cache,
                    const char *name,
                    char (*exchange_buf)[256], int *preferences,
                    int max_records, int *record_count);

int dns_resolve_cname(dns_resolver_t *resolver, dns_cache_t *cache,
                       const char *name,
                       char *cname_buf, size_t cname_buf_size);

int dns_resolve_ns(dns_resolver_t *resolver, dns_cache_t *cache,
                    const char *name,
                    char (*ns_buf)[256], int max_records, int *record_count);

#endif
