#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <ctype.h>
#include <time.h>
#include "http_request.h"

#define RESPONSE_BUF_SIZE 8192

const char *http_method_to_string(HttpMethod method) {
    switch (method) {
        case HTTP_METHOD_GET:
            return "GET";
        case HTTP_METHOD_POST:
            return "POST";
        case HTTP_METHOD_PUT:
            return "PUT";
        default:
            return "GET";
    }
}

static size_t http_request_calculate_size(const HttpRequest *req) {
    size_t size = 0;
    const char *method_str = http_method_to_string(req->method);
    
    size += strlen(method_str) + 1;
    size += strlen(req->path);
    if (req->query) {
        size += 1 + strlen(req->query);
    }
    size += 11;
    
    size += 6 + strlen(req->host) + 2;
    if (req->port != 80 && req->port != 443) {
        char port_str[8];
        snprintf(port_str, sizeof(port_str), ":%d", req->port);
        size += strlen(port_str);
    }
    size += 2;
    
    size += 19;
    
    HttpHeader *hdr = req->headers;
    while (hdr != NULL) {
        size += strlen(hdr->key) + 2 + strlen(hdr->value) + 2;
        hdr = hdr->next;
    }
    
    if (req->body && req->body_len > 0) {
        char cl_str[64];
        snprintf(cl_str, sizeof(cl_str), "Content-Length: %zu\r\n", req->body_len);
        size += strlen(cl_str);
    }
    
    size += 2;
    if (req->body && req->body_len > 0) {
        size += req->body_len;
    }
    
    return size;
}

static char *http_request_build(const HttpRequest *req) {
    size_t size = http_request_calculate_size(req);
    char *buffer = (char *)malloc(size + 1);
    if (buffer == NULL) {
        return NULL;
    }
    
    char *ptr = buffer;
    const char *method_str = http_method_to_string(req->method);
    
    ptr += sprintf(ptr, "%s %s", method_str, req->path);
    if (req->query) {
        ptr += sprintf(ptr, "?%s", req->query);
    }
    ptr += sprintf(ptr, " HTTP/1.1\r\n");
    
    ptr += sprintf(ptr, "Host: %s", req->host);
    if (req->port != 80 && req->port != 443) {
        ptr += sprintf(ptr, ":%d", req->port);
    }
    ptr += sprintf(ptr, "\r\n");
    
    ptr += sprintf(ptr, "Connection: close\r\n");
    
    HttpHeader *hdr = req->headers;
    while (hdr != NULL) {
        ptr += sprintf(ptr, "%s: %s\r\n", hdr->key, hdr->value);
        hdr = hdr->next;
    }
    
    if (req->body && req->body_len > 0) {
        ptr += sprintf(ptr, "Content-Length: %zu\r\n", req->body_len);
    }
    
    ptr += sprintf(ptr, "\r\n");
    
    if (req->body && req->body_len > 0) {
        memcpy(ptr, req->body, req->body_len);
        ptr += req->body_len;
    }
    
    *ptr = '\0';
    return buffer;
}

static int parse_status_line(const char *line, HttpResponse *resp) {
    const char *p = line;
    
    while (*p != ' ' && *p != '\0') p++;
    if (*p == ' ') p++;
    
    resp->status_code = 0;
    while (*p >= '0' && *p <= '9') {
        resp->status_code = resp->status_code * 10 + (*p - '0');
        p++;
    }
    
    while (*p == ' ') p++;
    
    const char *msg_start = p;
    const char *msg_end = p;
    while (*msg_end != '\0' && *msg_end != '\r' && *msg_end != '\n') {
        msg_end++;
    }
    
    size_t msg_len = msg_end - msg_start;
    resp->status_message = (char *)malloc(msg_len + 1);
    if (resp->status_message == NULL) {
        return HTTP_ALERT_ERR_MEMORY;
    }
    memcpy(resp->status_message, msg_start, msg_len);
    resp->status_message[msg_len] = '\0';
    
    return HTTP_ALERT_OK;
}

static int parse_header_line(const char *line, HttpHeader **headers) {
    const char *p = line;
    const char *key_end = p;
    
    while (*key_end != ':' && *key_end != '\0') {
        key_end++;
    }
    
    if (*key_end != ':') {
        return HTTP_ALERT_OK;
    }
    
    size_t key_len = key_end - p;
    while (key_len > 0 && isspace((unsigned char)p[key_len - 1])) {
        key_len--;
    }
    
    p = key_end + 1;
    while (*p == ' ' || *p == '\t') {
        p++;
    }
    
    const char *value_end = p;
    while (*value_end != '\0' && *value_end != '\r' && *value_end != '\n') {
        value_end++;
    }
    size_t value_len = value_end - p;
    
    while (value_len > 0 && isspace((unsigned char)p[value_len - 1])) {
        value_len--;
    }
    
    char *key = (char *)malloc(key_len + 1);
    char *value = (char *)malloc(value_len + 1);
    if (key == NULL || value == NULL) {
        free(key);
        free(value);
        return HTTP_ALERT_ERR_MEMORY;
    }
    
    memcpy(key, line, key_len);
    key[key_len] = '\0';
    memcpy(value, p, value_len);
    value[value_len] = '\0';
    
    HttpHeader *hdr = (HttpHeader *)malloc(sizeof(HttpHeader));
    if (hdr == NULL) {
        free(key);
        free(value);
        return HTTP_ALERT_ERR_MEMORY;
    }
    hdr->key = key;
    hdr->value = value;
    hdr->next = NULL;
    
    HttpHeader **tail = headers;
    while (*tail != NULL) {
        tail = &(*tail)->next;
    }
    *tail = hdr;
    
    return HTTP_ALERT_OK;
}

