#ifndef HEARTBEAT_IO_H
#define HEARTBEAT_IO_H

#include <netinet/in.h>
#include "heartbeat_protocol.h"

int heartbeat_io_create_socket(void);
int heartbeat_io_bind_socket(int sockfd, uint16_t port);
int heartbeat_io_receive(int sockfd, heartbeat_packet_t *packet, 
                          struct sockaddr_in *client_addr, socklen_t *addr_len);
int heartbeat_io_send(int sockfd, const heartbeat_packet_t *packet, 
                       const struct sockaddr_in *client_addr, socklen_t addr_len);
int heartbeat_io_validate_packet(const heartbeat_packet_t *packet);
void heartbeat_io_close_socket(int sockfd);

#endif
