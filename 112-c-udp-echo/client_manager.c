#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <arpa/inet.h>
#include "client_manager.h"
#include "heartbeat_protocol.h"

static client_info_t *g_clients = NULL;
static int g_client_count = 0;
static int g_max_clients = HEARTBEAT_MAX_CLIENTS;
static server_stats_t g_stats = {0, 0, 0, 0};
static client_timeout_callback_t g_timeout_callback = NULL;

int client_manager_init(void) {
    g_clients = (client_info_t *)calloc(g_max_clients, sizeof(client_info_t));
    if (g_clients == NULL) {
        perror("calloc failed");
        return -1;
    }
    g_client_count = 0;
    memset(&g_stats, 0, sizeof(g_stats));
    g_timeout_callback = NULL;
    return 0;
}

void client_manager_cleanup(void) {
    if (g_clients != NULL) {
        free(g_clients);
        g_clients = NULL;
    }
    g_client_count = 0;
}

client_info_t *client_manager_find_by_id(const char *client_id) {
    if (client_id == NULL || g_clients == NULL) {
        return NULL;
    }
    for (int i = 0; i < g_client_count; i++) {
        if (g_clients[i].is_online && 
            strncmp(g_clients[i].client_id, client_id, HEARTBEAT_CLIENT_ID_LEN) == 0) {
            return &g_clients[i];
        }
    }
    return NULL;
}

client_info_t *client_manager_find_by_addr(const struct sockaddr_in *addr) {
    if (addr == NULL || g_clients == NULL) {
        return NULL;
    }
    for (int i = 0; i < g_client_count; i++) {
        if (g_clients[i].is_online &&
            g_clients[i].addr.sin_addr.s_addr == addr->sin_addr.s_addr &&
            g_clients[i].addr.sin_port == addr->sin_port) {
            return &g_clients[i];
        }
    }
    return NULL;
}

client_info_t *client_manager_add_client(const char *client_id, 
                                           const struct sockaddr_in *addr) {
    if (client_id == NULL || addr == NULL || g_clients == NULL) {
        return NULL;
    }
    
    if (g_client_count >= g_max_clients) {
        fprintf(stderr, "Maximum clients reached: %d\n", g_max_clients);
        return NULL;
    }
    
    client_info_t *existing = client_manager_find_by_id(client_id);
    if (existing != NULL) {
        return existing;
    }
    
    client_info_t *new_client = &g_clients[g_client_count];
    strncpy(new_client->client_id, client_id, HEARTBEAT_CLIENT_ID_LEN - 1);
    new_client->client_id[HEARTBEAT_CLIENT_ID_LEN - 1] = '\0';
    memcpy(&new_client->addr, addr, sizeof(struct sockaddr_in));
    new_client->last_active = time(NULL);
    new_client->is_online = 1;
    
    g_client_count++;
    g_stats.total_connected++;
    g_stats.online_count++;
    
    return new_client;
}

int client_manager_update_activity(const char *client_id) {
    client_info_t *client = client_manager_find_by_id(client_id);
    if (client == NULL) {
        return -1;
    }
    client->last_active = time(NULL);
    return 0;
}

int client_manager_remove_client(const char *client_id) {
    if (client_id == NULL || g_clients == NULL) {
        return -1;
    }
    
    for (int i = 0; i < g_client_count; i++) {
        if (strncmp(g_clients[i].client_id, client_id, HEARTBEAT_CLIENT_ID_LEN) == 0) {
            if (g_clients[i].is_online) {
                g_stats.online_count--;
            }
            memmove(&g_clients[i], &g_clients[i + 1], 
                    (g_client_count - i - 1) * sizeof(client_info_t));
            g_client_count--;
            return 0;
        }
    }
    return -1;
}

void client_manager_get_stats(server_stats_t *stats) {
    if (stats != NULL) {
        memcpy(stats, &g_stats, sizeof(server_stats_t));
    }
}

void client_manager_print_stats(void) {
    printf("\n========== Server Statistics ==========\n");
    printf("Current online clients: %d\n", g_stats.online_count);
    printf("Total connected clients: %lu\n", (unsigned long)g_stats.total_connected);
    printf("Packets received: %lu\n", (unsigned long)g_stats.packets_received);
    printf("Packets sent: %lu\n", (unsigned long)g_stats.packets_sent);
    printf("========================================\n\n");
}

void client_manager_print_clients(void) {
    if (g_clients == NULL || g_client_count == 0) {
        printf("No active clients\n");
        return;
    }
    
    printf("\n========== Active Clients ==========\n");
    for (int i = 0; i < g_client_count; i++) {
        if (g_clients[i].is_online) {
            char ip_str[INET_ADDRSTRLEN];
            inet_ntop(AF_INET, &g_clients[i].addr.sin_addr, ip_str, INET_ADDRSTRLEN);
            printf("Client [%d]: ID='%s', IP=%s, Port=%d, LastActive=%ld sec ago\n",
                   i + 1,
                   g_clients[i].client_id,
                   ip_str,
                   ntohs(g_clients[i].addr.sin_port),
                   time(NULL) - g_clients[i].last_active);
        }
    }
    printf("=====================================\n\n");
}

void client_manager_set_timeout_callback(client_timeout_callback_t callback) {
    g_timeout_callback = callback;
}

int client_manager_check_timeouts(time_t current_time, int timeout_seconds) {
    if (g_clients == NULL) {
        return 0;
    }
    
    int timeout_count = 0;
    for (int i = g_client_count - 1; i >= 0; i--) {
        if (g_clients[i].is_online) {
            time_t idle_time = current_time - g_clients[i].last_active;
            if (idle_time >= timeout_seconds) {
                printf("Client '%s' timed out (idle %ld seconds)\n",
                       g_clients[i].client_id, (long)idle_time);
                
                if (g_timeout_callback != NULL) {
                    g_timeout_callback(&g_clients[i]);
                }
                
                memmove(&g_clients[i], &g_clients[i + 1], 
                        (g_client_count - i - 1) * sizeof(client_info_t));
                g_client_count--;
                g_stats.online_count--;
                timeout_count++;
            }
        }
    }
    return timeout_count;
}

void client_manager_inc_received(void) {
    g_stats.packets_received++;
}

void client_manager_inc_sent(void) {
    g_stats.packets_sent++;
}
