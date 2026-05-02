#include "http_server.h"
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <unistd.h>
#include <sys/socket.h>
#include <netinet/in.h>
#include <arpa/inet.h>
#include <sys/select.h>
#include <sys/time.h>
#include <time.h>
#include <errno.h>
#include <signal.h>
#include <inttypes.h>

#define READ_BUFFER_SIZE (MAX_HEADERS_SIZE + MAX_BODY_SIZE + 8192)

static volatile int running = 1;

void sigint_handler(int sig) {
    (void)sig;
    running = 0;
}

int server_init(ServerStats *stats, uint16_t port) {
    int sockfd = socket(AF_INET, SOCK_STREAM, 0);
    if (sockfd < 0) {
        perror("socket");
        return -1;
    }
    
    int opt = 1;
    if (setsockopt(sockfd, SOL_SOCKET, SO_REUSEADDR, &opt, sizeof(opt)) < 0) {
        perror("setsockopt");
        close(sockfd);
        return -1;
    }
    
    struct sockaddr_in server_addr;
    memset(&server_addr, 0, sizeof(server_addr));
    server_addr.sin_family = AF_INET;
    server_addr.sin_addr.s_addr = INADDR_ANY;
    server_addr.sin_port = htons(port);
    
    if (bind(sockfd, (struct sockaddr *)&server_addr, sizeof(server_addr)) < 0) {
        perror("bind");
        close(sockfd);
        return -1;
    }
    
    if (listen(sockfd, 128) < 0) {
        perror("listen");
        close(sockfd);
        return -1;
    }
    
    gettimeofday(&stats->start_time, NULL);
    stats->request_count = 0;
    stats->active_connections = 0;
    stats->port = port;
    
    return sockfd;
}

void server_cleanup(int sockfd) {
    if (sockfd >= 0) {
        close(sockfd);
    }
}

int server_accept_connection(int sockfd, struct sockaddr_in *client_addr, socklen_t *addr_len) {
    int client_fd = accept(sockfd, (struct sockaddr *)client_addr, addr_len);
    return client_fd;
}

int check_data_available(int sockfd, int timeout_ms) {
    fd_set readfds;
    struct timeval tv;
    
    FD_ZERO(&readfds);
    FD_SET(sockfd, &readfds);
    
    tv.tv_sec = timeout_ms / 1000;
    tv.tv_usec = (timeout_ms % 1000) * 1000;
    
    int result = select(sockfd + 1, &readfds, NULL, NULL, &tv);
    return result > 0;
}

int read_from_socket(int sockfd, char *buffer, size_t *buffer_pos, size_t buffer_max) {
    if (*buffer_pos >= buffer_max) {
        return -1;
    }
    
    ssize_t n = recv(sockfd, buffer + *buffer_pos, buffer_max - *buffer_pos, 0);
    
    if (n < 0) {
        if (errno == EINTR || errno == EAGAIN || errno == EWOULDBLOCK) {
            return 0;
        }
        return -1;
    }
    
    if (n == 0) {
        return 0;
    }
    
    *buffer_pos += (size_t)n;
    buffer[*buffer_pos] = '\0';
    return 1;
}

int parse_single_request_from_buffer(const char *buffer, size_t buffer_len, 
                                      HttpRequest *req, size_t *consumed, int *keep_alive) {
    size_t header_consumed = 0;
    int parse_result = http_parse_headers(buffer, buffer_len, req, &header_consumed);
    
    if (parse_result == -2) {
        return -431;
    } else if (parse_result < 0) {
        return -1;
    } else if (parse_result == 1) {
        return 1;
    }
    
    if (req->content_length > MAX_BODY_SIZE) {
        return -413;
    }
    
    size_t body_start = header_consumed;
    size_t body_available = buffer_len - header_consumed;
    
    if (req->method == HTTP_POST && req->content_length > 0) {
        if (body_available >= req->content_length) {
            req->body = (char *)malloc(req->content_length);
            if (!req->body) return -1;
            memcpy(req->body, buffer + body_start, req->content_length);
            req->body_length = req->content_length;
            *consumed = header_consumed + req->content_length;
        } else {
            return 1;
        }
    } else {
        *consumed = header_consumed;
    }
    
    *keep_alive = req->keep_alive;
    return 0;
}

int server_send_response(int sockfd, const HttpResponse *resp) {
    size_t response_len;
    char *response = http_response_serialize(resp, &response_len);
    
    if (!response) {
        return -1;
    }
    
    size_t total_sent = 0;
    while (total_sent < response_len) {
        ssize_t n = send(sockfd, response + total_sent, response_len - total_sent, 0);
        
        if (n < 0) {
            if (errno == EINTR) continue;
            free(response);
            return -1;
        }
        
        total_sent += (size_t)n;
    }
    
    free(response);
    return 0;
}

