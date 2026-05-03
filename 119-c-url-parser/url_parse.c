#include "url_parse.h"
#include <stdlib.h>
#include <string.h>
#include <stdio.h>
#include <ctype.h>

static URL *url_create(void) {
    URL *url = (URL *)malloc(sizeof(URL));
    if (!url) return NULL;
    memset(url, 0, sizeof(URL));
    url->port = -1;
    url->valid = 1;
    url->query_params = query_params_create();
    if (!url->query_params) {
        free(url);
        return NULL;
    }
    return url;
}

void url_destroy(URL *url) {
    if (!url) return;
    free(url->protocol_str);
    free(url->username);
    free(url->password);
    free(url->hostname);
    free(url->port_str);
    free(url->path);
    free(url->normalized_path);
    free(url->fragment);
    query_params_destroy(url->query_params);
    free(url);
}

int url_get_default_port(URLProtocol protocol) {
    switch (protocol) {
        case URL_PROTOCOL_HTTP: return 80;
        case URL_PROTOCOL_HTTPS: return 443;
        default: return -1;
    }
}

const char *url_protocol_to_string(URLProtocol protocol) {
    switch (protocol) {
        case URL_PROTOCOL_HTTP: return "http";
        case URL_PROTOCOL_HTTPS: return "https";
        default: return NULL;
    }
}

URLProtocol url_string_to_protocol(const char *str) {
    if (!str) return URL_PROTOCOL_UNKNOWN;
    if (strcasecmp(str, "http") == 0) return URL_PROTOCOL_HTTP;
    if (strcasecmp(str, "https") == 0) return URL_PROTOCOL_HTTPS;
    return URL_PROTOCOL_UNKNOWN;
}

static char *strndup_safe(const char *s, size_t n) {
    if (!s) return NULL;
    char *result = (char *)malloc(n + 1);
    if (!result) return NULL;
    strncpy(result, s, n);
    result[n] = '\0';
    return result;
}

static int is_valid_port(const char *str) {
    if (!str || !*str) return 0;
    for (const char *p = str; *p; p++) {
        if (!isdigit((unsigned char)*p)) return 0;
    }
    long port = atol(str);
    return (port >= 0 && port <= 65535);
}

char *normalize_path(const char *path, int *path_traversal) {
    if (!path || !*path) {
        *path_traversal = 0;
        return strdup("/");
    }
    
    int is_absolute = (path[0] == '/');
    char *path_copy = strdup(path);
    if (!path_copy) return NULL;
    
    char **segments = NULL;
    int segment_count = 0;
    int segment_cap = 0;
    
    char *p = strtok(path_copy, "/");
    while (p) {
        if (segment_count >= segment_cap) {
            int new_cap = segment_cap ? segment_cap * 2 : 8;
            char **new_segments = (char **)realloc(segments, sizeof(char *) * new_cap);
            if (!new_segments) {
                free(path_copy);
                free(segments);
                return NULL;
            }
            segments = new_segments;
            segment_cap = new_cap;
        }
        segments[segment_count++] = p;
        p = strtok(NULL, "/");
    }
    
    const char **result_segments = NULL;
    int result_count = 0;
    int result_cap = 0;
    int depth = 0;
    int traversal_occurred = 0;
    
    for (int i = 0; i < segment_count; i++) {
        const char *seg = segments[i];
        if (strcmp(seg, ".") == 0) {
            continue;
        } else if (strcmp(seg, "..") == 0) {
            if (result_count > 0) {
                result_count--;
                depth--;
            } else if (is_absolute) {
                depth--;
                if (depth < 0) {
                    traversal_occurred = 1;
                }
            } else {
                if (result_count >= result_cap) {
                    int new_cap = result_cap ? result_cap * 2 : 8;
                    const char **new_res = (const char **)realloc(result_segments, sizeof(const char *) * new_cap);
                    if (!new_res) {
                        free(path_copy);
                        free(segments);
                        free(result_segments);
                        return NULL;
                    }
                    result_segments = new_res;
                    result_cap = new_cap;
                }
                result_segments[result_count++] = "..";
                depth--;
            }
        } else if (seg[0] != '\0') {
            if (result_count >= result_cap) {
                int new_cap = result_cap ? result_cap * 2 : 8;
                const char **new_res = (const char **)realloc(result_segments, sizeof(const char *) * new_cap);
                if (!new_res) {
                    free(path_copy);
                    free(segments);
                    free(result_segments);
                    return NULL;
                }
                result_segments = new_res;
                result_cap = new_cap;
            }
            result_segments[result_count++] = seg;
            depth++;
        }
    }
    
    *path_traversal = traversal_occurred;
    
    size_t result_len = 1;
    if (is_absolute) result_len++;
    
    for (int i = 0; i < result_count; i++) {
        if (i > 0) result_len++;
        result_len += strlen(result_segments[i]);
    }
    
    char *result = (char *)malloc(result_len);
    if (!result) {
        free(path_copy);
        free(segments);
        free(result_segments);
        return NULL;
    }
    
    char *dst = result;
    if (is_absolute) {
        *dst++ = '/';
    }
    
    for (int i = 0; i < result_count; i++) {
        if (i > 0 && dst > result && *(dst - 1) != '/') {
            *dst++ = '/';
        }
        const char *seg = result_segments[i];
        while (*seg) {
            *dst++ = *seg++;
        }
    }
    
    if (dst == result) {
        *dst++ = '/';
    }
    
    *dst = '\0';
    
    free(path_copy);
    free(segments);
    free(result_segments);
    
    return result;
}

