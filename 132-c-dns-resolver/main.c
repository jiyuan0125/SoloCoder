#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <unistd.h>
#include <getopt.h>
#include "dns_protocol.h"
#include "dns_query.h"
#include "dns_cache.h"

static void print_usage(const char *prog_name) {
    printf("Usage: %s [OPTIONS] DOMAIN...\n", prog_name);
    printf("\nDNS Resolver - Custom DNS query implementation\n");
    printf("\nOptions:\n");
    printf("  -t, --type TYPE      Query type: A, AAAA, CNAME, MX, NS (default: A)\n");
    printf("  -s, --server IP      DNS server IP (can be used multiple times)\n");
    printf("  -T, --timeout SEC    Timeout in seconds (default: 2)\n");
    printf("  -r, --retries N      Number of retries (default: 3)\n");
    printf("  -n, --no-cache       Disable caching\n");
    printf("  -h, --help           Show this help message\n");
    printf("\nExamples:\n");
    printf("  %s www.example.com\n", prog_name);
    printf("  %s -t AAAA ipv6.google.com\n", prog_name);
    printf("  %s -t MX gmail.com\n", prog_name);
    printf("  %s -s 8.8.8.8 -s 1.1.1.1 www.baidu.com\n", prog_name);
}

static const char *query_result_to_string(int result) {
    switch (result) {
        case DNS_QUERY_SUCCESS: return "Success";
        case DNS_QUERY_TIMEOUT: return "Timeout";
        case DNS_QUERY_ERROR: return "Query Error";
        case DNS_QUERY_SERVER_ERROR: return "Server Error";
        case DNS_QUERY_NAME_NOT_FOUND: return "Name Not Found (NXDOMAIN)";
        case DNS_QUERY_TRUNCATED: return "Truncated (use TCP)";
        default: return "Unknown Error";
    }
}

static int query_domain(dns_resolver_t *resolver, dns_cache_t *cache,
                        const char *domain, uint16_t qtype) {
    char type_str[16];
    dns_type_to_string(qtype, type_str, sizeof(type_str));

    printf("\n");
    printf("============================================================\n");
    printf("Query: %s  Type: %s\n", domain, type_str);
    printf("============================================================\n");

    switch (qtype) {
        case DNS_TYPE_A: {
            char ips[16][16];
            int ip_count = 0;
            int result = dns_resolve_a(resolver, cache, domain, ips, 16, &ip_count);

            printf("Result: %s\n", query_result_to_string(result));
            if (result == DNS_QUERY_SUCCESS && ip_count > 0) {
                printf("\nA Records (%d found):\n", ip_count);
                for (int i = 0; i < ip_count; i++) {
                    printf("  %d. %s\n", i + 1, ips[i]);
                }
            }
            return result;
        }

        case DNS_TYPE_AAAA: {
            char ips[16][46];
            int ip_count = 0;
            int result = dns_resolve_aaaa(resolver, cache, domain, ips, 16, &ip_count);

            printf("Result: %s\n", query_result_to_string(result));
            if (result == DNS_QUERY_SUCCESS && ip_count > 0) {
                printf("\nAAAA Records (%d found):\n", ip_count);
                for (int i = 0; i < ip_count; i++) {
                    printf("  %d. %s\n", i + 1, ips[i]);
                }
            }
            return result;
        }

        case DNS_TYPE_MX: {
            char exchanges[16][256];
            int preferences[16];
            int record_count = 0;
            int result = dns_resolve_mx(resolver, cache, domain, exchanges, preferences, 16, &record_count);

            printf("Result: %s\n", query_result_to_string(result));
            if (result == DNS_QUERY_SUCCESS && record_count > 0) {
                printf("\nMX Records (%d found):\n", record_count);
                for (int i = 0; i < record_count; i++) {
                    printf("  %d. Pref=%d  %s\n", i + 1, preferences[i], exchanges[i]);
                }
            }
            return result;
        }

        case DNS_TYPE_CNAME: {
            char cname[DNS_MAX_NAME_LENGTH];
            int result = dns_resolve_cname(resolver, cache, domain, cname, sizeof(cname));

            printf("Result: %s\n", query_result_to_string(result));
            if (result == DNS_QUERY_SUCCESS && cname[0] != '\0') {
                printf("\nCNAME Record:\n");
                printf("  %s -> %s\n", domain, cname);
            }
            return result;
        }

        case DNS_TYPE_NS: {
            char nameservers[16][256];
            int record_count = 0;
            int result = dns_resolve_ns(resolver, cache, domain, nameservers, 16, &record_count);

            printf("Result: %s\n", query_result_to_string(result));
            if (result == DNS_QUERY_SUCCESS && record_count > 0) {
                printf("\nNS Records (%d found):\n", record_count);
                for (int i = 0; i < record_count; i++) {
                    printf("  %d. %s\n", i + 1, nameservers[i]);
                }
            }
            return result;
        }

        default: {
            dns_message_t msg;
            uint8_t packet[DNS_MAX_PACKET_SIZE];
            size_t packet_len = sizeof(packet);
            int result = dns_query_message(resolver, domain, qtype, &msg, packet, &packet_len);

            printf("Result: %s\n", query_result_to_string(result));
            if (result == DNS_QUERY_SUCCESS && msg.answer_count > 0) {
                printf("\nAnswers (%d found):\n", msg.answer_count);
                for (int i = 0; i < msg.answer_count; i++) {
                    char rr_type[16];
                    dns_type_to_string(msg.answers[i].type, rr_type, sizeof(rr_type));
                    printf("  %d. Type=%s, TTL=%u, Length=%u\n",
                           i + 1, rr_type, msg.answers[i].ttl, msg.answers[i].rdlength);
                }
            }
            return result;
        }
    }
}

