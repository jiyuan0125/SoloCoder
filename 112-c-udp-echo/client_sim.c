#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <unistd.h>
#include <signal.h>
#include <sys/socket.h>
#include <sys/types.h>
#include <netinet/in.h>
#include <arpa/inet.h>
#include <sys/time.h>
#include <errno.h>
#include <stdint.h>
#include <time.h>
#include <fcntl.h>
#include "heartbeat_protocol.h"

#define HEARTBEAT_INTERVAL_SEC 2
#define MAX_MISSING_ACKS 3

static volatile int g_running = 1;
static int g_sockfd = -1;
static uint32_t g_seq_num = 0;
static uint32_t g_last_acked_seq = 0;
static int g_missing_acks = 0;

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

static uint64_t get_timestamp_ms(void) {
    struct timeval tv;
    gettimeofday(&tv, NULL);
    return (uint64_t)tv.tv_sec * 1000 + (uint64_t)tv.tv_usec / 1000;
}

static int create_socket(void) {
    int sockfd = socket(AF_INET, SOCK_DGRAM, 0);
    if (sockfd < 0) {
        perror("socket creation failed");
        return -1;
    }
    
    int flags = fcntl(sockfd, F_GETFL, 0);
    if (flags < 0) {
        perror("fcntl F_GETFL failed");
        close(sockfd);
        return -1;
    }
    if (fcntl(sockfd, F_SETFL, flags | O_NONBLOCK) < 0) {
        perror("fcntl F_SETFL O_NONBLOCK failed");
        close(sockfd);
        return -1;
    }
    
    return sockfd;
}

static int send_heartbeat(int sockfd, const struct sockaddr_in *server_addr, 
                           const char *client_id) {
    heartbeat_packet_t packet;
    memset(&packet, 0, sizeof(packet));
    
    packet.magic = htonl(HEARTBEAT_MAGIC);
    packet.seq_num = htonl(g_seq_num);
    packet.timestamp = htonll(get_timestamp_ms());
    strncpy(packet.client_id, client_id, HEARTBEAT_CLIENT_ID_LEN - 1);
    packet.client_id[HEARTBEAT_CLIENT_ID_LEN - 1] = '\0';
    
    ssize_t n = sendto(sockfd, &packet, sizeof(packet), 0,
                        (struct sockaddr *)server_addr, sizeof(*server_addr));
    if (n < 0) {
        perror("sendto failed");
        return -1;
    }
    
    printf("Sent heartbeat: Seq=%u, ID='%s'\n", g_seq_num, client_id);
    g_seq_num++;
    return 0;
}

static int receive_ack(int sockfd, struct sockaddr_in *server_addr, 
                        socklen_t *addr_len, const char *expected_client_id) {
    heartbeat_packet_t packet;
    memset(&packet, 0, sizeof(packet));
    
    ssize_t n = recvfrom(sockfd, &packet, sizeof(packet), 0,
                          (struct sockaddr *)server_addr, addr_len);
    if (n < 0) {
        if (errno == EAGAIN || errno == EWOULDBLOCK) {
            return 0;
        }
        perror("recvfrom failed");
        return -1;
    }
    
    if (n != sizeof(packet)) {
        fprintf(stderr, "Invalid packet size: %zd (expected %zu)\n", n, sizeof(packet));
        return -1;
    }
    
    if (packet.magic != htonl(HEARTBEAT_MAGIC)) {
        fprintf(stderr, "Invalid magic number: 0x%08X\n", ntohl(packet.magic));
        return -1;
    }
    
    if (strncmp(packet.client_id, expected_client_id, HEARTBEAT_CLIENT_ID_LEN) != 0) {
        fprintf(stderr, "Mismatched client ID in response\n");
        return -1;
    }
    
    uint32_t received_seq = ntohl(packet.seq_num);
    uint64_t rtt = get_timestamp_ms() - ntohll(packet.timestamp);
    
    printf("Received ACK: Seq=%u, RTT=%lu ms, ID='%s'\n", 
           received_seq, (unsigned long)rtt, packet.client_id);
    
    if (received_seq > g_last_acked_seq) {
        g_last_acked_seq = received_seq;
        g_missing_acks = 0;
    }
    
    return 1;
}

void print_usage(const char *prog_name) {
    printf("Usage: %s [options] <server_ip> [port]\n", prog_name);
    printf("Options:\n");
    printf("  -i <id>       Client ID (max 19 chars, default: random)\n");
    printf("  -t <interval> Heartbeat interval in seconds (default: 2)\n");
    printf("  -c <count>    Number of heartbeats to send (0 = infinite, default: 0)\n");
    printf("Example:\n");
    printf("  %s 127.0.0.1\n", prog_name);
    printf("  %s -i \"Player_001\" -t 1 192.168.1.100 9999\n", prog_name);
}

