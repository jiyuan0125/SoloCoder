#include "dns_cache.h"
#include <string.h>
#include <stdlib.h>
#include <stdio.h>
#include <strings.h>

void dns_cache_init(dns_cache_t *cache) {
    memset(cache, 0, sizeof(*cache));
}

void dns_cache_cleanup(dns_cache_t *cache) {
    memset(cache, 0, sizeof(*cache));
}

void dns_cache_expire(dns_cache_t *cache) {
    time_t now = time(NULL);
    int new_count = 0;

    for (int i = 0; i < cache->count; i++) {
        if (cache->entries[i].expires_at > now) {
            if (i != new_count) {
                memcpy(&cache->entries[new_count], &cache->entries[i],
                       sizeof(dns_cache_entry_t));
            }
            new_count++;
        }
    }

    cache->count = new_count;
}

static int find_cache_entry(dns_cache_t *cache, const char *name, uint16_t qtype) {
    time_t now = time(NULL);

    for (int i = 0; i < cache->count; i++) {
        if (strcasecmp(cache->entries[i].name, name) == 0 &&
            cache->entries[i].qtype == qtype &&
            cache->entries[i].expires_at > now) {
            return i;
        }
    }

    return -1;
}

int dns_cache_lookup(dns_cache_t *cache, const char *name, uint16_t qtype,
                      dns_message_t *msg, uint8_t *raw_packet, size_t *raw_packet_len) {
    int index = find_cache_entry(cache, name, qtype);
    if (index < 0) {
        return -1;
    }

    dns_cache_entry_t *entry = &cache->entries[index];

    if (msg) {
        memset(msg, 0, sizeof(*msg));
        strncpy(msg->question.name, entry->name, sizeof(msg->question.name) - 1);
        msg->question.qtype = entry->qtype;
        msg->question.qclass = DNS_CLASS_IN;
        msg->answer_count = entry->answer_count;
        memcpy(msg->answers, entry->answers,
               sizeof(dns_rr_t) * entry->answer_count);
    }

    if (raw_packet && raw_packet_len && entry->raw_packet_len > 0) {
        if (*raw_packet_len >= entry->raw_packet_len) {
            memcpy(raw_packet, entry->raw_packet, entry->raw_packet_len);
            *raw_packet_len = entry->raw_packet_len;
        }
    }

    return 0;
}

int dns_cache_store(dns_cache_t *cache, const char *name, uint16_t qtype,
                    const dns_message_t *msg, const uint8_t *raw_packet, size_t raw_packet_len) {
    if (cache->count >= DNS_MAX_CACHE_ENTRIES) {
        dns_cache_expire(cache);
        if (cache->count >= DNS_MAX_CACHE_ENTRIES) {
            if (cache->count > 0) {
                memmove(&cache->entries[0], &cache->entries[1],
                        sizeof(dns_cache_entry_t) * (cache->count - 1));
                cache->count--;
            }
        }
    }

    int index = cache->count++;
    dns_cache_entry_t *entry = &cache->entries[index];

    memset(entry, 0, sizeof(*entry));
    strncpy(entry->name, name, sizeof(entry->name) - 1);
    entry->qtype = qtype;
    entry->answer_count = msg->answer_count;

    uint32_t min_ttl = 0xFFFFFFFF;
    for (int i = 0; i < msg->answer_count && i < 128; i++) {
        memcpy(&entry->answers[i], &msg->answers[i], sizeof(dns_rr_t));
        if (msg->answers[i].ttl < min_ttl) {
            min_ttl = msg->answers[i].ttl;
        }
    }

    entry->ttl = min_ttl;
    entry->expires_at = time(NULL) + min_ttl;

    if (raw_packet && raw_packet_len > 0 && raw_packet_len < sizeof(entry->raw_packet)) {
        memcpy(entry->raw_packet, raw_packet, raw_packet_len);
        entry->raw_packet_len = raw_packet_len;
    }

    return 0;
}

static int has_cname_answer(const dns_message_t *msg, char *cname_buf, size_t cname_buf_size) {
    for (int i = 0; i < msg->answer_count; i++) {
        if (msg->answers[i].type == DNS_TYPE_CNAME) {
            if (cname_buf && cname_buf_size > 0) {
                strncpy(cname_buf, msg->answers[i].name, cname_buf_size - 1);
                cname_buf[cname_buf_size - 1] = '\0';
            }
            return 1;
        }
    }
    return 0;
}

static int has_target_type_answer(const dns_message_t *msg, uint16_t qtype) {
    for (int i = 0; i < msg->answer_count; i++) {
        if (msg->answers[i].type == qtype) {
            return 1;
        }
    }
    return 0;
}

