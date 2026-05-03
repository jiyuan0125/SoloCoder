#ifndef CLIENT_MANAGER_H
#define CLIENT_MANAGER_H

#include <time.h>
#include <netinet/in.h>
#include "heartbeat_protocol.h"

typedef void (*client_timeout_callback_t)(const client_info_t *client);

int client_manager_init(void);
void client_manager_cleanup(void);

client_info_t *client_manager_find_by_id(const char *client_id);
client_info_t *client_manager_find_by_addr(const struct sockaddr_in *addr);
client_info_t *client_manager_add_client(const char *client_id, 
                                           const struct sockaddr_in *addr);
int client_manager_update_activity(const char *client_id);
int client_manager_remove_client(const char *client_id);

void client_manager_get_stats(server_stats_t *stats);
void client_manager_print_stats(void);
void client_manager_print_clients(void);

void client_manager_set_timeout_callback(client_timeout_callback_t callback);
int client_manager_check_timeouts(time_t current_time, int timeout_seconds);
void client_manager_inc_received(void);
void client_manager_inc_sent(void);

#endif
