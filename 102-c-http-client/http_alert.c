#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <ctype.h>
#include <stdarg.h>
#include "http_alert.h"

static char *str_duplicate(const char *s) {
    if (s == NULL) return NULL;
    size_t len = strlen(s) + 1;
    char *dup = (char *)malloc(len);
    if (dup == NULL) return NULL;
    memcpy(dup, s, len);
    return dup;
}

static HttpHeader *create_header(const char *key, const char *value) {
    HttpHeader *hdr = (HttpHeader *)malloc(sizeof(HttpHeader));
    if (hdr == NULL) return NULL;
    hdr->key = str_duplicate(key);
    hdr->value = str_duplicate(value);
    if (hdr->key == NULL || hdr->value == NULL) {
        free(hdr->key);
        free(hdr->value);
        free(hdr);
        return NULL;
    }
    hdr->next = NULL;
    return hdr;
}

void http_alert_free_headers(HttpHeader *headers) {
    HttpHeader *next;
    while (headers != NULL) {
        next = headers->next;
        free(headers->key);
        free(headers->value);
        free(headers);
        headers = next;
    }
}

HttpHeader *http_alert_copy_headers(const HttpHeader *headers) {
    if (headers == NULL) return NULL;
    
    HttpHeader *result = NULL;
    HttpHeader **tail = &result;
    
    while (headers != NULL) {
        HttpHeader *copy = create_header(headers->key, headers->value);
        if (copy == NULL) {
            http_alert_free_headers(result);
            return NULL;
        }
        *tail = copy;
        tail = &copy->next;
        headers = headers->next;
    }
    return result;
}

void http_alert_free_request(HttpRequest *req) {
    if (req == NULL) return;
    free(req->url);
    free(req->host);
    free(req->path);
    free(req->query);
    http_alert_free_headers(req->headers);
    free(req->body);
    free(req);
}

void http_alert_free_response(HttpResponse *resp) {
    if (resp == NULL) return;
    free(resp->status_message);
    http_alert_free_headers(resp->headers);
    free(resp->body);
    free(resp);
}

HttpRequest *http_alert_create_request(void) {
    HttpRequest *req = (HttpRequest *)calloc(1, sizeof(HttpRequest));
    return req;
}

HttpResponse *http_alert_create_response(void) {
    HttpResponse *resp = (HttpResponse *)calloc(1, sizeof(HttpResponse));
    return resp;
}

int http_alert_add_header(HttpRequest *req, const char *key, const char *value) {
    if (req == NULL || key == NULL || value == NULL) {
        return HTTP_ALERT_ERR_INVALID_ARG;
    }
    
    HttpHeader *hdr = create_header(key, value);
    if (hdr == NULL) {
        return HTTP_ALERT_ERR_MEMORY;
    }
    
    HttpHeader **tail = &req->headers;
    while (*tail != NULL) {
        tail = &(*tail)->next;
    }
    *tail = hdr;
    return HTTP_ALERT_OK;
}

int http_alert_set_body(HttpRequest *req, const char *body, size_t len) {
    if (req == NULL) {
        return HTTP_ALERT_ERR_INVALID_ARG;
    }
    
    free(req->body);
    req->body = NULL;
    req->body_len = 0;
    
    if (body == NULL || len == 0) {
        return HTTP_ALERT_OK;
    }
    
    req->body = (char *)malloc(len + 1);
    if (req->body == NULL) {
        return HTTP_ALERT_ERR_MEMORY;
    }
    memcpy(req->body, body, len);
    req->body[len] = '\0';
    req->body_len = len;
    return HTTP_ALERT_OK;
}