static void extract_target_answers(dns_message_t *dst, const dns_message_t *src, uint16_t qtype) {
    dst->answer_count = 0;
    for (int i = 0; i < src->answer_count; i++) {
        if (src->answers[i].type == qtype && dst->answer_count < 128) {
            memcpy(&dst->answers[dst->answer_count], &src->answers[i], sizeof(dns_rr_t));
            dst->answer_count++;
        }
    }
}

int dns_resolve_with_cname(dns_resolver_t *resolver, dns_cache_t *cache,
                            const char *name, uint16_t qtype,
                            dns_message_t *msg, uint8_t *raw_packet, size_t *raw_packet_len) {
    char current_name[DNS_MAX_NAME_LENGTH];
    char original_name[DNS_MAX_NAME_LENGTH];
    dns_message_t temp_msg;
    uint8_t temp_packet[DNS_MAX_PACKET_SIZE];
    size_t temp_packet_len = sizeof(temp_packet);

    strncpy(current_name, name, sizeof(current_name) - 1);
    strncpy(original_name, name, sizeof(original_name) - 1);

    for (int depth = 0; depth < DNS_MAX_CNAME_DEPTH; depth++) {
        if (depth > 0 && cache) {
            if (dns_cache_lookup(cache, current_name, qtype, &temp_msg,
                                  temp_packet, &temp_packet_len) == 0) {
                if (!has_cname_answer(&temp_msg, NULL, 0) ||
                    has_target_type_answer(&temp_msg, qtype)) {
                    memcpy(msg, &temp_msg, sizeof(dns_message_t));
                    if (raw_packet && raw_packet_len) {
                        if (*raw_packet_len >= temp_packet_len) {
                            memcpy(raw_packet, temp_packet, temp_packet_len);
                            *raw_packet_len = temp_packet_len;
                        }
                    }
                    return DNS_QUERY_SUCCESS;
                }
            }
        }

        temp_packet_len = sizeof(temp_packet);
        int result = dns_query_message(resolver, current_name, qtype,
                                        &temp_msg, temp_packet, &temp_packet_len);

        if (result != DNS_QUERY_SUCCESS) {
            return result;
        }

        if (cache && depth == 0) {
            dns_cache_store(cache, original_name, qtype, &temp_msg,
                            temp_packet, temp_packet_len);
        }

        if (has_target_type_answer(&temp_msg, qtype)) {
            extract_target_answers(msg, &temp_msg, qtype);
            if (raw_packet && raw_packet_len) {
                if (*raw_packet_len >= temp_packet_len) {
                    memcpy(raw_packet, temp_packet, temp_packet_len);
                    *raw_packet_len = temp_packet_len;
                }
            }
            return DNS_QUERY_SUCCESS;
        }

        char cname_target[DNS_MAX_NAME_LENGTH] = {0};
        int found_cname = 0;

        for (int i = 0; i < temp_msg.answer_count; i++) {
            if (temp_msg.answers[i].type == DNS_TYPE_CNAME) {
                if (temp_msg.answers[i].rdata_name[0] != '\0') {
                    strncpy(cname_target, temp_msg.answers[i].rdata_name,
                            sizeof(cname_target) - 1);
                    cname_target[sizeof(cname_target) - 1] = '\0';
                    found_cname = 1;
                    break;
                }
            }
        }

        if (!found_cname) {
            memcpy(msg, &temp_msg, sizeof(dns_message_t));
            if (raw_packet && raw_packet_len) {
                if (*raw_packet_len >= temp_packet_len) {
                    memcpy(raw_packet, temp_packet, temp_packet_len);
                    *raw_packet_len = temp_packet_len;
                }
            }
            return DNS_QUERY_SUCCESS;
        }

        if (strcasecmp(cname_target, current_name) == 0) {
            memcpy(msg, &temp_msg, sizeof(dns_message_t));
            return DNS_QUERY_SUCCESS;
        }

        strncpy(current_name, cname_target, sizeof(current_name) - 1);
        current_name[sizeof(current_name) - 1] = '\0';
    }

    return DNS_QUERY_ERROR;
}

int dns_resolve_a(dns_resolver_t *resolver, dns_cache_t *cache,
                  const char *name,
                  char (*ip_buf)[16], int max_ips, int *ip_count) {
    dns_message_t msg;
    uint8_t packet[DNS_MAX_PACKET_SIZE];
    size_t packet_len = sizeof(packet);

    int result = dns_resolve_with_cname(resolver, cache, name, DNS_TYPE_A,
                                          &msg, packet, &packet_len);
    if (result != DNS_QUERY_SUCCESS) {
        *ip_count = 0;
        return result;
    }

    *ip_count = 0;
    for (int i = 0; i < msg.answer_count && *ip_count < max_ips; i++) {
        if (msg.answers[i].type == DNS_TYPE_A) {
            dns_format_a_record(&msg.answers[i], ip_buf[*ip_count], 16);
            (*ip_count)++;
        }
    }

    return DNS_QUERY_SUCCESS;
}

