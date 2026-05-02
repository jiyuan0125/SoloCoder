#include "http_server.h"
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <time.h>

void http_response_init(HttpResponse *resp) {
    memset(resp, 0, sizeof(HttpResponse));
    resp->status = HTTP_OK;
    strcpy(resp->status_text, "OK");
    resp->header_count = 0;
    resp->body = NULL;
    resp->body_length = 0;
}

void http_response_free(HttpResponse *resp) {
    if (resp->body) {
        free(resp->body);
        resp->body = NULL;
    }
    resp->body_length = 0;
}

void http_response_set_status(HttpResponse *resp, HttpStatus status) {
    resp->status = status;
    const char *text = http_status_to_text(status);
    strncpy(resp->status_text, text, sizeof(resp->status_text) - 1);
}

void http_response_add_header(HttpResponse *resp, const char *name, const char *value) {
    if (resp->header_count >= MAX_HEADER_COUNT) {
        return;
    }
    
    const char *existing = http_response_get_header(resp, name);
    if (existing) {
        for (int i = 0; i < resp->header_count; i++) {
            if (strcasecmp(resp->headers[i].name, name) == 0) {
                strncpy(resp->headers[i].value, value, sizeof(resp->headers[i].value) - 1);
                return;
            }
        }
    }
    
    strncpy(resp->headers[resp->header_count].name, name, 
            sizeof(resp->headers[resp->header_count].name) - 1);
    strncpy(resp->headers[resp->header_count].value, value, 
            sizeof(resp->headers[resp->header_count].value) - 1);
    resp->header_count++;
}

void http_response_set_body(HttpResponse *resp, const char *body, size_t length) {
    if (resp->body) {
        free(resp->body);
    }
    
    if (body && length > 0) {
        resp->body = (char *)malloc(length + 1);
        if (resp->body) {
            memcpy(resp->body, body, length);
            resp->body[length] = '\0';
            resp->body_length = length;
        } else {
            resp->body_length = 0;
        }
    } else {
        resp->body = NULL;
        resp->body_length = 0;
    }
}

const char *http_response_get_header(const HttpResponse *resp, const char *name) {
    for (int i = 0; i < resp->header_count; i++) {
        if (strcasecmp(resp->headers[i].name, name) == 0) {
            return resp->headers[i].value;
        }
    }
    return NULL;
}

char *http_response_build_status_line(const HttpResponse *resp) {
    char *line = (char *)malloc(256);
    if (!line) return NULL;
    
    snprintf(line, 256, "HTTP/1.1 %d %s\r\n", 
             http_status_to_code(resp->status), resp->status_text);
    return line;
}

char *http_response_serialize(const HttpResponse *resp, size_t *out_length) {
    size_t status_len = 0;
    size_t headers_len = 0;
    size_t total_len = 0;
    
    char temp_status[256];
    snprintf(temp_status, sizeof(temp_status), "HTTP/1.1 %d %s\r\n", 
             http_status_to_code(resp->status), resp->status_text);
    status_len = strlen(temp_status);
    
    for (int i = 0; i < resp->header_count; i++) {
        headers_len += strlen(resp->headers[i].name) + 2 + 
                        strlen(resp->headers[i].value) + 2;
    }
    headers_len += 2;
    
    total_len = status_len + headers_len + resp->body_length;
    
    char *buffer = (char *)malloc(total_len + 1);
    if (!buffer) {
        *out_length = 0;
        return NULL;
    }
    
    char *ptr = buffer;
    
    memcpy(ptr, temp_status, status_len);
    ptr += status_len;
    
    for (int i = 0; i < resp->header_count; i++) {
        size_t name_len = strlen(resp->headers[i].name);
        size_t value_len = strlen(resp->headers[i].value);
        
        memcpy(ptr, resp->headers[i].name, name_len);
        ptr += name_len;
        
        memcpy(ptr, ": ", 2);
        ptr += 2;
        
        memcpy(ptr, resp->headers[i].value, value_len);
        ptr += value_len;
        
        memcpy(ptr, "\r\n", 2);
        ptr += 2;
    }
    
    memcpy(ptr, "\r\n", 2);
    ptr += 2;
    
    if (resp->body && resp->body_length > 0) {
        memcpy(ptr, resp->body, resp->body_length);
        ptr += resp->body_length;
    }
    
    *out_length = ptr - buffer;
    buffer[*out_length] = '\0';
    
    return buffer;
}

void format_rfc1123_date(char *buffer, size_t len) {
    time_t now = time(NULL);
    struct tm *gmt = gmtime(&now);
    
    strftime(buffer, len, "%a, %d %b %Y %H:%M:%S GMT", gmt);
}
