#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <unistd.h>
#include <fcntl.h>
#include <sys/socket.h>
#include <sys/types.h>
#include <netinet/in.h>
#include <arpa/inet.h>
#include <netdb.h>
#include <errno.h>
#include <sys/select.h>
#include "http_connection.h"

#ifdef WITH_SSL
#include <openssl/ssl.h>
#include <openssl/err.h>
#endif

static int set_nonblocking(int fd) {
    int flags = fcntl(fd, F_GETFL, 0);
    if (flags == -1) return -1;
    return fcntl(fd, F_SETFL, flags | O_NONBLOCK);
}

static int wait_for_connect(int sockfd, int timeout_sec, int timeout_usec) {
    fd_set wfds;
    struct timeval tv;
    int result;
    
    FD_ZERO(&wfds);
    FD_SET(sockfd, &wfds);
    
    tv.tv_sec = timeout_sec;
    tv.tv_usec = timeout_usec;
    
    result = select(sockfd + 1, NULL, &wfds, NULL, &tv);
    if (result <= 0) {
        return result == 0 ? HTTP_ALERT_ERR_CONNECT_TIMEOUT : HTTP_ALERT_ERR_CONNECT;
    }
    
    int error = 0;
    socklen_t len = sizeof(error);
    if (getsockopt(sockfd, SOL_SOCKET, SO_ERROR, &error, &len) == -1) {
        return HTTP_ALERT_ERR_CONNECT;
    }
    
    if (error != 0) {
        return HTTP_ALERT_ERR_CONNECT;
    }
    
    return HTTP_ALERT_OK;
}

int http_connection_init(HttpConnection *conn, int is_https) {
    if (conn == NULL) {
        return HTTP_ALERT_ERR_INVALID_ARG;
    }
    
    memset(conn, 0, sizeof(HttpConnection));
    conn->sockfd = -1;
    conn->is_https = is_https;
    conn->connected = 0;
    
#ifdef WITH_SSL
    if (is_https) {
        static int ssl_initialized = 0;
        if (!ssl_initialized) {
            SSL_load_error_strings();
            OpenSSL_add_ssl_algorithms();
            ssl_initialized = 1;
        }
        
        conn->ssl_ctx = SSL_CTX_new(TLS_client_method());
        if (conn->ssl_ctx == NULL) {
            return HTTP_ALERT_ERR_SSL_INIT;
        }
        
        SSL_CTX_set_verify(conn->ssl_ctx, SSL_VERIFY_NONE, NULL);
        SSL_CTX_set_options(conn->ssl_ctx, SSL_OP_ALL | SSL_OP_NO_SSLv2 | SSL_OP_NO_SSLv3);
    }
#else
    if (is_https) {
        http_alert_log("ERROR", "HTTPS support not compiled in.");
        http_alert_log("ERROR", "Install libssl-dev (sudo apt-get install libssl-dev) and rebuild.");
        return HTTP_ALERT_ERR_SSL_INIT;
    }
#endif
    
    return HTTP_ALERT_OK;
}

void http_connection_cleanup(HttpConnection *conn) {
    if (conn == NULL) return;
    
    http_connection_close(conn);
    
#ifdef WITH_SSL
    if (conn->ssl_ctx) {
        SSL_CTX_free(conn->ssl_ctx);
        conn->ssl_ctx = NULL;
    }
#endif
}