static int parse_response_headers(const char *data, size_t len, HttpResponse *resp) {
    const char *p = data;
    const char *end = data + len;
    int line_count = 0;
    
    while (p < end) {
        const char *line_start = p;
        const char *line_end = p;
        
        while (line_end < end && *line_end != '\r' && *line_end != '\n') {
            line_end++;
        }
        
        if (line_start == line_end) {
            return HTTP_ALERT_OK;
        }
        
        if (line_count == 0) {
            int ret = parse_status_line(line_start, resp);
            if (ret != HTTP_ALERT_OK) {
                return ret;
            }
        } else {
            int ret = parse_header_line(line_start, &resp->headers);
            if (ret != HTTP_ALERT_OK) {
                return ret;
            }
        }
        
        line_count++;
        p = line_end;
        if (p < end && *p == '\r') p++;
        if (p < end && *p == '\n') p++;
    }
    
    return HTTP_ALERT_OK;
}

static const char *find_content_length(HttpHeader *headers) {
    HttpHeader *hdr = headers;
    while (hdr != NULL) {
        if (strcasecmp(hdr->key, "Content-Length") == 0) {
            return hdr->value;
        }
        hdr = hdr->next;
    }
    return NULL;
}

int http_request_send(HttpConnection *conn, const HttpRequest *req) {
    if (conn == NULL || req == NULL) {
        return HTTP_ALERT_ERR_INVALID_ARG;
    }
    
    char *request = http_request_build(req);
    if (request == NULL) {
        return HTTP_ALERT_ERR_MEMORY;
    }
    
    size_t request_len = strlen(request);
    int ret = http_connection_send(conn, request, request_len);
    
    free(request);
    
    if (ret < 0) {
        return ret;
    }
    
    return HTTP_ALERT_OK;
}

int http_request_receive(HttpConnection *conn, HttpResponse *resp) {
    if (conn == NULL || resp == NULL) {
        return HTTP_ALERT_ERR_INVALID_ARG;
    }
    
    char *header_buf = NULL;
    size_t header_cap = 0;
    size_t header_len = 0;
    char recv_buf[RESPONSE_BUF_SIZE];
    int ret;
    int headers_complete = 0;
    size_t body_start_offset = 0;
    
    while (!headers_complete) {
        ret = http_connection_recv(conn, recv_buf, sizeof(recv_buf) - 1);
        if (ret < 0) {
            free(header_buf);
            return ret;
        }
        if (ret == 0) {
            break;
        }
        
        if (header_len + ret > header_cap) {
            size_t new_cap = header_cap == 0 ? RESPONSE_BUF_SIZE : header_cap * 2;
            while (new_cap < header_len + ret) {
                new_cap *= 2;
            }
            char *new_buf = (char *)realloc(header_buf, new_cap);
            if (new_buf == NULL) {
                free(header_buf);
                return HTTP_ALERT_ERR_MEMORY;
            }
            header_buf = new_buf;
            header_cap = new_cap;
        }
        
        memcpy(header_buf + header_len, recv_buf, ret);
        header_len += ret;
        
        size_t i;
        for (i = 3; i < header_len; i++) {
            if (header_buf[i - 3] == '\r' && header_buf[i - 2] == '\n' &&
                header_buf[i - 1] == '\r' && header_buf[i] == '\n') {
                headers_complete = 1;
                body_start_offset = i + 1;
                break;
            }
        }
    }
    
    if (header_len == 0) {
        free(header_buf);
        return HTTP_ALERT_ERR_RECV;
    }
    
    ret = parse_response_headers(header_buf, body_start_offset > 0 ? body_start_offset - 1 : header_len, resp);
    if (ret != HTTP_ALERT_OK) {
        free(header_buf);
        return ret;
    }
    
    size_t body_size = 0;
    size_t body_cap = 0;
    char *body_buf = NULL;
    
    if (body_start_offset > 0 && body_start_offset < header_len) {
        size_t initial_body = header_len - body_start_offset;
        if (initial_body > 0) {
            body_buf = (char *)malloc(initial_body);
            if (body_buf == NULL) {
                free(header_buf);
                return HTTP_ALERT_ERR_MEMORY;
            }
            memcpy(body_buf, header_buf + body_start_offset, initial_body);
            body_size = initial_body;
            body_cap = initial_body;
        }
    }
    
    free(header_buf);
    
    const char *cl_str = find_content_length(resp->headers);
    size_t content_length = 0;
    int has_content_length = 0;
    
    if (cl_str != NULL) {
        content_length = atoi(cl_str);
        has_content_length = 1;
    }
    
    while (http_connection_is_connected(conn)) {
        if (has_content_length && body_size >= content_length) {
            break;
        }
        
        ret = http_connection_recv(conn, recv_buf, sizeof(recv_buf));
        if (ret < 0) {
            if (ret == HTTP_ALERT_ERR_RECV_TIMEOUT) {
                break;
            }
            free(body_buf);
            return ret;
        }
        if (ret == 0) {
            break;
        }
        
        if (body_size + ret > body_cap) {
            size_t new_cap = body_cap == 0 ? RESPONSE_BUF_SIZE : body_cap * 2;
            while (new_cap < body_size + ret) {
                new_cap *= 2;
            }
            char *new_buf = (char *)realloc(body_buf, new_cap);
            if (new_buf == NULL) {
                free(body_buf);
                return HTTP_ALERT_ERR_MEMORY;
            }
            body_buf = new_buf;
            body_cap = new_cap;
        }
        
        memcpy(body_buf + body_size, recv_buf, ret);
        body_size += ret;
    }
    
    resp->body = body_buf;
    resp->body_len = body_size;
    
    return HTTP_ALERT_OK;
}

