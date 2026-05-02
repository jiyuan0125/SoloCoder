#include "http_server.h"
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <ctype.h>
#include <errno.h>

void http_request_init(HttpRequest *req) {
    memset(req, 0, sizeof(HttpRequest));
    req->method = HTTP_UNKNOWN;
    req->keep_alive = 0;
    req->content_length = 0;
    req->body = NULL;
    req->body_length = 0;
    req->header_count = 0;
}

void http_request_free(HttpRequest *req) {
    if (req->body) {
        free(req->body);
        req->body = NULL;
    }
    req->body_length = 0;
}

const char *http_method_to_string(HttpMethod method) {
    switch (method) {
        case HTTP_GET: return "GET";
        case HTTP_POST: return "POST";
        case HTTP_DELETE: return "DELETE";
        default: return "UNKNOWN";
    }
}

int http_parse_request_line(const char *line, HttpRequest *req) {
    char method[16] = {0};
    char url[8192] = {0};
    char version[16] = {0};
    
    if (sscanf(line, "%15s %8191s %15s", method, url, version) != 3) {
        return -1;
    }
    
    if (strcmp(method, "GET") == 0) {
        req->method = HTTP_GET;
    } else if (strcmp(method, "POST") == 0) {
        req->method = HTTP_POST;
    } else if (strcmp(method, "DELETE") == 0) {
        req->method = HTTP_DELETE;
    } else {
        req->method = HTTP_UNKNOWN;
    }
    strncpy(req->method_str, method, sizeof(req->method_str) - 1);
    strncpy(req->version, version, sizeof(req->version) - 1);
    
    char *query_start = strchr(url, '?');
    if (query_start) {
        *query_start = '\0';
        strncpy(req->query_string, query_start + 1, sizeof(req->query_string) - 1);
    } else {
        req->query_string[0] = '\0';
    }
    
    if (strlen(url) >= sizeof(req->path)) {
        return -1;
    }
    strncpy(req->path, url, sizeof(req->path) - 1);
    
    return 0;
}

int http_parse_header(const char *line, HttpRequest *req) {
    if (req->header_count >= MAX_HEADER_COUNT) {
        return -1;
    }
    
    char *colon = strchr(line, ':');
    if (!colon) {
        return -1;
    }
    
    *colon = '\0';
    const char *name = line;
    const char *value = colon + 1;
    
    while (*value == ' ' || *value == '\t') {
        value++;
    }
    
    size_t name_len = strlen(name);
    size_t value_len = strlen(value);
    
    if (name_len >= sizeof(req->headers[req->header_count].name) ||
        value_len >= sizeof(req->headers[req->header_count].value)) {
        return -1;
    }
    
    strcpy(req->headers[req->header_count].name, name);
    strcpy(req->headers[req->header_count].value, value);
    req->header_count++;
    
    if (strcasecmp(name, "Content-Length") == 0) {
        char *endptr;
        errno = 0;
        long long cl = strtoll(value, &endptr, 10);
        if (errno == 0 && *endptr == '\0' && cl >= 0) {
            req->content_length = (size_t)cl;
        }
    } else if (strcasecmp(name, "Connection") == 0) {
        if (strstr(value, "keep-alive") != NULL) {
            req->keep_alive = 1;
        } else {
            req->keep_alive = 0;
        }
    }
    
    return 0;
}

int http_parse_headers(const char *data, size_t len, HttpRequest *req, size_t *consumed) {
    *consumed = 0;
    size_t pos = 0;
    int request_line_parsed = 0;
    size_t headers_total = 0;
    
    while (pos < len) {
        size_t line_start = pos;
        int line_complete = 0;
        const char *line_end = NULL;
        
        while (pos + 1 < len) {
            if (data[pos] == '\r' && data[pos + 1] == '\n') {
                line_end = &data[pos];
                pos += 2;
                line_complete = 1;
                break;
            } else if (data[pos] == '\n') {
                line_end = &data[pos];
                pos += 1;
                line_complete = 1;
                break;
            }
            pos++;
        }
        
        if (!line_complete) {
            break;
        }
        
        size_t line_len = line_end - (data + line_start);
        char line[8192] = {0};
        if (line_len >= sizeof(line)) {
            return -1;
        }
        memcpy(line, data + line_start, line_len);
        line[line_len] = '\0';
        
        headers_total += line_len + 2;
        if (headers_total > MAX_HEADERS_SIZE) {
            return -2;
        }
        
        if (line_len == 0) {
            *consumed = pos;
            return 0;
        }
        
        if (!request_line_parsed) {
            if (http_parse_request_line(line, req) != 0) {
                return -1;
            }
            request_line_parsed = 1;
        } else {
            if (http_parse_header(line, req) != 0) {
                return -1;
            }
        }
    }
    
    return 1;
}

const char *http_status_to_text(HttpStatus status) {
    switch (status) {
        case HTTP_OK: return "OK";
        case HTTP_BAD_REQUEST: return "Bad Request";
        case HTTP_NOT_FOUND: return "Not Found";
        case HTTP_FORBIDDEN: return "Forbidden";
        case HTTP_REQUEST_URI_TOO_LONG: return "Request-URI Too Long";
        case HTTP_REQUEST_ENTITY_TOO_LARGE: return "Request Entity Too Large";
        case HTTP_REQUEST_HEADER_FIELDS_TOO_LARGE: return "Request Header Fields Too Large";
        case HTTP_METHOD_NOT_ALLOWED: return "Method Not Allowed";
        case HTTP_INTERNAL_SERVER_ERROR: return "Internal Server Error";
        case HTTP_NOT_IMPLEMENTED: return "Not Implemented";
        case HTTP_CONTINUE: return "Continue";
        default: return "Internal Server Error";
    }
}

int http_status_to_code(HttpStatus status) {
    return (int)status;
}