int http_connection_connect(HttpConnection *conn, const char *host, unsigned short port,
                             const TimeoutConfig *timeout) {
    if (conn == NULL || host == NULL || timeout == NULL) {
        return HTTP_ALERT_ERR_INVALID_ARG;
    }
    
    http_connection_close(conn);
    
    struct addrinfo hints, *res, *p;
    int gai_result;
    char port_str[8];
    
    snprintf(port_str, sizeof(port_str), "%d", port);
    
    memset(&hints, 0, sizeof(hints));
    hints.ai_family = AF_UNSPEC;
    hints.ai_socktype = SOCK_STREAM;
    
    gai_result = getaddrinfo(host, port_str, &hints, &res);
    if (gai_result != 0) {
        http_alert_log("ERROR", "Failed to resolve '%s': %s", host, gai_strerror(gai_result));
        return HTTP_ALERT_ERR_RESOLVE;
    }
    
    int sockfd = -1;
    int connect_result = HTTP_ALERT_ERR_CONNECT;
    
    for (p = res; p != NULL; p = p->ai_next) {
        sockfd = socket(p->ai_family, p->ai_socktype, p->ai_protocol);
        if (sockfd == -1) continue;
        
        if (set_nonblocking(sockfd) == -1) {
            close(sockfd);
            sockfd = -1;
            continue;
        }
        
        int ret = connect(sockfd, p->ai_addr, p->ai_addrlen);
        if (ret == 0) {
            connect_result = HTTP_ALERT_OK;
            break;
        } else if (errno == EINPROGRESS) {
            connect_result = wait_for_connect(sockfd, 
                timeout->connect_timeout_sec, 
                timeout->connect_timeout_usec);
            if (connect_result == HTTP_ALERT_OK) {
                break;
            }
        }
        
        close(sockfd);
        sockfd = -1;
    }
    
    freeaddrinfo(res);
    
    if (sockfd == -1) {
        return connect_result;
    }
    
    conn->sockfd = sockfd;
    conn->connected = 1;
    memcpy(&conn->timeout, timeout, sizeof(TimeoutConfig));
    
#ifdef WITH_SSL
    if (conn->is_https && conn->ssl_ctx) {
        conn->ssl = SSL_new(conn->ssl_ctx);
        if (conn->ssl == NULL) {
            http_connection_close(conn);
            return HTTP_ALERT_ERR_SSL_INIT;
        }
        
        if (SSL_set_fd(conn->ssl, sockfd) != 1) {
            http_connection_close(conn);
            return HTTP_ALERT_ERR_SSL_INIT;
        }
        
        int flags = fcntl(sockfd, F_GETFL, 0);
        fcntl(sockfd, F_SETFL, flags & ~O_NONBLOCK);
        
        struct timeval tv;
        tv.tv_sec = timeout->connect_timeout_sec;
        tv.tv_usec = timeout->connect_timeout_usec;
        setsockopt(sockfd, SOL_SOCKET, SO_RCVTIMEO, &tv, sizeof(tv));
        setsockopt(sockfd, SOL_SOCKET, SO_SNDTIMEO, &tv, sizeof(tv));
        
        int ssl_ret = SSL_connect(conn->ssl);
        if (ssl_ret != 1) {
            int ssl_err = SSL_get_error(conn->ssl, ssl_ret);
            http_alert_log("ERROR", "SSL_connect failed, error: %d", ssl_err);
            http_connection_close(conn);
            return HTTP_ALERT_ERR_SSL_CONNECT;
        }
    }
#endif
    
    return HTTP_ALERT_OK;
}

void http_connection_close(HttpConnection *conn) {
    if (conn == NULL) return;
    
#ifdef WITH_SSL
    if (conn->ssl) {
        SSL_shutdown(conn->ssl);
        SSL_free(conn->ssl);
        conn->ssl = NULL;
    }
#endif
    
    if (conn->sockfd != -1) {
        close(conn->sockfd);
        conn->sockfd = -1;
    }
    
    conn->connected = 0;
}

static int wait_for_read(int sockfd, int timeout_sec, int timeout_usec) {
    fd_set rfds;
    struct timeval tv;
    
    FD_ZERO(&rfds);
    FD_SET(sockfd, &rfds);
    
    tv.tv_sec = timeout_sec;
    tv.tv_usec = timeout_usec;
    
    int result = select(sockfd + 1, &rfds, NULL, NULL, &tv);
    if (result == 0) {
        return HTTP_ALERT_ERR_RECV_TIMEOUT;
    } else if (result < 0) {
        return HTTP_ALERT_ERR_RECV;
    }
    return HTTP_ALERT_OK;
}