int dns_resolve_aaaa(dns_resolver_t *resolver, dns_cache_t *cache,
                      const char *name,
                      char (*ip_buf)[46], int max_ips, int *ip_count) {
    dns_message_t msg;
    uint8_t packet[DNS_MAX_PACKET_SIZE];
    size_t packet_len = sizeof(packet);

    int result = dns_resolve_with_cname(resolver, cache, name, DNS_TYPE_AAAA,
                                          &msg, packet, &packet_len);
    if (result != DNS_QUERY_SUCCESS) {
        *ip_count = 0;
        return result;
    }

    *ip_count = 0;
    for (int i = 0; i < msg.answer_count && *ip_count < max_ips; i++) {
        if (msg.answers[i].type == DNS_TYPE_AAAA) {
            dns_format_aaaa_record(&msg.answers[i], ip_buf[*ip_count], 46);
            (*ip_count)++;
        }
    }

    return DNS_QUERY_SUCCESS;
}

int dns_resolve_mx(dns_resolver_t *resolver, dns_cache_t *cache,
                    const char *name,
                    char (*exchange_buf)[256], int *preferences,
                    int max_records, int *record_count) {
    dns_message_t msg;
    uint8_t packet[DNS_MAX_PACKET_SIZE];
    size_t packet_len = sizeof(packet);

    int result = dns_query_message(resolver, name, DNS_TYPE_MX, &msg, packet, &packet_len);
    if (result != DNS_QUERY_SUCCESS) {
        *record_count = 0;
        return result;
    }

    if (cache) {
        dns_cache_store(cache, name, DNS_TYPE_MX, &msg, packet, packet_len);
    }

    *record_count = 0;
    for (int i = 0; i < msg.answer_count && *record_count < max_records; i++) {
        if (msg.answers[i].type == DNS_TYPE_MX) {
            if (msg.answers[i].rdata_name[0] != '\0') {
                preferences[*record_count] = msg.answers[i].mx_preference;
                strncpy(exchange_buf[*record_count], msg.answers[i].rdata_name, 255);
                exchange_buf[*record_count][255] = '\0';
                (*record_count)++;
            }
        }
    }

    return DNS_QUERY_SUCCESS;
}

int dns_resolve_cname(dns_resolver_t *resolver, dns_cache_t *cache,
                       const char *name,
                       char *cname_buf, size_t cname_buf_size) {
    dns_message_t msg;
    uint8_t packet[DNS_MAX_PACKET_SIZE];
    size_t packet_len = sizeof(packet);

    int result = dns_query_message(resolver, name, DNS_TYPE_CNAME, &msg, packet, &packet_len);
    if (result != DNS_QUERY_SUCCESS) {
        if (cname_buf_size > 0) cname_buf[0] = '\0';
        return result;
    }

    if (cache) {
        dns_cache_store(cache, name, DNS_TYPE_CNAME, &msg, packet, packet_len);
    }

    for (int i = 0; i < msg.answer_count; i++) {
        if (msg.answers[i].type == DNS_TYPE_CNAME) {
            if (msg.answers[i].rdata_name[0] != '\0') {
                strncpy(cname_buf, msg.answers[i].rdata_name, cname_buf_size - 1);
                cname_buf[cname_buf_size - 1] = '\0';
                return DNS_QUERY_SUCCESS;
            }
        }
    }

    if (cname_buf_size > 0) cname_buf[0] = '\0';
    return DNS_QUERY_NAME_NOT_FOUND;
}

int dns_resolve_ns(dns_resolver_t *resolver, dns_cache_t *cache,
                    const char *name,
                    char (*ns_buf)[256], int max_records, int *record_count) {
    dns_message_t msg;
    uint8_t packet[DNS_MAX_PACKET_SIZE];
    size_t packet_len = sizeof(packet);

    int result = dns_query_message(resolver, name, DNS_TYPE_NS, &msg, packet, &packet_len);
    if (result != DNS_QUERY_SUCCESS) {
        *record_count = 0;
        return result;
    }

    if (cache) {
        dns_cache_store(cache, name, DNS_TYPE_NS, &msg, packet, packet_len);
    }

    *record_count = 0;
    for (int i = 0; i < msg.answer_count && *record_count < max_records; i++) {
        if (msg.answers[i].type == DNS_TYPE_NS) {
            if (msg.answers[i].rdata_name[0] != '\0') {
                strncpy(ns_buf[*record_count], msg.answers[i].rdata_name, 255);
                ns_buf[*record_count][255] = '\0';
                (*record_count)++;
            }
        }
    }

    return DNS_QUERY_SUCCESS;
}