void server_log_request(const HttpRequest *req, const HttpResponse *resp, struct timeval *start) {
    struct timeval end;
    gettimeofday(&end, NULL);
    
    long long elapsed_us = (long long)(end.tv_sec - start->tv_sec) * 1000000LL +
                           (end.tv_usec - start->tv_usec);
    
    time_t now = time(NULL);
    char time_buf[64];
    struct tm *tm_now = localtime(&now);
    strftime(time_buf, sizeof(time_buf), "%Y-%m-%d %H:%M:%S", tm_now);
    
    const char *method_str = req->method_str[0] ? req->method_str : "UNKNOWN";
    
    fprintf(stderr, "[%s] %s %s %d %lldus\n",
            time_buf, method_str, req->path,
            http_status_to_code(resp->status), elapsed_us);
}

void api_status_handler(HttpResponse *resp, const ServerStats *stats) {
    struct timeval now;
    gettimeofday(&now, NULL);
    
    uint64_t uptime_sec = (uint64_t)(now.tv_sec - stats->start_time.tv_sec);
    
    char json[1024];
    snprintf(json, sizeof(json),
             "{\n"
             "  \"server_name\": \"%s\",\n"
             "  \"port\": %u,\n"
             "  \"uptime_seconds\": %" PRIu64 ",\n"
             "  \"total_requests\": %" PRIu64 ",\n"
             "  \"active_connections\": %u\n"
             "}\n",
             SERVER_NAME, stats->port, uptime_sec,
             stats->request_count, stats->active_connections);
    
    http_response_set_status(resp, HTTP_OK);
    http_response_set_body(resp, json, strlen(json));
    
    char cl_buf[32];
    snprintf(cl_buf, sizeof(cl_buf), "%zu", resp->body_length);
    
    http_response_add_header(resp, "Content-Type", "application/json");
    http_response_add_header(resp, "Content-Length", cl_buf);
}

void build_error_response(HttpResponse *resp, HttpStatus status) {
    http_response_set_status(resp, status);
    
    const char *text = http_status_to_text(status);
    int code = http_status_to_code(status);
    
    char body[1024];
    snprintf(body, sizeof(body),
             "<html><head><title>%d %s</title></head>"
             "<body><h1>%d %s</h1></body></html>\n",
             code, text, code, text);
    
    http_response_set_body(resp, body, strlen(body));
    
    char cl_buf[32];
    snprintf(cl_buf, sizeof(cl_buf), "%zu", resp->body_length);
    
    http_response_add_header(resp, "Content-Type", "text/html");
    http_response_add_header(resp, "Content-Length", cl_buf);
}

void add_standard_headers(HttpResponse *resp, int keep_alive) {
    char date_buf[128];
    format_rfc1123_date(date_buf, sizeof(date_buf));
    
    http_response_add_header(resp, "Server", SERVER_NAME);
    http_response_add_header(resp, "Date", date_buf);
    
    if (keep_alive) {
        http_response_add_header(resp, "Connection", "keep-alive");
        http_response_add_header(resp, "Keep-Alive", "timeout=5, max=100");
    } else {
        http_response_add_header(resp, "Connection", "close");
    }
}

void process_single_request(ClientContext *ctx, HttpRequest *req, HttpResponse *resp, int read_result, int *keep_alive) {
    ctx->stats->request_count++;
    
    if (read_result == -431) {
        build_error_response(resp, HTTP_REQUEST_HEADER_FIELDS_TOO_LARGE);
        *keep_alive = 0;
    } else if (read_result == -413) {
        build_error_response(resp, HTTP_REQUEST_ENTITY_TOO_LARGE);
        *keep_alive = 0;
    } else if (read_result < 0) {
        build_error_response(resp, HTTP_BAD_REQUEST);
        *keep_alive = 0;
    } else if (req->method == HTTP_UNKNOWN) {
        build_error_response(resp, HTTP_NOT_IMPLEMENTED);
        *keep_alive = 0;
    } else {
        if (strcmp(req->path, "/api/status") == 0 && req->method == HTTP_GET) {
            api_status_handler(resp, ctx->stats);
        } else {
            int file_result = file_handler_serve(resp, ctx->root_dir, req->path);
            
            if (file_result != 0) {
                if (resp->status == HTTP_NOT_FOUND) {
                    build_error_response(resp, HTTP_NOT_FOUND);
                } else if (resp->status == HTTP_FORBIDDEN) {
                    build_error_response(resp, HTTP_FORBIDDEN);
                } else {
                    build_error_response(resp, HTTP_INTERNAL_SERVER_ERROR);
                }
            }
        }
    }
}