static int wait_for_write(int sockfd, int timeout_sec, int timeout_usec) {
    fd_set wfds;
    struct timeval tv;
    
    FD_ZERO(&wfds);
    FD_SET(sockfd, &wfds);
    
    tv.tv_sec = timeout_sec;
    tv.tv_usec = timeout_usec;
    
    int result = select(sockfd + 1, NULL, &wfds, NULL, &tv);
    if (result == 0) {
        return HTTP_ALERT_ERR_SEND;
    } else if (result < 0) {
        return HTTP_ALERT_ERR_SEND;
    }
    return HTTP_ALERT_OK;
}

int http_connection_send(HttpConnection *conn, const char *data, size_t len) {
    if (conn == NULL || data == NULL || !conn->connected) {
        return HTTP_ALERT_ERR_INVALID_ARG;
    }
    
    size_t sent = 0;
    
#ifdef WITH_SSL
    if (conn->is_https && conn->ssl) {
        while (sent < len) {
            int ret = SSL_write(conn->ssl, data + sent, (int)(len - sent));
            if (ret <= 0) {
                int err = SSL_get_error(conn->ssl, ret);
                if (err == SSL_ERROR_WANT_WRITE || err == SSL_ERROR_WANT_READ) {
                    int wret = wait_for_write(conn->sockfd, 
                        conn->timeout.read_timeout_sec, 
                        conn->timeout.read_timeout_usec);
                    if (wret != HTTP_ALERT_OK) {
                        return wret;
                    }
                    continue;
                }
                return HTTP_ALERT_ERR_SEND;
            }
            sent += ret;
        }
        return (int)sent;
    }
#endif
    
    while (sent < len) {
        int wret = wait_for_write(conn->sockfd, 
            conn->timeout.read_timeout_sec, 
            conn->timeout.read_timeout_usec);
        if (wret != HTTP_ALERT_OK) {
            return wret;
        }
        
        ssize_t ret = send(conn->sockfd, data + sent, len - sent, MSG_NOSIGNAL);
        if (ret <= 0) {
            if (ret == 0 || errno == EPIPE || errno == ECONNRESET) {
                conn->connected = 0;
                return HTTP_ALERT_ERR_SEND;
            }
            if (errno == EAGAIN || errno == EWOULDBLOCK) {
                continue;
            }
            return HTTP_ALERT_ERR_SEND;
        }
        sent += ret;
    }
    
    return (int)sent;
}

int http_connection_recv(HttpConnection *conn, char *buf, size_t buf_len) {
    if (conn == NULL || buf == NULL || buf_len == 0 || !conn->connected) {
        return HTTP_ALERT_ERR_INVALID_ARG;
    }
    
    int rret = wait_for_read(conn->sockfd, 
        conn->timeout.read_timeout_sec, 
        conn->timeout.read_timeout_usec);
    if (rret != HTTP_ALERT_OK) {
        return rret;
    }
    
#ifdef WITH_SSL
    if (conn->is_https && conn->ssl) {
        int ret = SSL_read(conn->ssl, buf, (int)buf_len);
        if (ret <= 0) {
            int err = SSL_get_error(conn->ssl, ret);
            if (err == SSL_ERROR_WANT_READ || err == SSL_ERROR_WANT_WRITE) {
                return 0;
            }
            if (err == SSL_ERROR_ZERO_RETURN) {
                conn->connected = 0;
                return 0;
            }
            return HTTP_ALERT_ERR_RECV;
        }
        return ret;
    }
#endif
    
    ssize_t ret = recv(conn->sockfd, buf, buf_len, 0);
    if (ret == 0) {
        conn->connected = 0;
        return 0;
    } else if (ret < 0) {
        if (errno == EAGAIN || errno == EWOULDBLOCK) {
            return 0;
        }
        return HTTP_ALERT_ERR_RECV;
    }
    
    return (int)ret;
}

int http_connection_set_timeout(HttpConnection *conn, const TimeoutConfig *timeout) {
    if (conn == NULL || timeout == NULL) {
        return HTTP_ALERT_ERR_INVALID_ARG;
    }
    memcpy(&conn->timeout, timeout, sizeof(TimeoutConfig));
    return HTTP_ALERT_OK;
}

int http_connection_is_connected(const HttpConnection *conn) {
    if (conn == NULL) return 0;
    return conn->connected;
}
