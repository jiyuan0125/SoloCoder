#ifndef HEARTBEAT_PROTOCOL_H
#define HEARTBEAT_PROTOCOL_H

#include <stdint.h>
#include <netinet/in.h>
#include <byteswap.h>

#if __BYTE_ORDER__ == __ORDER_LITTLE_ENDIAN__
#define htonll(x) bswap_64(x)
#define ntohll(x) bswap_64(x)
#else
#define htonll(x) (x)
#define ntohll(x) (x)
#endif

#define HEARTBEAT_MAGIC 0xDEADBEEF
#define HEARTBEAT_CLIENT_ID_LEN 20
#define HEARTBEAT_PORT 9999
#define HEARTBEAT_TIMEOUT_SEC 10
#define HEARTBEAT_SCAN_INTERVAL_SEC 1
#define HEARTBEAT_MAX_CLIENTS 1024
#define HEARTBEAT_MAX_MTU 1472

#pragma pack(push, 1)
typedef struct {
    uint32_t magic;
    uint32_t seq_num;
    uint64_t timestamp;
    char client_id[HEARTBEAT_CLIENT_ID_LEN];
} heartbeat_packet_t;
#pragma pack(pop)

#define HEARTBEAT_PACKET_SIZE sizeof(heartbeat_packet_t)

typedef struct {
    char client_id[HEARTBEAT_CLIENT_ID_LEN];
    struct sockaddr_in addr;
    time_t last_active;
    int is_online;
} client_info_t;

typedef struct {
    int online_count;
    uint64_t total_connected;
    uint64_t packets_received;
    uint64_t packets_sent;
} server_stats_t;

#endif