int main(int argc, char *argv[]) {
    const char *server_ip = NULL;
    uint16_t port = HEARTBEAT_PORT;
    char client_id[HEARTBEAT_CLIENT_ID_LEN];
    int interval_sec = HEARTBEAT_INTERVAL_SEC;
    int max_count = 0;
    int count = 0;
    
    memset(client_id, 0, sizeof(client_id));
    srand((unsigned int)time(NULL));
    snprintf(client_id, HEARTBEAT_CLIENT_ID_LEN, "Client_%04d", rand() % 10000);
    
    int opt;
    while ((opt = getopt(argc, argv, "i:t:c:h")) != -1) {
        switch (opt) {
            case 'i':
                strncpy(client_id, optarg, HEARTBEAT_CLIENT_ID_LEN - 1);
                client_id[HEARTBEAT_CLIENT_ID_LEN - 1] = '\0';
                break;
            case 't':
                interval_sec = atoi(optarg);
                if (interval_sec < 1) interval_sec = 1;
                break;
            case 'c':
                max_count = atoi(optarg);
                if (max_count < 0) max_count = 0;
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
        fprintf(stderr, "Error: Server IP is required\n\n");
        print_usage(argv[0]);
        return 1;
    }
    
    server_ip = argv[optind];
    if (optind + 1 < argc) {
        port = (uint16_t)atoi(argv[optind + 1]);
    }
    
    printf("Heartbeat Client starting...\n");
    printf("Server: %s:%d\n", server_ip, port);
    printf("Client ID: '%s'\n", client_id);
    printf("Heartbeat interval: %d sec\n", interval_sec);
    printf("Max missing ACKs before server considered down: %d\n", MAX_MISSING_ACKS);
    if (max_count > 0) {
        printf("Heartbeat count: %d\n", max_count);
    } else {
        printf("Heartbeat count: infinite (press Ctrl+C to stop)\n");
    }
    printf("========================================\n");
    
    setup_signals();
    
    struct sockaddr_in server_addr;
    memset(&server_addr, 0, sizeof(server_addr));
    server_addr.sin_family = AF_INET;
    server_addr.sin_port = htons(port);
    
    if (inet_pton(AF_INET, server_ip, &server_addr.sin_addr) <= 0) {
        perror("inet_pton failed");
        fprintf(stderr, "Invalid server IP: %s\n", server_ip);
        return 1;
    }
    
    g_sockfd = create_socket();
    if (g_sockfd < 0) {
        fprintf(stderr, "Failed to create socket\n");
        return 1;
    }
    
    printf("\nClient ready, sending heartbeats...\n\n");
    
    while (g_running) {
        if (max_count > 0 && count >= max_count) {
            printf("\nReached max heartbeat count (%d), stopping...\n", max_count);
            break;
        }
        
        if (send_heartbeat(g_sockfd, &server_addr, client_id) < 0) {
            fprintf(stderr, "Failed to send heartbeat\n");
            g_missing_acks++;
        } else {
            count++;
        }
        
        int wait_loops = interval_sec * 10;
        for (int i = 0; i < wait_loops && g_running; i++) {
            struct sockaddr_in recv_addr;
            socklen_t addr_len = sizeof(recv_addr);
            
            int recv_ret = receive_ack(g_sockfd, &recv_addr, &addr_len, client_id);
            
            if (recv_ret < 0) {
                fprintf(stderr, "Error receiving ACK\n");
            }
            
            usleep(100000);
        }
        
        uint32_t expected_seq = g_seq_num - 1;
        if (g_last_acked_seq < expected_seq) {
            g_missing_acks++;
            printf("Warning: Missing ACK for Seq=%u (missing=%d, last_acked=%u)\n",
                   expected_seq, g_missing_acks, g_last_acked_seq);
            
            if (g_missing_acks >= MAX_MISSING_ACKS) {
                printf("\n*** ERROR: Server seems to be down! ***\n");
                printf("Missing %d consecutive ACKs (last_acked=%u, next_expected=%u)\n",
                       g_missing_acks, g_last_acked_seq, expected_seq);
                printf("Server is not responding. Exiting...\n");
                break;
            }
        }
    }
    
    printf("\nClient stopping...\n");
    printf("Total heartbeats sent: %d\n", count);
    printf("Last ACKed sequence: %u\n", g_last_acked_seq);
    printf("Missing ACKs at exit: %d\n", g_missing_acks);
    
    if (g_sockfd >= 0) {
        close(g_sockfd);
    }
    
    printf("Client stopped\n");
    
    return 0;
}