int http_alert_parse_url(const char *url, HttpRequest *req) {
    if (url == NULL || req == NULL) {
        return HTTP_ALERT_ERR_INVALID_ARG;
    }
    
    const char *p = url;
    int is_https = 0;
    
    if (strncmp(p, "https://", 8) == 0) {
        is_https = 1;
        p += 8;
    } else if (strncmp(p, "http://", 7) == 0) {
        p += 7;
    } else {
        return HTTP_ALERT_ERR_URL_PARSE;
    }
    
    const char *host_start = p;
    const char *host_end = p;
    unsigned short port = is_https ? 443 : 80;
    
    if (*p == '[') {
        p++;
        host_start = p;
        while (*p != '\0' && *p != ']') p++;
        host_end = p;
        if (*p == ']') p++;
    } else {
        while (*p != '\0' && *p != ':' && *p != '/' && *p != '?') p++;
        host_end = p;
    }
    
    if (*p == ':') {
        p++;
        port = 0;
        while (*p >= '0' && *p <= '9') {
            port = port * 10 + (*p - '0');
            p++;
        }
        if (port == 0) {
            port = is_https ? 443 : 80;
        }
    }
    
    const char *path_start = p;
    const char *path_end = p;
    const char *query_start = NULL;
    
    if (*p == '\0') {
        path_start = "/";
        path_end = path_start + 1;
    } else {
        while (*p != '\0' && *p != '?') p++;
        path_end = p;
        if (*p == '?') {
            query_start = p + 1;
        }
    }
    
    free(req->url);
    free(req->host);
    free(req->path);
    free(req->query);
    
    req->url = str_duplicate(url);
    req->is_https = is_https;
    req->port = port;
    
    req->host = (char *)malloc(host_end - host_start + 1);
    if (req->host == NULL) {
        return HTTP_ALERT_ERR_MEMORY;
    }
    memcpy(req->host, host_start, host_end - host_start);
    req->host[host_end - host_start] = '\0';
    
    size_t path_len = path_end - path_start;
    if (path_len == 0) {
        req->path = str_duplicate("/");
    } else {
        req->path = (char *)malloc(path_len + 1);
        if (req->path == NULL) {
            return HTTP_ALERT_ERR_MEMORY;
        }
        memcpy(req->path, path_start, path_len);
        req->path[path_len] = '\0';
    }
    
    if (query_start != NULL) {
        size_t query_len = strlen(query_start);
        req->query = (char *)malloc(query_len + 1);
        if (req->query == NULL) {
            return HTTP_ALERT_ERR_MEMORY;
        }
        memcpy(req->query, query_start, query_len);
        req->query[query_len] = '\0';
    } else {
        req->query = NULL;
    }
    
    return HTTP_ALERT_OK;
}

const char *http_alert_strerror(int err_code) {
    switch (err_code) {
        case HTTP_ALERT_OK:
            return "Success";
        case HTTP_ALERT_ERR_INVALID_ARG:
            return "Invalid argument";
        case HTTP_ALERT_ERR_MEMORY:
            return "Memory allocation failed";
        case HTTP_ALERT_ERR_SOCKET:
            return "Socket operation failed";
        case HTTP_ALERT_ERR_CONNECT:
            return "Connection failed";
        case HTTP_ALERT_ERR_CONNECT_TIMEOUT:
            return "Connection timeout";
        case HTTP_ALERT_ERR_SEND:
            return "Send failed";
        case HTTP_ALERT_ERR_RECV:
            return "Receive failed";
        case HTTP_ALERT_ERR_RECV_TIMEOUT:
            return "Receive timeout";
        case HTTP_ALERT_ERR_SSL_INIT:
            return "SSL initialization failed";
        case HTTP_ALERT_ERR_SSL_CONNECT:
            return "SSL connection failed";
        case HTTP_ALERT_ERR_URL_PARSE:
            return "URL parse failed";
        case HTTP_ALERT_ERR_OVERALL_TIMEOUT:
            return "Overall timeout exceeded";
        case HTTP_ALERT_ERR_ALL_RETRIES_FAILED:
            return "All retries failed";
        case HTTP_ALERT_ERR_RESOLVE:
            return "Host resolution failed";
        default:
            return "Unknown error";
    }
}

void http_alert_log(const char *level, const char *fmt, ...) {
    va_list args;
    va_start(args, fmt);
    
    time_t now = time(NULL);
    struct tm *tm_now = localtime(&now);
    char time_buf[32];
    strftime(time_buf, sizeof(time_buf), "%Y-%m-%d %H:%M:%S", tm_now);
    
    fprintf(stderr, "[%s] [%s] ", time_buf, level ? level : "INFO");
    vfprintf(stderr, fmt, args);
    fprintf(stderr, "\n");
    va_end(args);
}
