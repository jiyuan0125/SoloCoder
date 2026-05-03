#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <unistd.h>
#include <sys/socket.h>
#include <sys/types.h>
#include <netinet/in.h>
#include <arpa/inet.h>
#include <errno.h>
#include "heartbeat_io.h"
#include "heartbeat_protocol.h"

int heartbeat_io_create_socket(void) {
    int sockfd = socket(AF_INET, SOCK_DGRAM, 0);
    if (sockfd < 0) {
        perror("socket creation failed");
        return -1;
    }
    return sockfd;
}

int heartbeat_io_bind_socket(int sockfd, uint16_t port) {
    struct sockaddr_in server_addr;
    memset(&server_addr, 0, sizeof(server_addr));
    server_addr.sin_family = AF_INET;
    server_addr.sin_addr.s_addr = htonl(INADDR_ANY);
    server_addr.sin_port = htons(port);
    
    int reuse = 1;
    if (setsockopt(sockfd, SOL_SOCKET, SO_REUSEADDR, &reuse, sizeof(reuse)) < 0) {
        perror("setsockopt SO_REUSEADDR failed");
        return -1;
    }
    
    if (bind(sockfd, (struct sockaddr *)&server_addr, sizeof(server_addr)) < 0) {
        perror("bind failed");
        return -1;
    }
    return 0;
}

int heartbeat_io_receive(int sockfd, heartbeat_packet_t *packet, 
                          struct sockaddr_in *client_addr, socklen_t *addr_len) {
    ssize_t n = recvfrom(sockfd, packet, HEARTBEAT_PACKET_SIZE, 0,
                          (struct sockaddr *)client_addr, addr_len);
    if (n < 0) {
        if (errno == EAGAIN || errno == EWOULDBLOCK) {
            return 0;
        }
        perror("recvfrom failed");
        return -1;
    }
    if (n != HEARTBEAT_PACKET_SIZE) {
        fprintf(stderr, "Invalid packet size: %zd (expected %zu)\n", n, HEARTBEAT_PACKET_SIZE);
        return -1;
    }
    return 1;
}

int heartbeat_io_send(int sockfd, const heartbeat_packet_t *packet, 
                       const struct sockaddr_in *client_addr, socklen_t addr_len) {
    if (HEARTBEAT_PACKET_SIZE > HEARTBEAT_MAX_MTU) {
        fprintf(stderr, "Packet size (%zu) exceeds MTU limit (%d)\n", 
                HEARTBEAT_PACKET_SIZE, HEARTBEAT_MAX_MTU);
        return -1;
    }
    
    ssize_t n = sendto(sockfd, packet, HEARTBEAT_PACKET_SIZE, 0,
                        (struct sockaddr *)client_addr, addr_len);
    if (n < 0) {
        perror("sendto failed");
        return -1;
    }
    if (n != HEARTBEAT_PACKET_SIZE) {
        fprintf(stderr, "Partial send: %zd/%zu bytes\n", n, HEARTBEAT_PACKET_SIZE);
        return -1;
    }
    return 0;
}

int heartbeat_io_validate_packet(const heartbeat_packet_t *packet) {
    if (packet->magic != htonl(HEARTBEAT_MAGIC)) {
        return 0;
    }
    return 1;
}

void heartbeat_io_close_socket(int sockfd) {
    if (sockfd >= 0) {
        close(sockfd);
    }
}
