#ifndef HTTP_ALERT_H
#define HTTP_ALERT_H

#include <stddef.h>
#include <time.h>

#define HTTP_ALERT_OK                      0
#define HTTP_ALERT_ERR_INVALID_ARG         -1
#define HTTP_ALERT_ERR_MEMORY              -2
#define HTTP_ALERT_ERR_SOCKET              -3
#define HTTP_ALERT_ERR_CONNECT             -4
#define HTTP_ALERT_ERR_CONNECT_TIMEOUT     -5
#define HTTP_ALERT_ERR_SEND                -6
#define HTTP_ALERT_ERR_RECV                -7
#define HTTP_ALERT_ERR_RECV_TIMEOUT        -8
#define HTTP_ALERT_ERR_SSL_INIT            -9
#define HTTP_ALERT_ERR_SSL_CONNECT         -10
#define HTTP_ALERT_ERR_URL_PARSE           -11
#define HTTP_ALERT_ERR_OVERALL_TIMEOUT     -12
#define HTTP_ALERT_ERR_ALL_RETRIES_FAILED  -13
#define HTTP_ALERT_ERR_RESOLVE             -14
#define HTTP_ALERT_ERR_HTTP_STATUS         -15

typedef enum {
    HTTP_METHOD_GET,
    HTTP_METHOD_POST,
    HTTP_METHOD_PUT
} HttpMethod;

typedef struct HttpHeader {
    char *key;
    char *value;
    struct HttpHeader *next;
} HttpHeader;

typedef struct HttpRequest {
    HttpMethod method;
    char *url;
    char *host;
    char *path;
    char *query;
    unsigned short port;
    int is_https;
    HttpHeader *headers;
    char *body;
    size_t body_len;
} HttpRequest;

typedef struct HttpResponse {
    int status_code;
    char *status_message;
    HttpHeader *headers;
    char *body;
    size_t body_len;
} HttpResponse;

typedef struct TimeoutConfig {
    int connect_timeout_sec;
    int connect_timeout_usec;
    int read_timeout_sec;
    int read_timeout_usec;
} TimeoutConfig;

typedef struct RetryConfig {
    int max_retries;
    int initial_delay_sec;
    int overall_timeout_sec;
} RetryConfig;

typedef struct AlertData {
    char *title;
    char *message;
    char *level;
    char *source;
    time_t timestamp;
} AlertData;

void http_alert_free_request(HttpRequest *req);
void http_alert_free_response(HttpResponse *resp);
void http_alert_free_headers(HttpHeader *headers);
HttpHeader *http_alert_copy_headers(const HttpHeader *headers);
HttpRequest *http_alert_create_request(void);
HttpResponse *http_alert_create_response(void);
int http_alert_add_header(HttpRequest *req, const char *key, const char *value);
int http_alert_set_body(HttpRequest *req, const char *body, size_t len);
int http_alert_parse_url(const char *url, HttpRequest *req);
const char *http_alert_strerror(int err_code);
void http_alert_log(const char *level, const char *fmt, ...);

#endif
