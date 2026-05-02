#ifndef HTTP_SERVER_H
#define HTTP_SERVER_H

#include <stddef.h>
#include <stdint.h>
#include <sys/socket.h>
#include <netinet/in.h>
#include <sys/time.h>

#define MAX_HEADERS_SIZE 8192
#define MAX_BODY_SIZE 1048576
#define MAX_HEADER_COUNT 64
#define SERVER_NAME "SimpleHTTP/1.0"
#define BUFFER_SIZE 4096

typedef enum {
    HTTP_GET,
    HTTP_POST,
    HTTP_DELETE,
    HTTP_UNKNOWN
} HttpMethod;

typedef enum {
    HTTP_CONTINUE = 100,
    HTTP_OK = 200,
    HTTP_BAD_REQUEST = 400,
    HTTP_FORBIDDEN = 403,
    HTTP_NOT_FOUND = 404,
    HTTP_METHOD_NOT_ALLOWED = 405,
    HTTP_REQUEST_ENTITY_TOO_LARGE = 413,
    HTTP_REQUEST_URI_TOO_LONG = 414,
    HTTP_REQUEST_HEADER_FIELDS_TOO_LARGE = 431,
    HTTP_INTERNAL_SERVER_ERROR = 500,
    HTTP_NOT_IMPLEMENTED = 501
} HttpStatus;

typedef struct {
    char name[256];
    char value[1024];
} HttpHeader;

typedef struct {
    HttpMethod method;
    char method_str[16];
    char path[4096];
    char query_string[4096];
    char version[16];
    HttpHeader headers[MAX_HEADER_COUNT];
    int header_count;
    char *body;
    size_t body_length;
    size_t content_length;
    int keep_alive;
} HttpRequest;

typedef struct {
    HttpStatus status;
    char status_text[128];
    HttpHeader headers[MAX_HEADER_COUNT];
    int header_count;
    char *body;
    size_t body_length;
} HttpResponse;

typedef struct {
    struct timeval start_time;
    uint64_t request_count;
    uint32_t active_connections;
    uint16_t port;
} ServerStats;

typedef struct {
    int sockfd;
    struct sockaddr_in addr;
    socklen_t addr_len;
    ServerStats *stats;
    char *root_dir;
} ClientContext;

void http_request_init(HttpRequest *req);
void http_request_free(HttpRequest *req);
int http_parse_request_line(const char *line, HttpRequest *req);
int http_parse_header(const char *line, HttpRequest *req);
int http_parse_headers(const char *data, size_t len, HttpRequest *req, size_t *consumed);
const char *http_method_to_string(HttpMethod method);
const char *http_status_to_text(HttpStatus status);
int http_status_to_code(HttpStatus status);

void http_response_init(HttpResponse *resp);
void http_response_free(HttpResponse *resp);
void http_response_set_status(HttpResponse *resp, HttpStatus status);
void http_response_add_header(HttpResponse *resp, const char *name, const char *value);
void http_response_set_body(HttpResponse *resp, const char *body, size_t length);
const char *http_response_get_header(const HttpResponse *resp, const char *name);
char *http_response_serialize(const HttpResponse *resp, size_t *out_length);
char *http_response_build_status_line(const HttpResponse *resp);

const char *get_content_type(const char *path);
int resolve_path(const char *root_dir, const char *url_path, char *resolved_path, size_t max_len);
int is_directory_traversal(const char *path);
int file_handler_serve(HttpResponse *resp, const char *root_dir, const char *url_path);

int server_init(ServerStats *stats, uint16_t port);
void server_cleanup(int sockfd);
int server_accept_connection(int sockfd, struct sockaddr_in *client_addr, socklen_t *addr_len);
void server_handle_client(ClientContext *ctx);
int server_read_request(int sockfd, HttpRequest *req, int *keep_alive);
int server_send_response(int sockfd, const HttpResponse *resp);
void server_log_request(const HttpRequest *req, const HttpResponse *resp, struct timeval *start);
void format_rfc1123_date(char *buffer, size_t len);
void api_status_handler(HttpResponse *resp, const ServerStats *stats);

#endif