int http_request_execute(const HttpRequest *req, HttpResponse *resp, 
                          const TimeoutConfig *timeout) {
    if (req == NULL || resp == NULL || timeout == NULL) {
        return HTTP_ALERT_ERR_INVALID_ARG;
    }
    
    HttpConnection conn;
    int ret = http_connection_init(&conn, req->is_https);
    if (ret != HTTP_ALERT_OK) {
        return ret;
    }
    
    ret = http_connection_connect(&conn, req->host, req->port, timeout);
    if (ret != HTTP_ALERT_OK) {
        http_connection_cleanup(&conn);
        return ret;
    }
    
    ret = http_request_send(&conn, req);
    if (ret != HTTP_ALERT_OK) {
        http_connection_cleanup(&conn);
        return ret;
    }
    
    ret = http_request_receive(&conn, resp);
    http_connection_cleanup(&conn);
    
    return ret;
}

static void json_escape_string(const char *input, char *output, size_t output_size) {
    if (input == NULL || output == NULL || output_size == 0) {
        return;
    }
    
    const char *p = input;
    char *op = output;
    size_t remaining = output_size - 1;
    
    while (*p != '\0' && remaining > 0) {
        switch (*p) {
            case '\"':
                if (remaining < 2) goto done;
                *op++ = '\\';
                *op++ = '\"';
                remaining -= 2;
                break;
            case '\\':
                if (remaining < 2) goto done;
                *op++ = '\\';
                *op++ = '\\';
                remaining -= 2;
                break;
            case '\n':
                if (remaining < 2) goto done;
                *op++ = '\\';
                *op++ = 'n';
                remaining -= 2;
                break;
            case '\r':
                if (remaining < 2) goto done;
                *op++ = '\\';
                *op++ = 'r';
                remaining -= 2;
                break;
            case '\t':
                if (remaining < 2) goto done;
                *op++ = '\\';
                *op++ = 't';
                remaining -= 2;
                break;
            default:
                if ((unsigned char)*p < 32) {
                    if (remaining < 6) goto done;
                    snprintf(op, remaining + 1, "\\u%04x", (unsigned char)*p);
                    op += 6;
                    remaining -= 6;
                } else {
                    *op++ = *p;
                    remaining--;
                }
                break;
        }
        p++;
    }
done:
    *op = '\0';
}

char *alert_data_to_json(const AlertData *alert) {
    if (alert == NULL) {
        return NULL;
    }
    
    char escaped_title[512];
    char escaped_message[4096];
    char escaped_level[64];
    char escaped_source[256];
    char time_buf[64];
    
    json_escape_string(alert->title ? alert->title : "", escaped_title, sizeof(escaped_title));
    json_escape_string(alert->message ? alert->message : "", escaped_message, sizeof(escaped_message));
    json_escape_string(alert->level ? alert->level : "info", escaped_level, sizeof(escaped_level));
    json_escape_string(alert->source ? alert->source : "", escaped_source, sizeof(escaped_source));
    
    time_t ts = alert->timestamp > 0 ? alert->timestamp : time(NULL);
    struct tm *tm_ts = gmtime(&ts);
    strftime(time_buf, sizeof(time_buf), "%Y-%m-%dT%H:%M:%SZ", tm_ts);
    
    size_t json_size = 512 + strlen(escaped_title) + strlen(escaped_message) + 
                       strlen(escaped_level) + strlen(escaped_source);
    char *json = (char *)malloc(json_size);
    if (json == NULL) {
        return NULL;
    }
    
    snprintf(json, json_size, 
             "{"
             "\"title\":\"%s\","
             "\"message\":\"%s\","
             "\"level\":\"%s\","
             "\"source\":\"%s\","
             "\"timestamp\":\"%s\""
             "}",
             escaped_title, escaped_message, escaped_level, escaped_source, time_buf);
    
    return json;
}
