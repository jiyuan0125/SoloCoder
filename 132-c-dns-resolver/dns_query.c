#include "dns_query.h"
#include <string.h>
#include <stdlib.h>
#include <stdio.h>
#include <unistd.h>
#include <sys/socket.h>
#include <netinet/in.h>
#include <arpa/inet.h>
#include <sys/select.h>
#include <errno.h>
#include <fcntl.h>

int dns_resolver_init(dns_resolver_t *resolver) {
    memset(resolver, 0, sizeof(*resolver));
    resolver->timeout_seconds = DNS_DEFAULT_TIMEOUT;
    resolver->max_retries = DNS_DEFAULT_MAX_RETRIES;
    resolver->server_count = 0;

    dns_resolver_add_server(resolver, "8.8.8.8", DNS_PORT);
    dns_resolver_add_server(resolver, "8.8.4.4", DNS_PORT);
    dns_resolver_add_server(resolver, "114.114.114.114", DNS_PORT);

    return 0;
}

int dns_resolver_add_server(dns_resolver_t *resolver, const char *server_ip, int port) {
    if (resolver->server_count >= DNS_MAX_SERVERS) {
        return -1;
    }

    strncpy(resolver->servers[resolver->server_count].server_ip,
            server_ip,
            sizeof(resolver->servers[resolver->server_count].server_ip) - 1);
    resolver->servers[resolver->server_count].port = port;
    resolver->server_count++;

    return 0;
}

void dns_resolver_set_timeout(dns_resolver_t *resolver, int seconds) {
    resolver->timeout_seconds = seconds;
}

void dns_resolver_set_retries(dns_resolver_t *resolver, int retries) {
    resolver->max_retries = retries;
}

static int set_socket_nonblocking(int sockfd) {
    int flags = fcntl(sockfd, F_GETFL, 0);
    if (flags == -1) return -1;
    if (fcntl(sockfd, F_SETFL, flags | O_NONBLOCK) == -1) return -1;
    return 0;
}

static int wait_for_socket(int sockfd, int timeout_seconds, int is_write) {
    fd_set fds;
    struct timeval tv;

    FD_ZERO(&fds);
    FD_SET(sockfd, &fds);

    tv.tv_sec = timeout_seconds;
    tv.tv_usec = 0;

    if (is_write) {
        return select(sockfd + 1, NULL, &fds, NULL, &tv);
    } else {
        return select(sockfd + 1, &fds, NULL, NULL, &tv);
    }
}

static int connect_with_timeout(int sockfd, const struct sockaddr *addr,
                                 socklen_t addrlen, int timeout_seconds) {
    set_socket_nonblocking(sockfd);

    int result = connect(sockfd, addr, addrlen);
    if (result == 0) {
        return 0;
    }
    if (result == -1 && errno != EINPROGRESS) {
        return -1;
    }

    result = wait_for_socket(sockfd, timeout_seconds, 1);
    if (result <= 0) {
        return result == 0 ? -2 : -1;
    }

    int error = 0;
    socklen_t len = sizeof(error);
    if (getsockopt(sockfd, SOL_SOCKET, SO_ERROR, &error, &len) == -1) {
        return -1;
    }

    if (error != 0) {
        errno = error;
        return -1;
    }

    return 0;
}

int dns_send_query_udp(dns_resolver_t *resolver, int server_index,
                        const uint8_t *query, size_t query_len,
                        uint8_t *response, size_t response_size, size_t *response_len,
                        int timeout_seconds) {
    if (server_index < 0 || server_index >= resolver->server_count) {
        return DNS_QUERY_ERROR;
    }

    dns_server_t *server = &resolver->servers[server_index];

    int sockfd = socket(AF_INET, SOCK_DGRAM, 0);
    if (sockfd < 0) {
        return DNS_QUERY_ERROR;
    }

    struct sockaddr_in server_addr;
    memset(&server_addr, 0, sizeof(server_addr));
    server_addr.sin_family = AF_INET;
    server_addr.sin_port = htons(server->port);

    if (inet_pton(AF_INET, server->server_ip, &server_addr.sin_addr) <= 0) {
        close(sockfd);
        return DNS_QUERY_ERROR;
    }

    ssize_t sent = sendto(sockfd, query, query_len, 0,
                          (struct sockaddr *)&server_addr, sizeof(server_addr));
    if (sent != (ssize_t)query_len) {
        close(sockfd);
        return DNS_QUERY_ERROR;
    }

    struct timeval tv;
    tv.tv_sec = timeout_seconds;
    tv.tv_usec = 0;
    setsockopt(sockfd, SOL_SOCKET, SO_RCVTIMEO, &tv, sizeof(tv));

    struct sockaddr_in from_addr;
    socklen_t from_len = sizeof(from_addr);

    ssize_t received = recvfrom(sockfd, response, response_size, 0,
                                 (struct sockaddr *)&from_addr, &from_len);

    close(sockfd);

    if (received < 0) {
        if (errno == EAGAIN || errno == EWOULDBLOCK) {
            return DNS_QUERY_TIMEOUT;
        }
        return DNS_QUERY_ERROR;
    }

    if (received == 0) {
        return DNS_QUERY_ERROR;
    }

    *response_len = (size_t)received;
    return DNS_QUERY_SUCCESS;
}

