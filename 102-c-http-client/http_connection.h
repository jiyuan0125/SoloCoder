#ifndef HTTP_CONNECTION_H
#define HTTP_CONNECTION_H

#include "http_alert.h"

#define HTTP_CONN_BUF_SIZE 4096

typedef struct HttpConnection HttpConnection;

struct HttpConnection {
    int sockfd;
    int is_https;
    void *ssl_ctx;
    void *ssl;
    int connected;
    TimeoutConfig timeout;
};

int http_connection_init(HttpConnection *conn, int is_https);
void http_connection_cleanup(HttpConnection *conn);
int http_connection_connect(HttpConnection *conn, const char *host, unsigned short port,
                             const TimeoutConfig *timeout);
void http_connection_close(HttpConnection *conn);
int http_connection_send(HttpConnection *conn, const char *data, size_t len);
int http_connection_recv(HttpConnection *conn, char *buf, size_t buf_len);
int http_connection_set_timeout(HttpConnection *conn, const TimeoutConfig *timeout);
int http_connection_is_connected(const HttpConnection *conn);

#endif