void server_handle_client(ClientContext *ctx) {
    int client_fd = ctx->sockfd;
    int keep_alive = 1;
    int consecutive_timeouts = 0;
    
    char *read_buffer = (char *)malloc(READ_BUFFER_SIZE);
    if (!read_buffer) {
        close(client_fd);
        return;
    }
    size_t buffer_pos = 0;
    
    ctx->stats->active_connections++;
    
    while (keep_alive && running && consecutive_timeouts < 2) {
        int has_data = (buffer_pos > 0);
        if (!has_data) {
            has_data = check_data_available(client_fd, 5000);
        }
        
        if (!has_data) {
            consecutive_timeouts++;
            continue;
        }
        consecutive_timeouts = 0;
        
        if (buffer_pos == 0) {
            int read_result = read_from_socket(client_fd, read_buffer, &buffer_pos, READ_BUFFER_SIZE - 1);
            if (read_result <= 0) {
                break;
            }
        }
        
        const char *end_of_headers = strstr(read_buffer, "\r\n\r\n");
        if (!end_of_headers) {
            if (buffer_pos > MAX_HEADERS_SIZE) {
                struct timeval request_start;
                gettimeofday(&request_start, NULL);
                
                HttpRequest req;
                HttpResponse resp;
                http_response_init(&resp);
                http_request_init(&req);
                
                build_error_response(&resp, HTTP_REQUEST_HEADER_FIELDS_TOO_LARGE);
                add_standard_headers(&resp, 0);
                server_send_response(client_fd, &resp);
                server_log_request(&req, &resp, &request_start);
                
                http_response_free(&resp);
                buffer_pos = 0;
                keep_alive = 0;
                break;
            }
            
            int read_result = read_from_socket(client_fd, read_buffer, &buffer_pos, READ_BUFFER_SIZE - 1);
            if (read_result < 0) {
                break;
            }
            continue;
        }
        
        struct timeval request_start;
        gettimeofday(&request_start, NULL);
        
        HttpRequest req;
        HttpResponse resp;
        http_request_init(&req);
        http_response_init(&resp);
        
        size_t consumed = 0;
        int parse_result = parse_single_request_from_buffer(read_buffer, buffer_pos, &req, &consumed, &keep_alive);
        
        if (parse_result == 1) {
            if (req.method == HTTP_POST && req.content_length > 0) {
                size_t body_available = buffer_pos - consumed;
                size_t remaining = req.content_length - body_available;
                
                req.body = (char *)malloc(req.content_length);
                if (!req.body) {
                    http_request_free(&req);
                    break;
                }
                
                if (body_available > 0) {
                    memcpy(req.body, read_buffer + consumed, body_available);
                }
                
                size_t total_read = body_available;
                while (total_read < req.content_length) {
                    if (!check_data_available(client_fd, 5000)) {
                        http_request_free(&req);
                        keep_alive = 0;
                        goto cleanup;
                    }
                    
                    ssize_t n = recv(client_fd, req.body + total_read, remaining, 0);
                    if (n <= 0) {
                        http_request_free(&req);
                        keep_alive = 0;
                        goto cleanup;
                    }
                    total_read += (size_t)n;
                    remaining -= (size_t)n;
                }
                
                req.body_length = req.content_length;
                buffer_pos = 0;
            }
        } else if (parse_result == 0) {
            if (consumed < buffer_pos) {
                memmove(read_buffer, read_buffer + consumed, buffer_pos - consumed);
                buffer_pos -= consumed;
                read_buffer[buffer_pos] = '\0';
            } else {
                buffer_pos = 0;
            }
        }
        
        process_single_request(ctx, &req, &resp, parse_result, &keep_alive);
        
        add_standard_headers(&resp, keep_alive);
        server_send_response(client_fd, &resp);
        server_log_request(&req, &resp, &request_start);
        
cleanup:
        http_request_free(&req);
        http_response_free(&resp);
    }
    
    ctx->stats->active_connections--;
    free(read_buffer);
    close(client_fd);
}

int main(int argc, char *argv[]) {
    uint16_t port = 8080;
    char *root_dir = "./public";
    
    if (argc >= 2) {
        port = (uint16_t)atoi(argv[1]);
    }
    if (argc >= 3) {
        root_dir = argv[2];
    }
    
    ServerStats stats;
    int server_fd = server_init(&stats, port);
    if (server_fd < 0) {
        fprintf(stderr, "Failed to initialize server\n");
        return 1;
    }
    
    signal(SIGINT, sigint_handler);
    signal(SIGPIPE, SIG_IGN);
    
    fprintf(stderr, "Server starting on port %d, root directory: %s\n", port, root_dir);
    fprintf(stderr, "Press Ctrl+C to stop\n");
    
    while (running) {
        struct sockaddr_in client_addr;
        socklen_t addr_len = sizeof(client_addr);
        
        if (!check_data_available(server_fd, 1000)) {
            continue;
        }
        
        int client_fd = server_accept_connection(server_fd, &client_addr, &addr_len);
        if (client_fd < 0) {
            if (errno == EINTR) continue;
            perror("accept");
            continue;
        }
        
        char client_ip[INET_ADDRSTRLEN];
        inet_ntop(AF_INET, &client_addr.sin_addr, client_ip, sizeof(client_ip));
        fprintf(stderr, "New connection from %s:%d\n", client_ip, ntohs(client_addr.sin_port));
        
        ClientContext ctx;
        ctx.sockfd = client_fd;
        ctx.addr = client_addr;
        ctx.addr_len = addr_len;
        ctx.stats = &stats;
        ctx.root_dir = root_dir;
        
        server_handle_client(&ctx);
        
        fprintf(stderr, "Connection closed from %s:%d\n", client_ip, ntohs(client_addr.sin_port));
    }
    
    fprintf(stderr, "\nShutting down server...\n");
    server_cleanup(server_fd);
    
    return 0;
}