int dns_send_query_tcp(dns_resolver_t *resolver, int server_index,
                        const uint8_t *query, size_t query_len,
                        uint8_t *response, size_t response_size, size_t *response_len,
                        int timeout_seconds) {
    if (server_index < 0 || server_index >= resolver->server_count) {
        return DNS_QUERY_ERROR;
    }

    dns_server_t *server = &resolver->servers[server_index];

    int sockfd = socket(AF_INET, SOCK_STREAM, 0);
    if (sockfd < 0) {
        return DNS_QUERY_ERROR;
    }

    struct sockaddr_in server_addr;
    memset(&server_addr, 0, sizeof(server_addr));
    server_addr.sin_family = AF_INET;
    server_addr.sin_port = htons(server->port);

    if (inet_pton(AF_INET, server->server_ip, &server_addr.sin_addr) <= 0) {
        close(sockfd);
        return DNS_QUERY_ERROR;
    }

    int connect_result = connect_with_timeout(sockfd,
                                               (struct sockaddr *)&server_addr,
                                               sizeof(server_addr),
                                               timeout_seconds);
    if (connect_result != 0) {
        close(sockfd);
        return connect_result == -2 ? DNS_QUERY_TIMEOUT : DNS_QUERY_ERROR;
    }

    uint8_t tcp_query[DNS_MAX_PACKET_SIZE + 2];
    if (query_len + 2 > sizeof(tcp_query)) {
        close(sockfd);
        return DNS_QUERY_ERROR;
    }

    tcp_query[0] = (query_len >> 8) & 0xFF;
    tcp_query[1] = query_len & 0xFF;
    memcpy(tcp_query + 2, query, query_len);

    size_t total_sent = 0;
    size_t total_to_send = query_len + 2;

    while (total_sent < total_to_send) {
        int select_result = wait_for_socket(sockfd, timeout_seconds, 1);
        if (select_result <= 0) {
            close(sockfd);
            return select_result == 0 ? DNS_QUERY_TIMEOUT : DNS_QUERY_ERROR;
        }

        ssize_t sent = send(sockfd, tcp_query + total_sent,
                            total_to_send - total_sent, 0);
        if (sent <= 0) {
            close(sockfd);
            return DNS_QUERY_ERROR;
        }
        total_sent += (size_t)sent;
    }

    uint8_t tcp_header[2];
    size_t header_received = 0;

    while (header_received < 2) {
        int select_result = wait_for_socket(sockfd, timeout_seconds, 0);
        if (select_result <= 0) {
            close(sockfd);
            return select_result == 0 ? DNS_QUERY_TIMEOUT : DNS_QUERY_ERROR;
        }

        ssize_t received = recv(sockfd, tcp_header + header_received,
                                2 - header_received, 0);
        if (received <= 0) {
            close(sockfd);
            return DNS_QUERY_ERROR;
        }
        header_received += (size_t)received;
    }

    uint16_t response_len_16 = (tcp_header[0] << 8) | tcp_header[1];

    if (response_len_16 > response_size) {
        close(sockfd);
        return DNS_QUERY_ERROR;
    }

    size_t data_received = 0;
    while (data_received < response_len_16) {
        int select_result = wait_for_socket(sockfd, timeout_seconds, 0);
        if (select_result <= 0) {
            close(sockfd);
            return select_result == 0 ? DNS_QUERY_TIMEOUT : DNS_QUERY_ERROR;
        }

        ssize_t received = recv(sockfd, response + data_received,
                                response_len_16 - data_received, 0);
        if (received <= 0) {
            close(sockfd);
            return DNS_QUERY_ERROR;
        }
        data_received += (size_t)received;
    }

    close(sockfd);
    *response_len = data_received;
    return DNS_QUERY_SUCCESS;
}