URL *url_parse(const char *url_str) {
    if (!url_str || !*url_str) return NULL;
    
    URL *url = url_create();
    if (!url) return NULL;
    
    const char *p = url_str;
    const char *end = p + strlen(p);
    
    const char *proto_end = strstr(p, "://");
    if (proto_end) {
        url->protocol_str = strndup_safe(p, proto_end - p);
        url->protocol = url_string_to_protocol(url->protocol_str);
        if (url->protocol == URL_PROTOCOL_UNKNOWN) {
            url->valid = 0;
        }
        p = proto_end + 3;
    } else {
        url->valid = 0;
    }
    
    const char *slash_pos = strchr(p, '/');
    const char *question_pos = strchr(p, '?');
    const char *hash_pos = strchr(p, '#');
    
    const char *authority_end = slash_pos;
    if (!authority_end || (question_pos && question_pos < authority_end)) {
        authority_end = question_pos;
    }
    if (!authority_end || (hash_pos && hash_pos < authority_end)) {
        authority_end = hash_pos;
    }
    if (!authority_end) {
        authority_end = end;
    }
    
    const char *authority_start = p;
    size_t authority_len = authority_end - authority_start;
    
    const char *at_pos = NULL;
    if (authority_len > 0) {
        const char *tmp = authority_start;
        while (tmp < authority_end) {
            if (*tmp == '@') {
                at_pos = tmp;
            }
            tmp++;
        }
    }
    
    if (at_pos) {
        const char *userinfo_start = authority_start;
        size_t userinfo_len = at_pos - userinfo_start;
        
        const char *colon_pos = NULL;
        const char *tmp = userinfo_start;
        while (tmp < at_pos) {
            if (*tmp == ':') {
                colon_pos = tmp;
                break;
            }
            tmp++;
        }
        
        if (colon_pos) {
            url->username = strndup_safe(userinfo_start, colon_pos - userinfo_start);
            url->password = strndup_safe(colon_pos + 1, at_pos - (colon_pos + 1));
        } else {
            url->username = strndup_safe(userinfo_start, userinfo_len);
        }
        
        p = at_pos + 1;
    }
    
    const char *host_start = p;
    const char *host_end = authority_end;
    const char *port_start = NULL;
    
    if (*host_start == '[') {
        const char *bracket_end = strchr(host_start, ']');
        if (bracket_end && bracket_end < host_end) {
            url->hostname = strndup_safe(host_start + 1, bracket_end - host_start - 1);
            if (*(bracket_end + 1) == ':') {
                port_start = bracket_end + 2;
            }
        } else {
            url->valid = 0;
        }
    } else {
        const char *tmp = host_start;
        while (tmp < host_end) {
            if (*tmp == ':') {
                port_start = tmp + 1;
                break;
            }
            tmp++;
        }
        
        if (port_start) {
            url->hostname = strndup_safe(host_start, port_start - host_start - 1);
        } else {
            url->hostname = strndup_safe(host_start, host_end - host_start);
        }
    }
    
    if (port_start && port_start < host_end) {
        url->port_str = strndup_safe(port_start, host_end - port_start);
        if (!is_valid_port(url->port_str)) {
            url->valid = 0;
        } else {
            url->port = atoi(url->port_str);
        }
    }
    
    if (url->port < 0 && url->protocol != URL_PROTOCOL_UNKNOWN) {
        url->port = url_get_default_port(url->protocol);
    }
    
    p = authority_end;
    
    const char *path_start = p;
    const char *path_end = question_pos ? question_pos : (hash_pos ? hash_pos : end);
    
    if (p < path_end) {
        url->path = strndup_safe(path_start, path_end - path_start);
        if (!url->path) {
            url_destroy(url);
            return NULL;
        }
        
        int pt = 0;
        url->normalized_path = normalize_path(url->path, &pt);
        url->path_traversal = pt;
        if (pt) {
            url->valid = 0;
        }
    } else {
        url->path = strdup("/");
        int pt = 0;
        url->normalized_path = normalize_path("/", &pt);
    }
    
    if (question_pos && (!hash_pos || question_pos < hash_pos)) {
        const char *query_start = question_pos + 1;
        const char *query_end = hash_pos ? hash_pos : end;
        
        if (query_start < query_end) {
            char *query_str = strndup_safe(query_start, query_end - query_start);
            if (query_str) {
                query_params_parse(url->query_params, query_str);
                free(query_str);
            }
        }
    }
    
    if (hash_pos) {
        url->fragment = strdup(hash_pos + 1);
    }
    
    return url;
}
