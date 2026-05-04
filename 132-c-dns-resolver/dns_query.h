#ifndef DNS_QUERY_H
#define DNS_QUERY_H

#include "dns_protocol.h"
#include <stdint.h>
#include <time.h>

#define DNS_MAX_SERVERS 8
#define DNS_MAX_RETRIES 3
#define DNS_DEFAULT_TIMEOUT 2
#define DNS_DEFAULT_MAX_RETRIES 3

#define DNS_QUERY_SUCCESS 0
#define DNS_QUERY_TIMEOUT -1
#define DNS_QUERY_ERROR -2
#define DNS_QUERY_SERVER_ERROR -3
#define DNS_QUERY_NAME_NOT_FOUND -4
#define DNS_QUERY_TRUNCATED -5

typedef struct {
    char server_ip[64];
    int port;
} dns_server_t;

typedef struct {
    dns_server_t servers[DNS_MAX_SERVERS];
    int server_count;
    int timeout_seconds;
    int max_retries;
} dns_resolver_t;

int dns_resolver_init(dns_resolver_t *resolver);
int dns_resolver_add_server(dns_resolver_t *resolver, const char *server_ip, int port);
void dns_resolver_set_timeout(dns_resolver_t *resolver, int seconds);
void dns_resolver_set_retries(dns_resolver_t *resolver, int retries);

int dns_send_query_udp(dns_resolver_t *resolver, int server_index,
                        const uint8_t *query, size_t query_len,
                        uint8_t *response, size_t response_size, size_t *response_len,
                        int timeout_seconds);

int dns_send_query_tcp(dns_resolver_t *resolver, int server_index,
                        const uint8_t *query, size_t query_len,
                        uint8_t *response, size_t response_size, size_t *response_len,
                        int timeout_seconds);

int dns_query_raw(dns_resolver_t *resolver,
                  const char *name, uint16_t qtype,
                  uint8_t *response, size_t response_size, size_t *response_len);

int dns_query_message(dns_resolver_t *resolver,
                      const char *name, uint16_t qtype,
                      dns_message_t *msg, uint8_t *raw_packet, size_t *raw_packet_len);

uint16_t dns_generate_id(void);

int dns_check_truncated(const uint8_t *packet, size_t packet_len);
int dns_check_response_id(const uint8_t *packet, size_t packet_len, uint16_t expected_id);
int dns_get_response_rcode(const uint8_t *packet, size_t packet_len);

#endif