uint16_t dns_generate_id(void) {
    static int initialized = 0;
    if (!initialized) {
        srand((unsigned int)time(NULL));
        initialized = 1;
    }
    return (uint16_t)(rand() & 0xFFFF);
}

int dns_check_truncated(const uint8_t *packet, size_t packet_len) {
    if (packet_len < 4) {
        return 0;
    }
    uint16_t flags = (packet[2] << 8) | packet[3];
    return (flags & DNS_FLAG_TC) ? 1 : 0;
}

int dns_check_response_id(const uint8_t *packet, size_t packet_len, uint16_t expected_id) {
    if (packet_len < 2) {
        return 0;
    }
    uint16_t actual_id = (packet[0] << 8) | packet[1];
    return (actual_id == expected_id) ? 1 : 0;
}

int dns_get_response_rcode(const uint8_t *packet, size_t packet_len) {
    if (packet_len < 4) {
        return DNS_RCODE_FORMAT_ERROR;
    }
    return (int)(packet[3] & 0x0F);
}

static int is_valid_response(const uint8_t *packet, size_t packet_len, uint16_t query_id) {
    if (packet_len < DNS_HEADER_SIZE) {
        return 0;
    }
    if (!dns_check_response_id(packet, packet_len, query_id)) {
        return 0;
    }
    uint16_t flags = (packet[2] << 8) | packet[3];
    if (!(flags & DNS_FLAG_QR)) {
        return 0;
    }
    return 1;
}

int dns_query_raw(dns_resolver_t *resolver,
                  const char *name, uint16_t qtype,
                  uint8_t *response, size_t response_size, size_t *response_len) {
    uint16_t query_id = dns_generate_id();

    uint8_t query[DNS_MAX_PACKET_SIZE];
    int query_len = dns_construct_query(query, sizeof(query), query_id,
                                         name, qtype, DNS_CLASS_IN);
    if (query_len < 0) {
        return DNS_QUERY_ERROR;
    }

    int retries = resolver->max_retries;
    int server_index = 0;
    int retry_delay = 1;

    for (int i = 0; i <= retries; i++) {
        if (i > 0) {
            sleep(retry_delay);
            retry_delay *= 2;
        }

        server_index = i % resolver->server_count;

        int result = dns_send_query_udp(resolver, server_index,
                                         query, (size_t)query_len,
                                         response, response_size, response_len,
                                         resolver->timeout_seconds);

        if (result == DNS_QUERY_TIMEOUT) {
            continue;
        }

        if (result != DNS_QUERY_SUCCESS) {
            continue;
        }

        if (!is_valid_response(response, *response_len, query_id)) {
            continue;
        }

        int rcode = dns_get_response_rcode(response, *response_len);
        if (rcode == DNS_RCODE_NAME_ERROR) {
            return DNS_QUERY_NAME_NOT_FOUND;
        } else if (rcode != DNS_RCODE_NO_ERROR) {
            continue;
        }

        if (dns_check_truncated(response, *response_len)) {
            result = dns_send_query_tcp(resolver, server_index,
                                         query, (size_t)query_len,
                                         response, response_size, response_len,
                                         resolver->timeout_seconds);

            if (result == DNS_QUERY_SUCCESS &&
                is_valid_response(response, *response_len, query_id)) {
                return DNS_QUERY_SUCCESS;
            }

            return DNS_QUERY_TRUNCATED;
        }

        return DNS_QUERY_SUCCESS;
    }

    return DNS_QUERY_TIMEOUT;
}

int dns_query_message(dns_resolver_t *resolver,
                      const char *name, uint16_t qtype,
                      dns_message_t *msg, uint8_t *raw_packet, size_t *raw_packet_len) {
    uint8_t response[DNS_MAX_PACKET_SIZE];
    size_t response_len = 0;

    int result = dns_query_raw(resolver, name, qtype, response, sizeof(response), &response_len);

    if (result != DNS_QUERY_SUCCESS) {
        return result;
    }

    if (raw_packet && raw_packet_len) {
        if (*raw_packet_len >= response_len) {
            memcpy(raw_packet, response, response_len);
            *raw_packet_len = response_len;
        }
    }

    if (dns_parse_response(response, response_len, msg) != 0) {
        return DNS_QUERY_ERROR;
    }

    return DNS_QUERY_SUCCESS;
}