int main(int argc, char *argv[]) {
    int opt;
    uint16_t qtype = DNS_TYPE_A;
    int timeout = DNS_DEFAULT_TIMEOUT;
    int retries = DNS_DEFAULT_MAX_RETRIES;
    int use_cache = 1;
    const char *custom_servers[DNS_MAX_SERVERS];
    int custom_server_count = 0;

    static struct option long_options[] = {
        {"type", required_argument, 0, 't'},
        {"server", required_argument, 0, 's'},
        {"timeout", required_argument, 0, 'T'},
        {"retries", required_argument, 0, 'r'},
        {"no-cache", no_argument, 0, 'n'},
        {"help", no_argument, 0, 'h'},
        {0, 0, 0, 0}
    };

    while ((opt = getopt_long(argc, argv, "t:s:T:r:nh", long_options, NULL)) != -1) {
        switch (opt) {
            case 't':
                if (dns_string_to_type(optarg, &qtype) != 0) {
                    fprintf(stderr, "Unknown record type: %s\n", optarg);
                    fprintf(stderr, "Supported types: A, AAAA, CNAME, MX, NS\n");
                    return 1;
                }
                break;
            case 's':
                if (custom_server_count < DNS_MAX_SERVERS) {
                    custom_servers[custom_server_count++] = optarg;
                }
                break;
            case 'T':
                timeout = atoi(optarg);
                if (timeout < 1) timeout = 1;
                break;
            case 'r':
                retries = atoi(optarg);
                if (retries < 0) retries = 0;
                break;
            case 'n':
                use_cache = 0;
                break;
            case 'h':
                print_usage(argv[0]);
                return 0;
            default:
                print_usage(argv[0]);
                return 1;
        }
    }

    if (optind >= argc) {
        print_usage(argv[0]);
        printf("\n");
        printf("Running demo queries...\n");

        dns_resolver_t resolver;
        dns_cache_t cache;

        dns_resolver_init(&resolver);
        dns_cache_init(&cache);

        dns_resolver_set_timeout(&resolver, 3);
        dns_resolver_set_retries(&resolver, 2);

        printf("\n=== DNS Resolver Demo ===\n");
        printf("\nConfigured DNS Servers:\n");
        for (int i = 0; i < resolver.server_count; i++) {
            printf("  %d. %s:%d\n", i + 1, resolver.servers[i].server_ip, resolver.servers[i].port);
        }

        const char *demo_domains[] = {
            "www.baidu.com",
            "www.google.com",
            "www.github.com",
            NULL
        };

        int success_count = 0;
        int total_count = 0;

        for (int i = 0; demo_domains[i] != NULL; i++) {
            total_count++;
            int result = query_domain(&resolver, use_cache ? &cache : NULL,
                                       demo_domains[i], DNS_TYPE_A);
            if (result == DNS_QUERY_SUCCESS) {
                success_count++;
            }
        }

        if (use_cache) {
            printf("\n");
            printf("------------------------------------------------------------\n");
            printf("Testing cache - querying www.baidu.com again...\n");
            printf("------------------------------------------------------------\n");
            int result = query_domain(&resolver, &cache, "www.baidu.com", DNS_TYPE_A);
            if (result == DNS_QUERY_SUCCESS) {
                printf("(Second query should be faster - coming from cache)\n");
            }
        }

        printf("\n");
        printf("============================================================\n");
        printf("Summary: %d/%d queries succeeded\n", success_count, total_count);
        printf("============================================================\n");

        dns_cache_cleanup(&cache);
        return 0;
    }

    dns_resolver_t resolver;
    dns_cache_t cache;

    dns_resolver_init(&resolver);
    dns_cache_init(&cache);

    dns_resolver_set_timeout(&resolver, timeout);
    dns_resolver_set_retries(&resolver, retries);

    if (custom_server_count > 0) {
        resolver.server_count = 0;
        for (int i = 0; i < custom_server_count; i++) {
            dns_resolver_add_server(&resolver, custom_servers[i], DNS_PORT);
        }
    }

    printf("\n=== DNS Resolver ===\n");
    printf("\nConfigured DNS Servers:\n");
    for (int i = 0; i < resolver.server_count; i++) {
        printf("  %d. %s:%d\n", i + 1, resolver.servers[i].server_ip, resolver.servers[i].port);
    }
    printf("Timeout: %d seconds\n", resolver.timeout_seconds);
    printf("Retries: %d\n", resolver.max_retries);
    printf("Cache: %s\n", use_cache ? "Enabled" : "Disabled");

    int success_count = 0;
    int total_count = 0;

    for (int i = optind; i < argc; i++) {
        total_count++;
        int result = query_domain(&resolver, use_cache ? &cache : NULL, argv[i], qtype);
        if (result == DNS_QUERY_SUCCESS) {
            success_count++;
        }
    }

    printf("\n");
    printf("============================================================\n");
    printf("Summary: %d/%d queries succeeded\n", success_count, total_count);
    printf("============================================================\n");

    dns_cache_cleanup(&cache);
    return (success_count == total_count) ? 0 : 1;
}
