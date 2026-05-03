#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <unistd.h>
#include <signal.h>
#include <sys/select.h>
#include <arpa/inet.h>
#include <errno.h>
#include "heartbeat_protocol.h"
#include "heartbeat_io.h"
#include "client_manager.h"
#include "timeout_checker.h"

static volatile int g_running = 1;
static int g_sockfd = -1;

static void handle_signal(int sig) {
    (void)sig;
    g_running = 0;
    printf("\nReceived shutdown signal, stopping...\n");
}

static void setup_signals(void) {
    struct sigaction sa;
    memset(&sa, 0, sizeof(sa));
    sa.sa_handler = handle_signal;
    sigaction(SIGINT, &sa, NULL);
    sigaction(SIGTERM, &sa, NULL);
}

static void on_client_timeout(const client_info_t *client) {
    char ip_str[INET_ADDRSTRLEN];
    inet_ntop(AF_INET, &client->addr.sin_addr, ip_str, INET_ADDRSTRLEN);
    printf("Client timeout: ID='%s', IP=%s, Port=%d\n",
           client->client_id, ip_str, ntohs(client->addr.sin_port));
}

static void on_timeout_notify(void) {
    client_manager_print_stats();
}

int main(int argc, char *argv[]) {
    uint16_t port = HEARTBEAT_PORT;
    
    if (argc > 1) {
        port = (uint16_t)atoi(argv[1]);
    }
    
    printf("Heartbeat Server starting...\n");
    printf("Port: %d\n", port);
    printf("Packet size: %zu bytes\n", HEARTBEAT_PACKET_SIZE);
    printf("MTU limit: %d bytes\n", HEARTBEAT_MAX_MTU);
    printf("Timeout: %d seconds\n", HEARTBEAT_TIMEOUT_SEC);
    printf("Scan interval: %d second\n", HEARTBEAT_SCAN_INTERVAL_SEC);
    printf("Max clients: %d\n", HEARTBEAT_MAX_CLIENTS);
    printf("========================================\n");
    
    setup_signals();
    
    if (client_manager_init() < 0) {
        fprintf(stderr, "Failed to initialize client manager\n");
        return 1;
    }
    
    client_manager_set_timeout_callback(on_client_timeout);
    
    if (timeout_checker_init(HEARTBEAT_SCAN_INTERVAL_SEC, HEARTBEAT_TIMEOUT_SEC) < 0) {
        fprintf(stderr, "Failed to initialize timeout checker\n");
        client_manager_cleanup();
        return 1;
    }
    
    timeout_checker_set_notify_callback(on_timeout_notify);
    
    g_sockfd = heartbeat_io_create_socket();
    if (g_sockfd < 0) {
        fprintf(stderr, "Failed to create socket\n");
        timeout_checker_cleanup();
        client_manager_cleanup();
        return 1;
    }
    
    if (heartbeat_io_bind_socket(g_sockfd, port) < 0) {
        fprintf(stderr, "Failed to bind socket\n");
        heartbeat_io_close_socket(g_sockfd);
        timeout_checker_cleanup();
        client_manager_cleanup();
        return 1;
    }
    
    if (timeout_checker_start() < 0) {
        fprintf(stderr, "Failed to start timeout checker\n");
        heartbeat_io_close_socket(g_sockfd);
        timeout_checker_cleanup();
        client_manager_cleanup();
        return 1;
    }
    
    printf("Server ready, waiting for heartbeats...\n");
    printf("Press Ctrl+C to stop\n\n");
    
    fd_set readfds;
    struct timeval tv;
    
    while (g_running) {
        FD_ZERO(&readfds);
        FD_SET(g_sockfd, &readfds);
        
        tv.tv_sec = 1;
        tv.tv_usec = 0;
        
        int ret = select(g_sockfd + 1, &readfds, NULL, NULL, &tv);
        
        if (ret < 0) {
            if (errno == EINTR) {
                continue;
            }
            perror("select failed");
            break;
        }
        
        if (ret == 0) {
            continue;
        }
        
        if (FD_ISSET(g_sockfd, &readfds)) {
            heartbeat_packet_t packet;
            struct sockaddr_in client_addr;
            socklen_t addr_len = sizeof(client_addr);
            
            int recv_ret = heartbeat_io_receive(g_sockfd, &packet, &client_addr, &addr_len);
            
            if (recv_ret < 0) {
                fprintf(stderr, "Error receiving packet\n");
                continue;
            }
            
            if (recv_ret == 0) {
                continue;
            }
            
            if (!heartbeat_io_validate_packet(&packet)) {
                char ip_str[INET_ADDRSTRLEN];
                inet_ntop(AF_INET, &client_addr.sin_addr, ip_str, INET_ADDRSTRLEN);
                fprintf(stderr, "Invalid magic number from %s:%d (got 0x%08X, expected 0x%08X)\n",
                        ip_str, ntohs(client_addr.sin_port),
                        ntohl(packet.magic), HEARTBEAT_MAGIC);
                continue;
            }
            
            client_manager_inc_received();
            
            char ip_str[INET_ADDRSTRLEN];
            inet_ntop(AF_INET, &client_addr.sin_addr, ip_str, INET_ADDRSTRLEN);
            
            char safe_client_id[HEARTBEAT_CLIENT_ID_LEN + 1];
            strncpy(safe_client_id, packet.client_id, HEARTBEAT_CLIENT_ID_LEN);
            safe_client_id[HEARTBEAT_CLIENT_ID_LEN] = '\0';
            
            printf("Received heartbeat from %s:%d - ID='%s', Seq=%u, TS=%lu\n",
                   ip_str, ntohs(client_addr.sin_port),
                   safe_client_id,
                   ntohl(packet.seq_num),
                   (unsigned long)ntohll(packet.timestamp));
            
            client_info_t *client = client_manager_find_by_id(safe_client_id);
            
            if (client == NULL) {
                client = client_manager_add_client(safe_client_id, &client_addr);
                if (client == NULL) {
                    fprintf(stderr, "Failed to add client: %s\n", safe_client_id);
                    continue;
                }
                printf("New client registered: '%s' (%s:%d)\n",
                       safe_client_id, ip_str, ntohs(client_addr.sin_port));
            } else {
                client_manager_update_activity(safe_client_id);
            }
            
            int send_ret = heartbeat_io_send(g_sockfd, &packet, &client_addr, addr_len);
            if (send_ret == 0) {
                client_manager_inc_sent();
                printf("Echoed heartbeat to %s:%d\n", ip_str, ntohs(client_addr.sin_port));
            } else {
                fprintf(stderr, "Failed to echo heartbeat to %s:%d\n",
                        ip_str, ntohs(client_addr.sin_port));
            }
        }
    }
    
    printf("\nShutting down server...\n");
    
    timeout_checker_stop();
    printf("Timeout checker stopped\n");
    
    heartbeat_io_close_socket(g_sockfd);
    printf("Socket closed\n");
    
    client_manager_print_stats();
    client_manager_print_clients();
    
    timeout_checker_cleanup();
    client_manager_cleanup();
    
    printf("Server stopped\n");
    
    return 0;
}
